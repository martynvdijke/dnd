class CompendiumSearchResult {
  final String type;
  final String id;
  final String name;
  final String? subtype;
  final String? cr;
  final int? level;
  final String? hitDie;
  final String? primaryAbility;
  const CompendiumSearchResult({
    required this.type,
    required this.id,
    required this.name,
    this.subtype,
    this.cr,
    this.level,
    this.hitDie,
    this.primaryAbility,
  });
  factory CompendiumSearchResult.fromJson(Map<String, dynamic> j) =>
      CompendiumSearchResult(
        type: j['type'] as String? ?? '',
        id: j['id'] as String? ?? '',
        name: j['name'] as String? ?? '',
        subtype: j['subtype'] as String?,
        cr: j['cr'] as String?,
        level: (j['level'] as num?)?.toInt(),
        hitDie: j['hit_die'] as String? ?? j['hitDie'] as String?,
        primaryAbility:
            j['primary_ability'] as String? ?? j['primaryAbility'] as String?,
      );
  Map<String, dynamic> toJson() => {
    'type': type,
    'id': id,
    'name': name,
    if (subtype != null) 'subtype': subtype,
    if (cr != null) 'cr': cr,
    if (level != null) 'level': level,
    if (hitDie != null) 'hit_die': hitDie,
    if (primaryAbility != null) 'primary_ability': primaryAbility,
  };
  @override
  bool operator ==(Object other) =>
      other is CompendiumSearchResult &&
      other.type == type &&
      other.id == id &&
      other.name == name &&
      other.subtype == subtype &&
      other.cr == cr &&
      other.level == level &&
      other.hitDie == hitDie &&
      other.primaryAbility == primaryAbility;
  @override
  int get hashCode => Object.hashAll([
    type,
    id,
    name,
    subtype,
    cr,
    level,
    hitDie,
    primaryAbility,
  ]);
}
