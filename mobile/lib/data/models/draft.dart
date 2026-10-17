class DraftMessage {
  final String role;
  final String content;
  const DraftMessage({required this.role, required this.content});
  factory DraftMessage.fromJson(Map<String, dynamic> j) => DraftMessage(
    role: j['role'] as String? ?? '',
    content: j['content'] as String? ?? '',
  );
  Map<String, dynamic> toJson() => {'role': role, 'content': content};
  @override
  bool operator ==(Object other) =>
      other is DraftMessage && other.role == role && other.content == content;
  @override
  int get hashCode => Object.hash(role, content);
}

class DraftSession {
  final String id;
  final String entityType;
  final String status;
  final String message;
  final Map<String, dynamic>? draft;
  final List<DraftMessage> messages;
  const DraftSession({
    required this.id,
    required this.entityType,
    required this.status,
    required this.message,
    this.draft,
    required this.messages,
  });
  factory DraftSession.fromJson(Map<String, dynamic> j) => DraftSession(
    id: j['id'].toString(),
    entityType: j['entity_type'] as String? ?? j['entityType'] as String? ?? '',
    status: j['status'] as String? ?? '',
    message: j['message'] as String? ?? '',
    draft: j['draft'] as Map<String, dynamic>?,
    messages: (j['messages'] as List? ?? [])
        .map((e) => DraftMessage.fromJson(e as Map<String, dynamic>))
        .toList(),
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'entity_type': entityType,
    'status': status,
    'message': message,
    if (draft != null) 'draft': draft,
    'messages': messages.map((e) => e.toJson()).toList(),
  };
  @override
  bool operator ==(Object other) => other is DraftSession && other.id == id;
  @override
  int get hashCode => id.hashCode;
}
