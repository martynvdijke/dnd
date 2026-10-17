import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/features/compendium/compendium_screen.dart';

import '../helpers/fake_api.dart';
import '../helpers/pump.dart';

void main() {
  testWidgets('typing issues debounced search and renders results', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    int callCount = 0;
    fake.onGet('/api/compendium/search', (o) async {
      callCount++;
      final q = o.queryParameters['q']?.toString() ?? '';
      if (q.contains('fire')) {
        return jsonBody([
          {
            'type': 'spell',
            'id': '1',
            'name': 'Fireball',
            'subtype': 'Evocation',
            'level': 3,
          },
        ], 200);
      }
      return jsonBody([], 200);
    });
    await pumpScreen(
      tester,
      const CompendiumScreen(),
      fake: fake,
      store: store,
    );
    await tester.enterText(find.byType(TextField), 'f');
    await tester.pump(const Duration(milliseconds: 100));
    await tester.enterText(find.byType(TextField), 'fi');
    await tester.pump(const Duration(milliseconds: 100));
    await tester.enterText(find.byType(TextField), 'fire');
    // before debounce, no request
    expect(callCount, 0);
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump(const Duration(milliseconds: 50));
    expect(callCount, 1);
    expect(find.text('Fireball'), findsOneWidget);
  });

  testWidgets('does not fire per keystroke', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onGet('/api/compendium/search', (o) async => jsonBody([], 200));
    await pumpScreen(
      tester,
      const CompendiumScreen(),
      fake: fake,
      store: store,
    );
    for (final c in 'hello'.split('')) {
      await tester.enterText(find.byType(TextField), c);
      await tester.pump(const Duration(milliseconds: 50));
    }
    expect(fake.requests.where((r) => r.path.contains('compendium')).length, 0);
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();
    // only one request after debounce
    expect(fake.requests.where((r) => r.path.contains('compendium')).length, 1);
  });

  testWidgets('empty/no-match state renders', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onGet('/api/compendium/search', (o) async => jsonBody([], 200));
    await pumpScreen(
      tester,
      const CompendiumScreen(),
      fake: fake,
      store: store,
    );
    // initial empty prompts search
    expect(find.textContaining('Search the compendium'), findsOneWidget);
    await tester.enterText(find.byType(TextField), 'zzzznope');
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump(const Duration(milliseconds: 50));
    expect(find.textContaining('No results'), findsOneWidget);
  });
}
