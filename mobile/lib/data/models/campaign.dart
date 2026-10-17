class Campaign {
  final int id;
  final int userId;
  final String name;
  final String partyName;
  final String description;
  final String dmNotes;
  final String createdAt;
  const Campaign({
    required this.id,
    required this.userId,
    required this.name,
    required this.partyName,
    required this.description,
    required this.dmNotes,
    required this.createdAt,
  });
  factory Campaign.fromJson(Map<String, dynamic> j) => Campaign(
    id: (j['id'] as num).toInt(),
    userId:
        (j['user_id'] as num?)?.toInt() ?? (j['userId'] as num?)?.toInt() ?? 0,
    name: j['name'] as String? ?? '',
    partyName: j['party_name'] as String? ?? j['partyName'] as String? ?? '',
    description: j['description'] as String? ?? '',
    dmNotes: j['dm_notes'] as String? ?? j['dmNotes'] as String? ?? '',
    createdAt: j['created_at'] as String? ?? j['createdAt'] as String? ?? '',
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'user_id': userId,
    'name': name,
    'party_name': partyName,
    'description': description,
    'dm_notes': dmNotes,
    'created_at': createdAt,
  };
  @override
  bool operator ==(Object other) =>
      other is Campaign &&
      other.id == id &&
      other.userId == userId &&
      other.name == name &&
      other.partyName == partyName &&
      other.description == description &&
      other.dmNotes == dmNotes &&
      other.createdAt == createdAt;
  @override
  int get hashCode => Object.hashAll([
    id,
    userId,
    name,
    partyName,
    description,
    dmNotes,
    createdAt,
  ]);
}
