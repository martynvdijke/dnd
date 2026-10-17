import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/auth_repository.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/features/dm_tools/dm_tools_screen.dart';

import '../helpers/fake_api.dart';
import '../helpers/pump.dart';

void main() {
  testWidgets('non-DM sees permission state and no DM API calls', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onGet('/api/campaigns', (o) async => jsonBody([], 200));
    await pumpScreen(
      tester,
      const DmToolsScreen(campaignId: 1),
      fake: fake,
      store: store,
      authState: const AuthAuthenticated(
        UserSession(id: 1, username: 'player', role: 'player'),
        'http://localhost:6270',
      ),
    );
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.textContaining('Dungeon Master'), findsOneWidget);
    // No dm-screen or ai calls
    expect(
      fake.requests.any(
        (r) => r.path.contains('dm-screen') || r.path.contains('/ai/'),
      ),
      isFalse,
    );
  });

  testWidgets('DM sees generator list', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    fake.onGet(
      '/api/campaigns',
      (o) async => jsonBody([
        {
          'id': 1,
          'user_id': 1,
          'name': 'Camp',
          'party_name': '',
          'description': '',
          'dm_notes': '',
          'created_at': '',
        },
      ], 200),
    );
    fake.onGet(
      '/api/campaigns/1/dm-screen',
      (o) async => jsonBody({
        'campaign': {'id': 1, 'name': 'Camp', 'party_name': '', 'dm_notes': ''},
        'combat': {'active': false, 'current_turn': '', 'entries': []},
        'party': [],
        'ambience': {'track': '', 'action': ''},
      }, 200),
    );
    fake.onGet(
      '/api/generate/npc',
      (o) async => jsonBody({'result': 'A gruff dwarf'}, 200),
    );
    await pumpScreen(
      tester,
      const DmToolsScreen(campaignId: 1),
      fake: fake,
      store: store,
      authState: const AuthAuthenticated(
        UserSession(id: 1, username: 'dm', role: 'dm'),
        'http://localhost:6270',
      ),
    );
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pump(const Duration(milliseconds: 100));
    // Generators tab is second; need to tap it — use ensureVisible for robustness
    final genTab = find.text('GENERATORS');
    await tester.ensureVisible(genTab);
    await tester.pumpAndSettle();
    await tester.tap(genTab, warnIfMissed: false);
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.byType(ChoiceChip), findsWidgets);
    expect(find.text('npc'), findsWidgets);
    expect(find.widgetWithText(FilledButton, 'Roll npc'), findsOneWidget);
  });
}
