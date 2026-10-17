class CharacterSummary {
  final int id;
  final int userId;
  final String name;
  final String race;
  final String class_;
  final int level;
  final int hpMax;
  final int hpCurrent;
  final String portraitUrl;
  final String raceColor;
  final String characterType;
  final bool canEdit;
  final List<dynamic> campaigns;
  const CharacterSummary({
    required this.id,
    required this.userId,
    required this.name,
    required this.race,
    required this.class_,
    required this.level,
    required this.hpMax,
    required this.hpCurrent,
    required this.portraitUrl,
    required this.raceColor,
    required this.characterType,
    required this.canEdit,
    required this.campaigns,
  });
  factory CharacterSummary.fromJson(Map<String, dynamic> j) => CharacterSummary(
    id: (j['id'] as num).toInt(),
    userId:
        (j['user_id'] as num?)?.toInt() ?? (j['userId'] as num?)?.toInt() ?? 0,
    name: j['name'] as String? ?? '',
    race: j['race'] as String? ?? '',
    class_: j['class'] as String? ?? j['class_'] as String? ?? '',
    level: (j['level'] as num?)?.toInt() ?? 1,
    hpMax: (j['hp_max'] as num?)?.toInt() ?? (j['hpMax'] as num?)?.toInt() ?? 0,
    hpCurrent:
        (j['hp_current'] as num?)?.toInt() ??
        (j['hpCurrent'] as num?)?.toInt() ??
        0,
    portraitUrl:
        j['portrait_url'] as String? ?? j['portraitUrl'] as String? ?? '',
    raceColor: j['race_color'] as String? ?? j['raceColor'] as String? ?? '',
    characterType:
        j['character_type'] as String? ?? j['characterType'] as String? ?? '',
    canEdit: j['can_edit'] as bool? ?? j['canEdit'] as bool? ?? false,
    campaigns: j['campaigns'] as List? ?? [],
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'user_id': userId,
    'name': name,
    'race': race,
    'class': class_,
    'level': level,
    'hp_max': hpMax,
    'hp_current': hpCurrent,
    'portrait_url': portraitUrl,
    'race_color': raceColor,
    'character_type': characterType,
    'can_edit': canEdit,
    'campaigns': campaigns,
  };
  @override
  bool operator ==(Object other) =>
      other is CharacterSummary &&
      other.id == id &&
      other.userId == userId &&
      other.name == name &&
      other.race == race &&
      other.class_ == class_ &&
      other.level == level &&
      other.hpMax == hpMax &&
      other.hpCurrent == hpCurrent &&
      other.portraitUrl == portraitUrl &&
      other.raceColor == raceColor &&
      other.characterType == characterType &&
      other.canEdit == canEdit;
  @override
  int get hashCode => Object.hashAll([
    id,
    userId,
    name,
    race,
    class_,
    level,
    hpMax,
    hpCurrent,
    portraitUrl,
    raceColor,
    characterType,
    canEdit,
  ]);
}
