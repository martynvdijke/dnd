import '../../core/api_client.dart';
import '../models/combat_entry.dart';
import '../models/attack.dart';

class CombatApi {
  final ApiClient client;
  CombatApi(this.client);
  Future<List<CombatEntry>> list(int campaignId) async {
    final r = await client.get(
      '/api/combat',
      queryParameters: {'campaign_id': campaignId.toString()},
    );
    final data = r.data;
    if (data is List) {
      return data
          .map((e) => CombatEntry.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<void> create(CombatEntry e) async =>
      await client.post('/api/combat', data: e.toJson());
  Future<AttackResult> attack(AttackRequest req) async {
    final r = await client.post('/api/combat/attack', data: req.toJson());
    return AttackResult.fromJson(r.data as Map<String, dynamic>);
  }
}
