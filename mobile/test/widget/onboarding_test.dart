import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/core/config.dart';
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
    await tester.tap(find.widgetWithText(FilledButton, 'Log in'));
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
      await tester.enterText(
        find.byType(TextField).first,
        'http://villum.test',
      );
      await tester.pump(const Duration(milliseconds: 700));
      await tester.pump();
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
    await tester.enterText(find.byType(TextField).first, 'http://villum.test');
    await tester.enterText(find.byType(TextField).at(1), 'admin');
    await tester.enterText(find.byType(TextField).at(2), 'longpassword');
    await tester.tap(find.byType(FilledButton));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));
    // After login, error should be null and store should have session
    final stored = await store.read();
    expect(stored, isNotNull);
    expect(stored!.sessionId, 's1');
    expect(stored.serverUrl, 'http://villum.test');
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

  testWidgets('prefills the hosted server address when none is saved', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    final urlField = tester.widget<TextField>(find.byType(TextField).first);
    expect(urlField.controller?.text, defaultServerUrl);
    expect(find.textContaining('self-hosted DnD server'), findsOneWidget);
  });

  testWidgets('saved server address wins over the hosted default', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await store.write(
      const StoredAuth(
        serverUrl: 'https://selfhosted.example',
        sessionId: 'saved-session',
        userId: 1,
        username: 'admin',
        role: 'admin',
      ),
    );
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    await tester.pump(const Duration(milliseconds: 700));
    await tester.pump();
    final urlField = tester.widget<TextField>(find.byType(TextField).first);
    expect(urlField.controller?.text, 'https://selfhosted.example');
  });

  testWidgets('auto-detects Authelia sign-in from the server URL', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onGet('/api/check-setup', (o) async => jsonBody({'setup': true}, 200));
    fake.onGet(
      '/api/auth/oidc/status',
      (o) async => jsonBody({'enabled': true}, 200),
    );
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );

    // Entering a valid address is enough to reveal the SSO action.
    await tester.enterText(find.byType(TextField).first, 'http://villum.test');
    await tester.pump(const Duration(milliseconds: 700));
    await tester.pump();

    expect(
      fake.requests.where((r) => r.path == '/api/auth/oidc/status'),
      hasLength(1),
    );
    expect(find.byKey(const Key('sso-sign-in')), findsOneWidget);
    expect(find.text('Password sign-in (break-glass)'), findsOneWidget);
  });

  testWidgets('offers Authelia sign-in when the server enables OIDC', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onGet('/api/check-setup', (o) async => jsonBody({'setup': true}, 200));
    fake.onGet(
      '/api/auth/oidc/status',
      (o) async => jsonBody({'enabled': true}, 200),
    );
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    await tester.enterText(find.byType(TextField).first, 'http://villum.test');
    await tester.pump(const Duration(milliseconds: 700));
    await tester.pump();

    expect(find.byKey(const Key('sso-sign-in')), findsOneWidget);
    expect(find.text('Sign in with Authelia'), findsWidgets);
    // The password path stays available but is demoted to break-glass.
    expect(find.text('Password sign-in (break-glass)'), findsOneWidget);
  });

  testWidgets('hides Authelia sign-in when the server disables OIDC', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onGet('/api/check-setup', (o) async => jsonBody({'setup': true}, 200));
    fake.onGet(
      '/api/auth/oidc/status',
      (o) async => jsonBody({'enabled': false}, 200),
    );
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    await tester.enterText(find.byType(TextField).first, 'http://villum.test');
    await tester.pump(const Duration(milliseconds: 700));
    await tester.pump();

    expect(find.byKey(const Key('sso-sign-in')), findsNothing);
    expect(find.widgetWithText(FilledButton, 'Log in'), findsOneWidget);
  });

  testWidgets('SSO sign-in persists the captured session', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    String? launchedWith;
    fake.onGet('/api/check-setup', (o) async => jsonBody({'setup': true}, 200));
    fake.onGet(
      '/api/auth/oidc/status',
      (o) async => jsonBody({'enabled': true}, 200),
    );
    fake.onGet(
      '/api/user/me',
      (o) async =>
          jsonBody({'id': 4, 'username': 'martyn', 'role': 'dm'}, 200),
    );
    fake.onPost('/api/tokens', (o) async => jsonBody({'token': 'vlt_sso'}, 201));

    await pumpScreen(
      tester,
      OnboardingScreen(
        ssoLauncher: (context, serverUrl) async {
          launchedWith = serverUrl;
          return 'oidc-session';
        },
      ),
      fake: fake,
      store: store,
    );
    await tester.enterText(find.byType(TextField).first, 'http://villum.test/');
    await tester.pump(const Duration(milliseconds: 700));
    await tester.pump();

    final sso = find.byKey(const Key('sso-sign-in'));
    await tester.ensureVisible(sso);
    await tester.pumpAndSettle();
    await tester.tap(sso);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(launchedWith, 'http://villum.test');
    final stored = await store.read();
    expect(stored, isNotNull);
    expect(stored!.sessionId, 'oidc-session');
    expect(stored.serverUrl, 'http://villum.test');
  });

  testWidgets('password 403 explains that Authelia sign-in is required', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onPost(
      '/api/login',
      (o) async => jsonBody({'error': 'password login disabled'}, 403),
    );
    fake.onGet('/api/check-setup', (o) async => jsonBody({'setup': true}, 200));
    fake.onGet(
      '/api/auth/oidc/status',
      (o) async => jsonBody({'enabled': true}, 200),
    );
    await pumpScreen(
      tester,
      const OnboardingScreen(),
      fake: fake,
      store: store,
    );
    await tester.enterText(find.byType(TextField).first, 'http://villum.test');
    await tester.enterText(find.byType(TextField).at(1), 'admin');
    await tester.enterText(find.byType(TextField).at(2), 'longpassword');
    await tester.tap(find.byType(FilledButton));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(find.textContaining('Sign in with Authelia'), findsWidgets);
    expect(find.textContaining('Wrong username or password'), findsNothing);
  });
}
