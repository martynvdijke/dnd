import '../../core/api_client.dart';
import '../models/draft.dart';

class AiApi {
  final ApiClient client;
  AiApi(this.client);
  Future<DraftSession> createDraft(Map<String, dynamic> data) async {
    final r = await client.post('/api/ai/draft', data: data);
    return DraftSession.fromJson(r.data as Map<String, dynamic>);
  }

  Future<List<DraftSession>> listDrafts() async {
    final r = await client.get('/api/ai/draft');
    final d = r.data;
    if (d is List) {
      return d
          .map((e) => DraftSession.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<DraftSession> getDraft(String id) async {
    final r = await client.get('/api/ai/draft/$id');
    return DraftSession.fromJson(r.data as Map<String, dynamic>);
  }

  Future<DraftSession> turn(String id, Map<String, dynamic> data) async {
    final r = await client.post('/api/ai/draft/$id/turn', data: data);
    return DraftSession.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> commit(String id) async =>
      await client.post('/api/ai/draft/$id/commit');
  Future<void> deleteDraft(String id) async =>
      await client.delete('/api/ai/draft/$id');
}
