import '../../core/api_client.dart';
import '../models/character.dart';
import '../models/character_summary.dart';
import '../models/hp_change_result.dart';
import '../models/currency.dart';
import '../models/inventory_item.dart';

/// Server answer to `POST /api/characters/:id/death-save`. The roll itself
/// happens on the server, so the app only displays what came back.
class DeathSaveResult {
  final int roll;
  final int successes;
  final int failures;
  final bool stable;
  final bool dead;
  final bool revived;
  final int hpCurrent;
  final String text;
  const DeathSaveResult({
    required this.roll,
    required this.successes,
    required this.failures,
    required this.stable,
    required this.dead,
    required this.revived,
    required this.hpCurrent,
    required this.text,
  });
  factory DeathSaveResult.fromJson(Map<String, dynamic> j) => DeathSaveResult(
    roll: (j['roll'] as num?)?.toInt() ?? 0,
    successes: (j['successes'] as num?)?.toInt() ?? 0,
    failures: (j['failures'] as num?)?.toInt() ?? 0,
    stable: j['stable'] as bool? ?? false,
    dead: j['dead'] as bool? ?? false,
    revived: j['revived'] as bool? ?? false,
    hpCurrent: (j['hp_current'] as num?)?.toInt() ?? 0,
    text: j['text'] as String? ?? '',
  );

  /// One-line summary for the sheet, built only from server-reported facts.
  String get summary {
    final parts = <String>[
      if (roll > 0) 'Rolled $roll',
      if (text.isNotEmpty) text,
      '$successes successes / $failures failures',
      if (stable) 'stable',
      if (dead) 'dead',
      if (revived) 'revived',
    ];
    return parts.join(' · ');
  }
}

class CharacterApi {
  final ApiClient client;
  CharacterApi(this.client);
  Future<List<CharacterSummary>> list() async {
    final r = await client.get('/api/characters');
    final data = r.data;
    if (data is List) {
      return data
          .map((e) => CharacterSummary.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    if (data is Map && data['characters'] is List) {
      return (data['characters'] as List)
          .map((e) => CharacterSummary.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<Character> get(int id) async {
    final r = await client.get('/api/characters/$id');
    return Character.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> update(int id, Map<String, dynamic> data) async =>
      await client.put('/api/characters/$id', data: data);
  Future<void> delete(int id) async =>
      await client.delete('/api/characters/$id');
  Future<HpChangeResult> changeHp(
    int id,
    int delta,
    String type,
    String source,
  ) async {
    final r = await client.post(
      '/api/characters/$id/hp',
      data: {'delta': delta, 'type': type, 'source': source},
    );
    return HpChangeResult.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> rest(int id, String restType, {int? hitDiceCount}) async =>
      await client.post(
        '/api/characters/$id/rest',
        data: {'rest_type': restType, 'hit_dice_count': ?hitDiceCount},
      );

  /// Rolls a death saving throw server-side; the response carries the roll and
  /// the persisted success/failure counters.
  Future<DeathSaveResult> deathSave(int id) async {
    final r = await client.post('/api/characters/$id/death-save');
    final data = r.data;
    return DeathSaveResult.fromJson(
      data is Map ? Map<String, dynamic>.from(data) : const {},
    );
  }

  Future<void> setExhaustion(int id, int level) async => await client.patch(
    '/api/characters/$id/exhaustion',
    data: {'level': level},
  );

  /// Expends a slot for the spell identified by [spellId]. The server consumes
  /// the slot at the spell's own level; [slotLevel] is sent as the requested
  /// level per the frozen contract and the response's `slot_consumed` is
  /// authoritative.
  Future<Map<String, dynamic>> castSpell(
    int id,
    int spellId, {
    int? slotLevel,
  }) async {
    final r = await client.post(
      '/api/characters/$id/cast-spell',
      data: {'spell_id': spellId, 'slot_level': ?slotLevel},
    );
    final data = r.data;
    return data is Map ? Map<String, dynamic>.from(data) : const {};
  }

  Future<void> updateSpellcasting(int id, Map<String, dynamic> data) async =>
      await client.put('/api/characters/$id/spellcasting', data: data);
  Future<void> updateCurrency(int id, Currency c) async =>
      await client.put('/api/characters/$id/currency', data: c.toJson());

  // ─── Inventory (`POST /characters/:id/inventory`, `PUT|DELETE /inventory/:iid`) ───

  /// Creates an item; returns the server-assigned id.
  Future<int> addInventoryItem(
    int characterId,
    Map<String, dynamic> data,
  ) async {
    final r = await client.post(
      '/api/characters/$characterId/inventory',
      data: data,
    );
    final payload = r.data;
    final id = payload is Map ? payload['id'] : null;
    return (id as num?)?.toInt() ?? 0;
  }

  /// Replaces an item wholesale — the endpoint is a full update, so callers
  /// must send the complete item (an equip toggle flips one field on the
  /// server's own copy).
  Future<void> updateInventoryItem(int itemId, InventoryItem item) async =>
      await client.put('/api/inventory/$itemId', data: item.toJson());

  Future<void> deleteInventoryItem(int itemId) async =>
      await client.delete('/api/inventory/$itemId');
}
