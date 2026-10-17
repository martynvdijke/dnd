import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/features/combat/combat_screen.dart';

import '../helpers/fake_api.dart';
import '../helpers/pump.dart';

void main() {
  Future<void> setup(FakeApi fake) async {
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
      '/api/characters',
      (o) async => jsonBody([
        {
          'id': 10,
          'user_id': 1,
          'name': 'Attacker',
          'race': 'Human',
          'class': 'Fighter',
          'level': 5,
          'hp_max': 30,
          'hp_current': 30,
          'portrait_url': '',
          'race_color': '',
          'character_type': 'pc',
          'can_edit': true,
          'campaigns': [],
        },
      ], 200),
    );
    fake.onGet(
      '/api/combat',
      (o) async => jsonBody([
        {
          'id': 1,
          'name': 'Hero',
          'type': 'pc',
          'initiative_roll': 18,
          'initiative_mod': 2,
          'hp_max': 30,
          'hp_current': 30,
          'ac': 16,
          'is_active': true,
          'turn_order': 0,
          'condition_ids': '',
          'notes': '',
          'action_used': false,
          'bonus_action_used': false,
          'reaction_used': false,
          'movement_used': 0,
          'movement_max': 30,
        },
        {
          'id': 2,
          'name': 'Goblin',
          'type': 'monster',
          'initiative_roll': 12,
          'initiative_mod': 1,
          'hp_max': 10,
          'hp_current': 10,
          'ac': 13,
          'is_active': false,
          'turn_order': 1,
          'condition_ids': '',
          'notes': '',
          'action_used': false,
          'bonus_action_used': false,
          'reaction_used': false,
          'movement_used': 0,
          'movement_max': 30,
        },
      ], 200),
    );
    fake.onPost(
      '/api/combat/attack',
      (o) async => jsonBody({
        'attack_roll': 15,
        'attack_rolls': [15],
        'attack_bonus': 5,
        'attack_total': 20,
        'target_ac': 15,
        'cover_bonus': 0,
        'hit': true,
        'critical': false,
        'fumble': false,
        'cannot_target': false,
        'situational_advantage': '',
        'damage': 8,
        'damage_type': 'slashing',
        'damage_breakdown': '',
        'condition_applied': '',
        'log_id': 1,
      }, 200),
    );
  }

  testWidgets('initiative list renders and marks active', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await setup(fake);
    await pumpScreen(
      tester,
      const CombatScreen(campaignId: 1),
      fake: fake,
      store: store,
    );
    await tester.pump(const Duration(milliseconds: 150));
    // wait for async
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.text('Hero'), findsWidgets);
    expect(find.text('Goblin'), findsWidgets);
    expect(find.text('ACTIVE'), findsOneWidget);
    expect(find.byIcon(Icons.play_arrow_rounded), findsOneWidget);
  });

  testWidgets('attack roller submits and shows hit/miss plus damage', (
    tester,
  ) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await setup(fake);
    await pumpScreen(
      tester,
      const CombatScreen(campaignId: 1),
      fake: fake,
      store: store,
    );
    await tester.pump(const Duration(milliseconds: 150));
    // Find Roll Attack
    final btn = find.widgetWithText(FilledButton, 'Roll Attack');
    expect(btn, findsOneWidget);
    await tester.tap(btn);
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.textContaining('HIT'), findsOneWidget);
    // Damage shown
    expect(find.textContaining('Damage'), findsWidgets);
    expect(fake.requests.any((r) => r.path == '/api/combat/attack'), isTrue);
    final req = fake.requests.firstWhere((r) => r.path.contains('/attack'));
    expect((req.body as Map)['attacker_type'], 'character');
  });

  testWidgets('advantage/disadvantage toggles', (tester) async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await setup(fake);
    await pumpScreen(
      tester,
      const CombatScreen(campaignId: 1),
      fake: fake,
      store: store,
    );
    await tester.pump(const Duration(milliseconds: 150));
    expect(find.text('Normal'), findsOneWidget);
    expect(find.text('Adv'), findsOneWidget);
    expect(find.text('Dis'), findsOneWidget);
    await tester.tap(find.text('Adv'));
    await tester.pump();
    await tester.tap(find.widgetWithText(FilledButton, 'Roll Attack'));
    await tester.pump(const Duration(milliseconds: 100));
    var req = fake.requests.where((r) => r.path.contains('/attack')).last;
    expect((req.body as Map)['advantage'], 'advantage');
    await tester.tap(find.text('Dis'));
    await tester.pump();
    await tester.tap(find.widgetWithText(FilledButton, 'Roll Attack'));
    await tester.pump(const Duration(milliseconds: 100));
    req = fake.requests.where((r) => r.path.contains('/attack')).last;
    expect((req.body as Map)['advantage'], 'disadvantage');
  });
}
