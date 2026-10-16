class AttackRequest {
  final String attackerType;
  final int attackerId;
  final String targetType;
  final int targetId;
  final int? itemId;
  final int? attackBonus;
  final String? damageDice;
  final String? damageType;
  final String? advantage;
  final bool apply;
  final String? condition;
  final int? conditionDuration;
  final int? campaignId;
  final String? cover;
  final bool? flanking;
  final double? distance;
  final double? normalRange;
  final double? longRange;
  const AttackRequest({
    required this.attackerType,
    required this.attackerId,
    required this.targetType,
    required this.targetId,
    this.itemId,
    this.attackBonus,
    this.damageDice,
    this.damageType,
    this.advantage,
    this.apply = true,
    this.condition,
    this.conditionDuration,
    this.campaignId,
    this.cover,
    this.flanking,
    this.distance,
    this.normalRange,
    this.longRange,
  });
  factory AttackRequest.fromJson(Map<String, dynamic> j) => AttackRequest(
    attackerType:
        j['attacker_type'] as String? ?? j['attackerType'] as String? ?? '',
    attackerId:
        (j['attacker_id'] as num?)?.toInt() ??
        (j['attackerId'] as num?)?.toInt() ??
        0,
    targetType: j['target_type'] as String? ?? j['targetType'] as String? ?? '',
    targetId:
        (j['target_id'] as num?)?.toInt() ??
        (j['targetId'] as num?)?.toInt() ??
        0,
    itemId: (j['item_id'] as num?)?.toInt(),
    attackBonus: (j['attack_bonus'] as num?)?.toInt(),
    damageDice: j['damage_dice'] as String?,
    damageType: j['damage_type'] as String?,
    advantage: j['advantage'] as String?,
    apply: j['apply'] as bool? ?? true,
    condition: j['condition'] as String?,
    conditionDuration: (j['condition_duration'] as num?)?.toInt(),
    campaignId: (j['campaign_id'] as num?)?.toInt(),
    cover: j['cover'] as String?,
    flanking: j['flanking'] as bool?,
    distance: (j['distance'] as num?)?.toDouble(),
    normalRange: (j['normal_range'] as num?)?.toDouble(),
    longRange: (j['long_range'] as num?)?.toDouble(),
  );
  Map<String, dynamic> toJson() => {
    'attacker_type': attackerType,
    'attacker_id': attackerId,
    'target_type': targetType,
    'target_id': targetId,
    if (itemId != null) 'item_id': itemId,
    if (attackBonus != null) 'attack_bonus': attackBonus,
    if (damageDice != null) 'damage_dice': damageDice,
    if (damageType != null) 'damage_type': damageType,
    if (advantage != null) 'advantage': advantage,
    'apply': apply,
    if (condition != null) 'condition': condition,
    if (conditionDuration != null) 'condition_duration': conditionDuration,
    if (campaignId != null) 'campaign_id': campaignId,
    if (cover != null) 'cover': cover,
    if (flanking != null) 'flanking': flanking,
    if (distance != null) 'distance': distance,
    if (normalRange != null) 'normal_range': normalRange,
    if (longRange != null) 'long_range': longRange,
  };
  @override
  bool operator ==(Object other) =>
      other is AttackRequest &&
      other.attackerType == attackerType &&
      other.attackerId == attackerId &&
      other.targetType == targetType &&
      other.targetId == targetId &&
      other.itemId == itemId &&
      other.attackBonus == attackBonus &&
      other.damageDice == damageDice &&
      other.damageType == damageType &&
      other.advantage == advantage &&
      other.apply == apply &&
      other.condition == condition &&
      other.conditionDuration == conditionDuration &&
      other.campaignId == campaignId &&
      other.cover == cover &&
      other.flanking == flanking &&
      other.distance == distance &&
      other.normalRange == normalRange &&
      other.longRange == longRange;
  @override
  int get hashCode => Object.hashAll([
    attackerType,
    attackerId,
    targetType,
    targetId,
    itemId,
    attackBonus,
    damageDice,
    damageType,
    advantage,
    apply,
    condition,
    conditionDuration,
    campaignId,
    cover,
    flanking,
    distance,
    normalRange,
    longRange,
  ]);
}

