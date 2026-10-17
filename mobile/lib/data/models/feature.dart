class Feature {
  final int id;
  final String name;
  final String description;
  final String source;
  final int levelGained;
  const Feature({
    required this.id,
    required this.name,
    required this.description,
    required this.source,
    required this.levelGained,
  });
  factory Feature.fromJson(Map<String, dynamic> j) => Feature(
    id: (j['id'] as num).toInt(),
    name: j['name'] as String? ?? '',
    description: j['description'] as String? ?? '',
    source: j['source'] as String? ?? '',
    levelGained:
        (j['level_gained'] as num?)?.toInt() ??
        (j['levelGained'] as num?)?.toInt() ??
        0,
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'description': description,
    'source': source,
    'level_gained': levelGained,
  };
  @override
  bool operator ==(Object other) =>
      other is Feature &&
      other.id == id &&
      other.name == name &&
      other.description == description &&
      other.source == source &&
      other.levelGained == levelGained;
  @override
  int get hashCode =>
      Object.hashAll([id, name, description, source, levelGained]);
}
