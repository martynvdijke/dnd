import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/features/onboarding/onboarding_screen.dart';

import '../helpers/fake_api.dart';
import '../helpers/pump.dart';

void main() {
  testWidgets('URL validation rejects non-http', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    final urlField = find.byType(TextField).first;
    await tester.enterText(urlField, 'ftp://example.com');
    await tester.tap(find.widgetWithText(OutlinedButton, 'Test connection'));
    await tester.pump();
    expect(find.textContaining('valid URL'), findsOneWidget);
    expect(fake.requests, isEmpty);
  });

  testWidgets('login failure renders error', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onPost(
      '/api/login',
      (o) async => jsonBody({'error': 'invalid credentials'}, 401),
    );
    fake.onGet('/api/check-setup', (o) async => jsonBody({'setup': true}, 200));
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    // need to fill username/password and tap login
    await tester.enterText(find.byType(TextField).at(1), 'bob');
    await tester.enterText(find.byType(TextField).at(2), 'wrong');
    await tester.tap(find.widgetWithText(FilledButton, 'Log in'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));
    // error container should appear
    expect(find.byIcon(Icons.error_outline), findsOneWidget);
  });

  testWidgets(
    'when checkSetup reports setup:false switches to first-admin mode and enforces 8 chars',
    (tester) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onGet(
        '/api/check-setup',
        (o) async => jsonBody({'setup': false}, 200),
      );
      await pumpScreen(
        tester,
        const OnboardingScreen(),
        fake: fake,
        store: store,
      );
      await tester.tap(find.widgetWithText(OutlinedButton, 'Test connection'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));
      expect(find.textContaining('first admin'), findsWidgets);
      // try login with short password
      await tester.enterText(find.byType(TextField).at(1), 'admin');
      await tester.enterText(find.byType(TextField).at(2), 'short');
      final fb = find.byType(FilledButton);
      await tester.ensureVisible(fb);
      await tester.pumpAndSettle();
      await tester.tap(fb, warnIfMissed: false);
      await tester.pump();
      expect(find.textContaining('at least 8 characters'), findsOneWidget);
    },
  );

  testWidgets('successful login transitions auth state', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onPost(
      '/api/login',
      (o) async => jsonBody({
        'session': 's1',
        'user': {'id': 1, 'username': 'admin', 'role': 'admin'},
      }, 200),
    );
    fake.onPost('/api/tokens', (o) async => jsonBody({'token': 'vlt_x'}, 201));
    fake.onGet('/api/check-setup', (o) async => jsonBody({'setup': true}, 200));
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    await tester.enterText(find.byType(TextField).at(1), 'admin');
    await tester.enterText(find.byType(TextField).at(2), 'longpassword');
    await tester.tap(find.byType(FilledButton));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));
    // After login, error should be null and store should have session
    final stored = await store.read();
    expect(stored, isNotNull);
    expect(stored!.sessionId, 's1');
  });

  testWidgets('forceSetup true shows create admin UI', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await pumpScreen(
      tester,
      const OnboardingScreen(forceSetup: true),
      fake: fake,
      store: store,
    );
    expect(find.textContaining('Create admin'), findsWidgets);
  });
}
