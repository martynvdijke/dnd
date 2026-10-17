import '../../core/api_client.dart';
import '../models/character.dart';
import '../models/character_summary.dart';
import '../models/hp_change_result.dart';
import '../models/currency.dart';

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
        data: {
          'rest_type': restType,
          'hit_dice_count': ?hitDiceCount,
        },
      );
  Future<void> deathSave(int id) async =>
      await client.post('/api/characters/$id/death-save');
  Future<void> setExhaustion(int id, int level) async => await client.patch(
    '/api/characters/$id/exhaustion',
    data: {'level': level},
  );
  Future<void> castSpell(int id, int spellId, {int? slotLevel}) async =>
      await client.post(
        '/api/characters/$id/cast-spell',
        data: {
          'spell_id': spellId,
          'slot_level': ?slotLevel,
        },
      );
  Future<void> updateSpellcasting(int id, Map<String, dynamic> data) async =>
      await client.put('/api/characters/$id/spellcasting', data: data);
  Future<void> updateCurrency(int id, Currency c) async =>
      await client.put('/api/characters/$id/currency', data: c.toJson());
}
