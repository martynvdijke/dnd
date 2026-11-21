import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/data/models/character.dart';

import '../helpers/fake_api.dart';

void main() {
  group('Character.fromJson sub-resources', () {
    test('parses spellcasting, spells, inventory, currency and exhaustion', () {
      final c = Character.fromJson(
        characterJson(
          exhaustionLevel: 2,
          deathSavesSuccesses: 1,
          deathSavesFailures: 2,
        ),
      );

      expect(c.exhaustionLevel, 2);
      expect(c.deathSavesSuccesses, 1);
      expect(c.deathSavesFailures, 2);

      expect(c.spellcasting, isNotNull);
      expect(c.spellcasting!.saveDc, 13);
      expect(c.spellcasting!.attackBonus, 5);
      expect(c.spellcasting!.slotsMax.length, 9);
      expect(c.spellcasting!.slotsMax[0], 4);
      expect(c.spellcasting!.slotsUsed[0], 0);

      expect(c.spells, hasLength(1));
      expect(c.spells.single.id, 1);
      expect(c.spells.single.prepared, isTrue);

      expect(c.inventory, hasLength(1));
      expect(c.inventory.single.name, 'Longsword');
      expect(c.inventory.single.quantity, 1);
      expect(c.inventory.single.isEquipped, isTrue);

      expect(c.currency, isNotNull);
      expect(c.currency!.gp, 50);

      expect(c.features.map((f) => f.name), contains('Second Wind'));
    });

    test('absent nullable sub-resources stay null and lists stay empty', () {
      final c = Character.fromJson(
        characterJson(
          spellcasting: null,
          currency: null,
          spells: [],
          inventory: [],
        ),
      );

      expect(c.spellcasting, isNull);
      expect(c.currency, isNull);
      expect(c.spells, isEmpty);
      expect(c.inventory, isEmpty);
    });

    test('renders the server slot usage for an expended character', () {
      final c = Character.fromJson(
        characterJson(
          spellcasting: spellcastingJson(slots1Max: 4, slots1Used: 3),
        ),
      );

      expect(c.spellcasting!.slotsMax[0], 4);
      expect(c.spellcasting!.slotsUsed[0], 3);
    });
  });
}