class AttackResult {
  final int attackRoll;
  final List<int> attackRolls;
  final int attackBonus;
  final int attackTotal;
  final int targetAc;
  final int coverBonus;
  final bool hit;
  final bool critical;
  final bool fumble;
  final bool cannotTarget;
  final String situationalAdvantage;
  final int damage;
  final String damageType;
  final String damageBreakdown;
  final String conditionApplied;
  final int logId;
  final bool? applied;
  final int? targetHp;
  const AttackResult({
    required this.attackRoll,
    required this.attackRolls,
    required this.attackBonus,
    required this.attackTotal,
    required this.targetAc,
    required this.coverBonus,
    required this.hit,
    required this.critical,
    required this.fumble,
    required this.cannotTarget,
    required this.situationalAdvantage,
    required this.damage,
    required this.damageType,
    required this.damageBreakdown,
    required this.conditionApplied,
    required this.logId,
    this.applied,
    this.targetHp,
  });
  factory AttackResult.fromJson(Map<String, dynamic> j) => AttackResult(
    attackRoll: (j['attack_roll'] as num?)?.toInt() ?? 0,
    attackRolls:
        (j['attack_rolls'] as List?)?.map((e) => (e as num).toInt()).toList() ??
        [],
    attackBonus: (j['attack_bonus'] as num?)?.toInt() ?? 0,
    attackTotal: (j['attack_total'] as num?)?.toInt() ?? 0,
    targetAc: (j['target_ac'] as num?)?.toInt() ?? 0,
    coverBonus: (j['cover_bonus'] as num?)?.toInt() ?? 0,
    hit: j['hit'] as bool? ?? false,
    critical: j['critical'] as bool? ?? false,
    fumble: j['fumble'] as bool? ?? false,
    cannotTarget: j['cannot_target'] as bool? ?? false,
    situationalAdvantage: j['situational_advantage'] as String? ?? '',
    damage: (j['damage'] as num?)?.toInt() ?? 0,
    damageType: j['damage_type'] as String? ?? '',
    damageBreakdown: j['damage_breakdown'] as String? ?? '',
    conditionApplied: j['condition_applied'] as String? ?? '',
    logId: (j['log_id'] as num?)?.toInt() ?? (j['logId'] as num?)?.toInt() ?? 0,
    applied: j['applied'] as bool?,
    targetHp: (j['target_hp'] as num?)?.toInt(),
  );
  Map<String, dynamic> toJson() => {
    'attack_roll': attackRoll,
    'attack_rolls': attackRolls,
    'attack_bonus': attackBonus,
    'attack_total': attackTotal,
    'target_ac': targetAc,
    'cover_bonus': coverBonus,
    'hit': hit,
    'critical': critical,
    'fumble': fumble,
    'cannot_target': cannotTarget,
    'situational_advantage': situationalAdvantage,
    'damage': damage,
    'damage_type': damageType,
    'damage_breakdown': damageBreakdown,
    'condition_applied': conditionApplied,
    'log_id': logId,
    if (applied != null) 'applied': applied,
    if (targetHp != null) 'target_hp': targetHp,
  };
  @override
  bool operator ==(Object other) =>
      other is AttackResult &&
      other.attackRoll == attackRoll &&
      other.attackRolls.length == attackRolls.length &&
      _listEq(other.attackRolls, attackRolls) &&
      other.attackBonus == attackBonus &&
      other.attackTotal == attackTotal &&
      other.targetAc == targetAc &&
      other.coverBonus == coverBonus &&
      other.hit == hit &&
      other.critical == critical &&
      other.fumble == fumble &&
      other.cannotTarget == cannotTarget &&
      other.situationalAdvantage == situationalAdvantage &&
      other.damage == damage &&
      other.damageType == damageType &&
      other.damageBreakdown == damageBreakdown &&
      other.conditionApplied == conditionApplied &&
      other.logId == logId &&
      other.applied == applied &&
      other.targetHp == targetHp;
  @override
  int get hashCode => Object.hashAll([
    attackRoll,
    Object.hashAll(attackRolls),
    attackBonus,
    attackTotal,
    targetAc,
    coverBonus,
    hit,
    critical,
    fumble,
    cannotTarget,
    situationalAdvantage,
    damage,
    damageType,
    damageBreakdown,
    conditionApplied,
    logId,
    applied,
    targetHp,
  ]);

  static bool _listEq(List<int> a, List<int> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }
}
