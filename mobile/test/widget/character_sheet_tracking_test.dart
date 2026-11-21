import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/features/character_sheet/character_sheet_screen.dart';

import '../helpers/fake_api.dart';
import '../helpers/pump.dart';

Map<String, dynamic> _summary(Map<String, dynamic> character) => {
  'id': character['id'],
  'user_id': 1,
  'name': character['name'],
  'race': 'Human',
  'class': 'Wizard',
  'level': 3,
  'hp_max': character['hp_max'],
  'hp_current': character['hp_current'],
  'portrait_url': '',
  'race_color': '',
  'character_type': 'pc',
  'can_edit': true,
  'campaigns': [],
};

/// Pumps the sheet for [character] with the list, detail, and campaign routes
/// wired up. Every test builds its own character payload so the sub-resources
/// under test are the ones the server would send.
Future<void> _pumpSheet(
  WidgetTester tester, {
  required FakeApi fake,
  required InMemoryAuthStore store,
  Map<String, dynamic>? character,
}) async {
  final char = character ?? characterJson();
  final id = char['id'] as int;
  fake.onGet('/api/characters', (o) async => jsonBody([_summary(char)], 200));
  fake.onGet('/api/characters/$id', (o) async => jsonBody(char, 200));
  fake.onGet('/api/campaigns', (o) async => jsonBody([], 200));
  await pumpScreen(
    tester,
    CharacterSheetScreen(initialCharacterId: id),
    fake: fake,
    store: store,
  );
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 50));
}

Future<void> _openTab(WidgetTester tester, String label) async {
  await tester.tap(find.text(label));
  await tester.pumpAndSettle();
}

int _detailGets(FakeApi fake, int id) => fake.requests
    .where((r) => r.method == 'GET' && r.path == '/api/characters/$id')
    .length;

List<RecordedRequest> _requestsTo(FakeApi fake, String method, String path) =>
    fake.requests.where((r) => r.method == method && r.path == path).toList();

