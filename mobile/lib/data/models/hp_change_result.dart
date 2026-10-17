class HpChangeResult {
  final int hpCurrent;
  final int tempHp;
  final int hpMax;
  final int deathSavesSuccesses;
  final int deathSavesFailures;
  final Map<String, dynamic>? concentration;
  final String? defenseApplied;
  final int? effectiveDamage;
  final int? hpMaxEffective;
  const HpChangeResult({
    required this.hpCurrent,
    required this.tempHp,
    required this.hpMax,
    required this.deathSavesSuccesses,
    required this.deathSavesFailures,
    this.concentration,
    this.defenseApplied,
    this.effectiveDamage,
    this.hpMaxEffective,
  });
  factory HpChangeResult.fromJson(Map<String, dynamic> j) => HpChangeResult(
    hpCurrent: (j['hp_current'] as num).toInt(),
    tempHp: (j['temp_hp'] as num).toInt(),
    hpMax: (j['hp_max'] as num).toInt(),
    deathSavesSuccesses: (j['death_saves_successes'] as num).toInt(),
    deathSavesFailures: (j['death_saves_failures'] as num).toInt(),
    concentration: j['concentration'] as Map<String, dynamic>?,
    defenseApplied: j['defense_applied'] as String?,
    effectiveDamage: (j['effective_damage'] as num?)?.toInt(),
    hpMaxEffective: (j['hp_max_effective'] as num?)?.toInt(),
  );
  Map<String, dynamic> toJson() => {
    'hp_current': hpCurrent,
    'temp_hp': tempHp,
    'hp_max': hpMax,
    'death_saves_successes': deathSavesSuccesses,
    'death_saves_failures': deathSavesFailures,
    if (concentration != null) 'concentration': concentration,
    if (defenseApplied != null) 'defense_applied': defenseApplied,
    if (effectiveDamage != null) 'effective_damage': effectiveDamage,
    if (hpMaxEffective != null) 'hp_max_effective': hpMaxEffective,
  };
  @override
  bool operator ==(Object other) =>
      other is HpChangeResult &&
      other.hpCurrent == hpCurrent &&
      other.tempHp == tempHp &&
      other.hpMax == hpMax &&
      other.deathSavesSuccesses == deathSavesSuccesses &&
      other.deathSavesFailures == deathSavesFailures &&
      other.defenseApplied == defenseApplied &&
      other.effectiveDamage == effectiveDamage &&
      other.hpMaxEffective == hpMaxEffective;
  @override
  int get hashCode => Object.hashAll([
    hpCurrent,
    tempHp,
    hpMax,
    deathSavesSuccesses,
    deathSavesFailures,
    defenseApplied,
    effectiveDamage,
    hpMaxEffective,
  ]);
}
