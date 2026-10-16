class CharClass {
  final int id;
  final String class_;
  final String subclass;
  final int level;
  final String hitDice;
  const CharClass({
    required this.id,
    required this.class_,
    required this.subclass,
    required this.level,
    required this.hitDice,
  });
  factory CharClass.fromJson(Map<String, dynamic> j) => CharClass(
    id: (j['id'] as num).toInt(),
    class_: j['class'] as String? ?? j['class_'] as String? ?? '',
    subclass: j['subclass'] as String? ?? '',
    level: (j['level'] as num?)?.toInt() ?? 1,
    hitDice: j['hit_dice'] as String? ?? j['hitDice'] as String? ?? '',
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'class': class_,
    'subclass': subclass,
    'level': level,
    'hit_dice': hitDice,
  };
  @override
  bool operator ==(Object other) =>
      other is CharClass &&
      other.id == id &&
      other.class_ == class_ &&
      other.subclass == subclass &&
      other.level == level &&
      other.hitDice == hitDice;
  @override
  int get hashCode => Object.hashAll([id, class_, subclass, level, hitDice]);
}
