import 'dart:typed_data';

import 'package:dio/dio.dart';

import '../../core/api_client.dart';
import '../models/ink_capture.dart';

class InkApi {
  final ApiClient client;
  InkApi(this.client);

  Future<List<InkCapture>> list({
    required int campaignId,
    int? characterId,
  }) async {
    final r = await client.get(
      '/api/ink/captures',
      queryParameters: {
        'campaign_id': campaignId,
        'character_id': ?characterId,
      },
    );
    final data = r.data;
    if (data is List) {
      return data
          .map((e) => InkCapture.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<InkCapture> create({
    required int campaignId,
    int? characterId,
    String title = '',
    String sessionKey = '',
    String source = 'canvas',
    required List<Uint8List> pages,
    List<String> texts = const [],
  }) async {
    final form = FormData();
    form.fields.add(MapEntry('campaign_id', '$campaignId'));
    if (characterId != null) {
      form.fields.add(MapEntry('character_id', '$characterId'));
    }
    if (title.isNotEmpty) form.fields.add(MapEntry('title', title));
    if (sessionKey.isNotEmpty) {
      form.fields.add(MapEntry('session_key', sessionKey));
    }
    form.fields.add(MapEntry('source', source));
    for (final text in texts) {
      if (text.trim().isNotEmpty) form.fields.add(MapEntry('text', text));
    }
    for (var i = 0; i < pages.length; i++) {
      form.files.add(
        MapEntry(
          'pages',
          MultipartFile.fromBytes(
            pages[i],
            filename: 'page-$i.png',
            contentType: DioMediaType('image', 'png'),
          ),
        ),
      );
    }
    final r = await client.post('/api/ink/captures', data: form);
    final data = r.data as Map<String, dynamic>;
    return InkCapture.fromJson(data['capture'] as Map<String, dynamic>);
  }

  Future<InkCapture> ocr(int id) async {
    final r = await client.post('/api/ink/captures/$id/ocr', data: {});
    return InkCapture.fromJson(r.data as Map<String, dynamic>);
  }

  Future<InkCapture> updateText(int id, Map<int, String> pageTexts) async {
    final r = await client.patch(
      '/api/ink/captures/$id',
      data: {
        'pages': pageTexts.entries
            .map((e) => {'id': e.key, 'text': e.value})
            .toList(),
      },
    );
    return InkCapture.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> promote(int id, String target) async {
    await client.post(
      '/api/ink/captures/$id/promote',
      data: {'target': target},
    );
  }

  Future<void> delete(int id) async {
    await client.delete('/api/ink/captures/$id');
  }
}
