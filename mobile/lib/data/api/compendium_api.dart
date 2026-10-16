import '../../core/api_client.dart';
import '../models/compendium.dart';

class CompendiumApi {
  final ApiClient client;
  CompendiumApi(this.client);
  Future<List<CompendiumSearchResult>> search({
    required String q,
    String? type,
  }) async {
    final r = await client.get(
      '/api/compendium/search',
      queryParameters: {'q': q, 'type': ?type},
    );
    final data = r.data;
    if (data is List) {
      return data
          .map(
            (e) => CompendiumSearchResult.fromJson(e as Map<String, dynamic>),
          )
          .toList();
    }
    if (data is Map && data['results'] is List) {
      return (data['results'] as List)
          .map(
            (e) => CompendiumSearchResult.fromJson(e as Map<String, dynamic>),
          )
          .toList();
    }
    return [];
  }
}
