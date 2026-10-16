import '../../core/api_client.dart';
import '../models/battlemap.dart';

class BattlemapApi {
  final ApiClient client;
  BattlemapApi(this.client);
  Future<BattlemapState> get(int campaignId) async {
    final r = await client.get('/api/campaigns/$campaignId/battlemap');
    return BattlemapState.fromJson(r.data as Map<String, dynamic>);
  }

  Future<BattlemapToken> createToken(
    int campaignId,
    Map<String, dynamic> data,
  ) async {
    final r = await client.post(
      '/api/campaigns/$campaignId/battlemap/tokens',
      data: data,
    );
    if (r.data is Map && (r.data as Map).containsKey('token')) {
      return BattlemapToken.fromJson(
        (r.data as Map)['token'] as Map<String, dynamic>,
      );
    }
    return BattlemapToken.fromJson(r.data as Map<String, dynamic>);
  }

  Future<BattlemapToken> updateToken(
    int tokenId,
    Map<String, dynamic> data,
  ) async {
    final r = await client.put('/api/battlemap-tokens/$tokenId', data: data);
    if (r.data is Map && (r.data as Map).containsKey('token')) {
      return BattlemapToken.fromJson(
        (r.data as Map)['token'] as Map<String, dynamic>,
      );
    }
    return BattlemapToken.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> deleteToken(int tokenId) async =>
      await client.delete('/api/battlemap-tokens/$tokenId');
  Future<void> sync(int campaignId) async =>
      await client.post('/api/campaigns/$campaignId/battlemap/sync');
}
