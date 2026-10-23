class InkPage {
  final int id;
  final int pageIndex;
  final int uploadId;
  final String url;
  final String thumbnailUrl;
  final String text;
  final String ocrSource;

  const InkPage({
    required this.id,
    required this.pageIndex,
    required this.uploadId,
    required this.url,
    required this.thumbnailUrl,
    required this.text,
    required this.ocrSource,
  });

  factory InkPage.fromJson(Map<String, dynamic> j) => InkPage(
    id: (j['id'] as num?)?.toInt() ?? 0,
    pageIndex: (j['page_index'] as num?)?.toInt() ?? 0,
    uploadId: (j['upload_id'] as num?)?.toInt() ?? 0,
    url: j['url'] as String? ?? '',
    thumbnailUrl: j['thumbnail_url'] as String? ?? '',
    text: j['text'] as String? ?? '',
    ocrSource: j['ocr_source'] as String? ?? 'none',
  );
}

class InkCapture {
  final int id;
  final int campaignId;
  final int? characterId;
  final int ownerUserId;
  final String title;
  final String sessionKey;
  final String source;
  final String status;
  final String error;
  final int? recapId;
  final int? journalEntryId;
  final List<InkPage> pages;

  const InkCapture({
    required this.id,
    required this.campaignId,
    required this.characterId,
    required this.ownerUserId,
    required this.title,
    required this.sessionKey,
    required this.source,
    required this.status,
    required this.error,
    required this.recapId,
    required this.journalEntryId,
    required this.pages,
  });

  factory InkCapture.fromJson(Map<String, dynamic> j) => InkCapture(
    id: (j['id'] as num?)?.toInt() ?? 0,
    campaignId: (j['campaign_id'] as num?)?.toInt() ?? 0,
    characterId: (j['character_id'] as num?)?.toInt(),
    ownerUserId: (j['owner_user_id'] as num?)?.toInt() ?? 0,
    title: j['title'] as String? ?? '',
    sessionKey: j['session_key'] as String? ?? '',
    source: j['source'] as String? ?? 'canvas',
    status: j['status'] as String? ?? 'uploaded',
    error: j['error'] as String? ?? '',
    recapId: (j['recap_id'] as num?)?.toInt(),
    journalEntryId: (j['journal_entry_id'] as num?)?.toInt(),
    pages: (j['pages'] as List? ?? const [])
        .map((e) => InkPage.fromJson(e as Map<String, dynamic>))
        .toList(),
  );

  String get text =>
      pages.map((p) => p.text.trim()).where((t) => t.isNotEmpty).join('\n\n');

  bool get isPromoted => recapId != null || journalEntryId != null;
}
