class Spell {
  final int id;
  final String name;
  final int level;
  final String school;
  final String castingTime;
  final String range;
  final String components;
  final String duration;
  final String description;
  final bool prepared;
  final bool alwaysPrepared;
  final String source;
  final String notes;
  const Spell({
    required this.id,
    required this.name,
    required this.level,
    required this.school,
    required this.castingTime,
    required this.range,
    required this.components,
    required this.duration,
    required this.description,
    required this.prepared,
    required this.alwaysPrepared,
    required this.source,
    required this.notes,
  });
  factory Spell.fromJson(Map<String, dynamic> j) => Spell(
    id: (j['id'] as num).toInt(),
    name: j['name'] as String? ?? '',
    level: (j['level'] as num?)?.toInt() ?? 0,
    school: j['school'] as String? ?? '',
    castingTime:
        j['casting_time'] as String? ?? j['castingTime'] as String? ?? '',
    range: j['range'] as String? ?? '',
    components: j['components'] as String? ?? '',
    duration: j['duration'] as String? ?? '',
    description: j['description'] as String? ?? '',
    prepared: j['prepared'] as bool? ?? false,
    alwaysPrepared:
        j['always_prepared'] as bool? ?? j['alwaysPrepared'] as bool? ?? false,
    source: j['source'] as String? ?? '',
    notes: j['notes'] as String? ?? '',
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'level': level,
    'school': school,
    'casting_time': castingTime,
    'range': range,
    'components': components,
    'duration': duration,
    'description': description,
    'prepared': prepared,
    'always_prepared': alwaysPrepared,
    'source': source,
    'notes': notes,
  };
  @override
  bool operator ==(Object other) =>
      other is Spell &&
      other.id == id &&
      other.name == name &&
      other.level == level &&
      other.school == school &&
      other.castingTime == castingTime &&
      other.range == range &&
      other.components == components &&
      other.duration == duration &&
      other.description == description &&
      other.prepared == prepared &&
      other.alwaysPrepared == alwaysPrepared &&
      other.source == source &&
      other.notes == notes;
  @override
  int get hashCode => Object.hashAll([
    id,
    name,
    level,
    school,
    castingTime,
    range,
    components,
    duration,
    description,
    prepared,
    alwaysPrepared,
    source,
    notes,
  ]);
}
