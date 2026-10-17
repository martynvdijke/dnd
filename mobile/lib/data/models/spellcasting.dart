class Spellcasting {
  final String ability;
  final int saveDc;
  final int attackBonus;
  final List<int> slotsMax;
  final List<int> slotsUsed;
  const Spellcasting({
    required this.ability,
    required this.saveDc,
    required this.attackBonus,
    required this.slotsMax,
    required this.slotsUsed,
  });
  factory Spellcasting.fromJson(Map<String, dynamic> j) {
    List<int> max = List.generate(
      9,
      (i) => (j['slots_${i + 1}_max'] as num?)?.toInt() ?? 0,
    );
    List<int> used = List.generate(
      9,
      (i) => (j['slots_${i + 1}_used'] as num?)?.toInt() ?? 0,
    );
    return Spellcasting(
      ability: j['ability'] as String? ?? '',
      saveDc: (j['save_dc'] as num?)?.toInt() ?? 0,
      attackBonus: (j['attack_bonus'] as num?)?.toInt() ?? 0,
      slotsMax: max,
      slotsUsed: used,
    );
  }
  Map<String, dynamic> toJson() {
    final m = <String, dynamic>{
      'ability': ability,
      'save_dc': saveDc,
      'attack_bonus': attackBonus,
    };
    for (var i = 0; i < 9; i++) {
      m['slots_${i + 1}_max'] = slotsMax[i];
      m['slots_${i + 1}_used'] = slotsUsed[i];
    }
    return m;
  }

  @override
  bool operator ==(Object other) =>
      other is Spellcasting &&
      other.ability == ability &&
      other.saveDc == saveDc &&
      other.attackBonus == attackBonus &&
      other.slotsMax.length == slotsMax.length &&
      other.slotsUsed.length == slotsUsed.length &&
      _listEq(other.slotsMax, slotsMax) &&
      _listEq(other.slotsUsed, slotsUsed);
  @override
  int get hashCode => Object.hashAll([
    ability,
    saveDc,
    attackBonus,
    Object.hashAll(slotsMax),
    Object.hashAll(slotsUsed),
  ]);
  static bool _listEq(List<int> a, List<int> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }
}
