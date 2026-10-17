import '../../core/api_client.dart';

class GeneratorApi {
  final ApiClient client;
  GeneratorApi(this.client);
  Future<Map<String, dynamic>> generate(
    String kind, {
    Map<String, dynamic>? params,
  }) async {
    final r = await client.get('/api/generate/$kind', queryParameters: params);
    return r.data as Map<String, dynamic>;
  }
}