void main() {
  group('spell casting', () {
    testWidgets('slot row casts the parsed spell id at its slot level', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      // Level 1 only, so the Cast button under test is unambiguous.
      final character = characterJson(
        spellcasting: spellcastingJson(slots2Max: 0),
      );
      fake.onPost(
        '/api/characters/1/cast-spell',
        (o) async => jsonBody({'slot_consumed': true, 'slots_1_used': 1}, 200),
      );
      await _pumpSheet(tester, fake: fake, store: store, character: character);
      await _openTab(tester, 'SPELLS');

      expect(find.text('LVL 1'), findsOneWidget);
      expect(find.text('Cast'), findsOneWidget);
      final getsBefore = _detailGets(fake, 1);

      await tester.tap(find.text('Cast'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      final casts = _requestsTo(fake, 'POST', '/api/characters/1/cast-spell');
      expect(casts, hasLength(1));
      final body = casts.single.body as Map;
      expect(body['spell_id'], 1);
      expect(body['slot_level'], 1);
      // Success refetches so the pips show server truth.
      expect(_detailGets(fake, 1), greaterThan(getsBefore));
      expect(
        find.textContaining('Magic Missile cast at level 1'),
        findsOneWidget,
      );
    });

    testWidgets('spell tile offers upcast slot levels and posts the choice', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onPost(
        '/api/characters/1/cast-spell',
        (o) async => jsonBody({'slot_consumed': true}, 200),
      );
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'SPELLS');

      await tester.tap(find.text('Magic Missile'));
      await tester.pumpAndSettle();
      expect(find.text('Cast Magic Missile'), findsOneWidget);
      expect(find.text('Level 1 (spell level)'), findsOneWidget);
      expect(find.text('Level 2 (upcast)'), findsOneWidget);

      await tester.tap(find.text('Level 2 (upcast)'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      final casts = _requestsTo(fake, 'POST', '/api/characters/1/cast-spell');
      expect(casts, hasLength(1));
      final body = casts.single.body as Map;
      expect(body['spell_id'], 1);
      expect(body['slot_level'], 2);
    });

    testWidgets('a rejected cast surfaces the error and leaves slots alone', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onPost(
        '/api/characters/1/cast-spell',
        (o) async => jsonBody({'error': 'No slot available'}, 500),
      );
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'SPELLS');
      final getsBefore = _detailGets(fake, 1);

      await tester.tap(find.text('Cast').first);
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.textContaining('Cast failed:'), findsOneWidget);
      expect(find.textContaining('No slot available'), findsOneWidget);
      // No invalidation on failure, so the pips keep the server's state.
      expect(_detailGets(fake, 1), getsBefore);
      expect(find.text('0 / 4 used'), findsOneWidget);
    });

    testWidgets('a cast the server did not charge does not claim a slot', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onPost(
        '/api/characters/1/cast-spell',
        (o) async => jsonBody({'slot_consumed': false}, 200),
      );
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'SPELLS');

      await tester.tap(find.text('Cast').first);
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));

      expect(find.text('Magic Missile cast at level 1'), findsOneWidget);
      expect(find.textContaining('slot expended'), findsNothing);
    });
  });

  group('SPELLS tab rendering', () {
    testWidgets(
      'expended slots and the prepared/known split come from server',
      (tester) async {
        final fake = FakeApi();
        final store = InMemoryAuthStore();
        final character = characterJson(
          spellcasting: spellcastingJson(
            slots1Max: 4,
            slots1Used: 3,
            slots2Max: 0,
          ),
          spells: [
            {'id': 11, 'name': 'Magic Missile', 'level': 1, 'prepared': true},
            {'id': 12, 'name': 'Shield', 'level': 1, 'prepared': false},
          ],
        );
        await _pumpSheet(
          tester,
          fake: fake,
          store: store,
          character: character,
        );
        await _openTab(tester, 'SPELLS');

        expect(find.text('3 / 4 used'), findsOneWidget);
        expect(find.text('PREPARED SPELLS · 1'), findsOneWidget);
        expect(find.text('KNOWN SPELLS · 1'), findsOneWidget);
        expect(find.text('Magic Missile'), findsOneWidget);
        expect(find.text('Shield'), findsOneWidget);
        expect(find.text('Prepared'), findsOneWidget);
        // Level 2 has no slots, so no row is rendered for it.
        expect(find.text('LVL 2'), findsNothing);
      },
    );

    testWidgets('a spellcaster without slots is not called "no spellcasting"', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      final character = characterJson(
        spellcasting: spellcastingJson(slots1Max: 0, slots2Max: 0),
        spells: [],
      );
      await _pumpSheet(tester, fake: fake, store: store, character: character);
      await _openTab(tester, 'SPELLS');

      expect(find.text('No spell slots configured.'), findsOneWidget);
      expect(find.textContaining('No spellcasting'), findsNothing);
      expect(find.textContaining('Spellcasting · INT'), findsOneWidget);
    });

    testWidgets('a character without spellcasting says so', (tester) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      final character = characterJson(spellcasting: null);
      await _pumpSheet(tester, fake: fake, store: store, character: character);
      await _openTab(tester, 'SPELLS');

      expect(find.text('No spellcasting'), findsOneWidget);
      expect(find.textContaining('LVL 1'), findsNothing);
    });
  });

  group('inventory CRUD', () {
    testWidgets(
      'adding an item posts name, quantity and weight, then refetches',
      (tester) async {
        final fake = FakeApi();
        final store = InMemoryAuthStore();
        fake.onPost(
          '/api/characters/1/inventory',
          (o) async => jsonBody({'id': 2, 'name': 'Rope'}, 200),
        );
        await _pumpSheet(tester, fake: fake, store: store);
        await _openTab(tester, 'INVENTORY');
        final getsBefore = _detailGets(fake, 1);

        await tester.tap(find.text('Add item'));
        await tester.pumpAndSettle();
        await tester.enterText(find.widgetWithText(TextField, 'Name'), 'Rope');
        await tester.enterText(find.widgetWithText(TextField, 'Quantity'), '2');
        await tester.enterText(
          find.widgetWithText(TextField, 'Weight (lb each)'),
          '10',
        );
        await tester.pump();
        await tester.tap(find.widgetWithText(FilledButton, 'Save'));
        await tester.pumpAndSettle();

        final posts = _requestsTo(fake, 'POST', '/api/characters/1/inventory');
        expect(posts, hasLength(1));
        final body = posts.single.body as Map;
        expect(body['name'], 'Rope');
        expect(body['quantity'], 2);
        expect(body['weight'], 10);
        expect(body['is_equipped'], false);
        expect(_detailGets(fake, 1), greaterThan(getsBefore));
      },
    );

    testWidgets('editing an item preserves the untouched fields', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onPut(
        '/api/inventory/1',
        (o) async => jsonBody(inventoryItemJson(quantity: 3), 200),
      );
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'INVENTORY');
      final getsBefore = _detailGets(fake, 1);

      await tester.tap(find.text('Longsword ×1'));
      await tester.pumpAndSettle();
      expect(find.text('Edit item'), findsOneWidget);
      await tester.enterText(find.widgetWithText(TextField, 'Quantity'), '3');
      await tester.pump();
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();

      final puts = _requestsTo(fake, 'PUT', '/api/inventory/1');
      expect(puts, hasLength(1));
      final body = puts.single.body as Map;
      expect(body['quantity'], 3);
      // Everything the player did not touch is the server's copy, not a zero.
      expect(body['name'], 'Longsword');
      expect(body['weight'], 3);
      expect(body['category'], 'weapon');
      expect(body['damage_dice'], '1d8');
      expect(body['damage_type'], 'slashing');
      expect(body['is_equipped'], true);
      expect(body['attack_ability'], 'str');
      expect(body['attack_bonus'], 5);
      expect(_detailGets(fake, 1), greaterThan(getsBefore));
    });

    testWidgets('equip toggle changes only is_equipped', (tester) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onPut(
        '/api/inventory/1',
        (o) async => jsonBody(inventoryItemJson(isEquipped: false), 200),
      );
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'INVENTORY');

      await tester.tap(find.byTooltip('Unequip'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      final puts = _requestsTo(fake, 'PUT', '/api/inventory/1');
      expect(puts, hasLength(1));
      final body = Map<String, dynamic>.from(puts.single.body as Map);
      final expected = <String, dynamic>{
        ...inventoryItemJson(isEquipped: false),
        'weight': 3.0,
      };
      expect(body.keys.toSet(), expected.keys.toSet());
      expect(body, expected);
    });

    testWidgets('deleting an item calls the delete endpoint and refetches', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onDelete('/api/inventory/1', (o) async => jsonBody({}, 200));
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'INVENTORY');
      final getsBefore = _detailGets(fake, 1);

      await tester.tap(find.text('Longsword ×1'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();

      expect(_requestsTo(fake, 'DELETE', '/api/inventory/1'), hasLength(1));
      expect(_detailGets(fake, 1), greaterThan(getsBefore));
    });
  });

  group('currency', () {
    testWidgets('edit refetches before opening and saves the complete map', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      final character = characterJson(
        currency: {'cp': 0, 'sp': 3, 'ep': 0, 'gp': 50, 'pp': 1},
      );
      fake.onPut(
        '/api/characters/1/currency',
        (o) async => jsonBody(character['currency'], 200),
      );
      await _pumpSheet(tester, fake: fake, store: store, character: character);
      await _openTab(tester, 'INVENTORY');
      expect(find.text('50'), findsWidgets);

      final getsBefore = _detailGets(fake, 1);
      await tester.tap(find.text('Edit coin'));
      await tester.pumpAndSettle();
      // The dialog is seeded from a fresh read, not the possibly stale tab.
      expect(_detailGets(fake, 1), greaterThan(getsBefore));

      await tester.enterText(find.widgetWithText(TextField, 'CP'), '7');
      await tester.pump();
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();

      final puts = _requestsTo(fake, 'PUT', '/api/characters/1/currency');
      expect(puts, hasLength(1));
      expect(puts.single.body, {'cp': 7, 'sp': 3, 'ep': 0, 'gp': 50, 'pp': 1});
    });

    testWidgets('no server currency disables the coin editor', (tester) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      final character = characterJson(currency: null);
      await _pumpSheet(tester, fake: fake, store: store, character: character);
      await _openTab(tester, 'INVENTORY');

      expect(
        find.text('The server returned no coin for this character.'),
        findsOneWidget,
      );
      final button = tester.widget<OutlinedButton>(
        find.widgetWithText(OutlinedButton, 'Edit coin'),
      );
      expect(button.onPressed, isNull);
    });
  });

  group('death saves and exhaustion', () {
    testWidgets('rolling a death save shows the server result and refetches', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      final character = characterJson(hpCurrent: 0);
      fake.onPost(
        '/api/characters/1/death-save',
        (o) async => jsonBody({
          'roll': 14,
          'successes': 1,
          'failures': 0,
          'stable': false,
          'dead': false,
          'revived': false,
          'hp_current': 0,
          'text': 'Success',
        }, 200),
      );
      await _pumpSheet(tester, fake: fake, store: store, character: character);
      await _openTab(tester, 'COMBAT');
      final getsBefore = _detailGets(fake, 1);

      await tester.tap(find.text('Roll death save'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(
        _requestsTo(fake, 'POST', '/api/characters/1/death-save'),
        hasLength(1),
      );
      expect(find.textContaining('Rolled 14'), findsOneWidget);
      expect(find.textContaining('1 successes / 0 failures'), findsOneWidget);
      expect(_detailGets(fake, 1), greaterThan(getsBefore));
    });

    testWidgets('exhaustion patches the next level and shows server state', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      var serverLevel = 0;
      final character = characterJson(exhaustionLevel: serverLevel);
      fake.onGet(
        '/api/characters/1',
        (o) async => jsonBody(characterJson(exhaustionLevel: serverLevel), 200),
      );
      fake.onPatch('/api/characters/1/exhaustion', (o) async {
        serverLevel = ((o.data as Map)['level'] as num).toInt();
        return jsonBody({}, 200);
      });
      // Register the list/campaign routes the sheet needs as well.
      fake.onGet(
        '/api/characters',
        (o) async => jsonBody([_summary(character)], 200),
      );
      fake.onGet('/api/campaigns', (o) async => jsonBody([], 200));
      await pumpScreen(
        tester,
        const CharacterSheetScreen(initialCharacterId: 1),
        fake: fake,
        store: store,
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));
      await _openTab(tester, 'COMBAT');

      expect(find.text('Level 0 / 6'), findsOneWidget);
      await tester.ensureVisible(find.byTooltip('Raise exhaustion'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Raise exhaustion'));
      await tester.pumpAndSettle();

      final patches = _requestsTo(
        fake,
        'PATCH',
        '/api/characters/1/exhaustion',
      );
      expect(patches, hasLength(1));
      expect(patches.single.body, {'level': 1});
      expect(find.text('Level 1 / 6'), findsOneWidget);
    });

    testWidgets('exhaustion controls stop at the 0 and 6 bounds', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      final character = characterJson(exhaustionLevel: 6);
      fake.onPatch(
        '/api/characters/1/exhaustion',
        (o) async => jsonBody({}, 200),
      );
      await _pumpSheet(tester, fake: fake, store: store, character: character);
      await _openTab(tester, 'COMBAT');

      final raise = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.add_circle_outline),
      );
      final lower = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.remove_circle_outline),
      );
      expect(raise.onPressed, isNull, reason: 'cannot exceed 6');
      expect(lower.onPressed, isNotNull);

      await tester.ensureVisible(find.byTooltip('Lower exhaustion'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Lower exhaustion'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));
      expect(
        _requestsTo(fake, 'PATCH', '/api/characters/1/exhaustion').single.body,
        {'level': 5},
      );
    });

    testWidgets('exhaustion cannot go below zero', (tester) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'COMBAT');

      final lower = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.remove_circle_outline),
      );
      final raise = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.add_circle_outline),
      );
      expect(lower.onPressed, isNull, reason: 'cannot go below 0');
      expect(raise.onPressed, isNotNull);
    });

    testWidgets('a failed exhaustion update reports it and keeps the badge', (
      tester,
    ) async {
      final fake = FakeApi();
      final store = InMemoryAuthStore();
      fake.onPatch(
        '/api/characters/1/exhaustion',
        (o) async => jsonBody({'error': 'db down'}, 500),
      );
      await _pumpSheet(tester, fake: fake, store: store);
      await _openTab(tester, 'COMBAT');

      await tester.ensureVisible(find.byTooltip('Raise exhaustion'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Raise exhaustion'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.textContaining('Exhaustion update failed:'), findsOneWidget);
      expect(find.textContaining('db down'), findsOneWidget);
      expect(find.text('Level 0 / 6'), findsOneWidget);
    });
  });
}
