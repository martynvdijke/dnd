import '../../core/api_client.dart';
import '../models/campaign.dart';
import '../models/dm_screen.dart';

class CampaignApi {
  final ApiClient client;
  CampaignApi(this.client);
  Future<List<Campaign>> list() async {
    final r = await client.get('/api/campaigns');
    final d = r.data;
    if (d is List) {
      return d
          .map((e) => Campaign.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    if (d is Map && d['campaigns'] is List) {
      return (d['campaigns'] as List)
          .map((e) => Campaign.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<Campaign> get(int id) async {
    final r = await client.get('/api/campaigns/$id');
    return Campaign.fromJson(r.data as Map<String, dynamic>);
  }

  Future<DmScreen> dmScreen(int id) async {
    final r = await client.get('/api/campaigns/$id/dm-screen');
    return DmScreen.fromJson(r.data as Map<String, dynamic>);
  }
}
