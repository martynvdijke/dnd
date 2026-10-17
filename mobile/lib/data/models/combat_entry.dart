class CombatEntry {
  final int id;
  final int? characterId;
  final int? campaignId;
  final String name;
  final String type;
  final int initiativeRoll;
  final int initiativeMod;
  final int hpMax;
  final int hpCurrent;
  final int ac;
  final bool isActive;
  final int turnOrder;
  final String conditionIds;
  final String notes;
  final bool actionUsed;
  final bool bonusActionUsed;
  final bool reactionUsed;
  final int movementUsed;
  final int movementMax;
  const CombatEntry({
    required this.id,
    this.characterId,
    this.campaignId,
    required this.name,
    required this.type,
    required this.initiativeRoll,
    required this.initiativeMod,
    required this.hpMax,
    required this.hpCurrent,
    required this.ac,
    required this.isActive,
    required this.turnOrder,
    required this.conditionIds,
    required this.notes,
    required this.actionUsed,
    required this.bonusActionUsed,
    required this.reactionUsed,
    required this.movementUsed,
    required this.movementMax,
  });
  factory CombatEntry.fromJson(Map<String, dynamic> j) => CombatEntry(
    id: (j['id'] as num).toInt(),
    characterId: (j['character_id'] as num?)?.toInt(),
    campaignId: (j['campaign_id'] as num?)?.toInt(),
    name: j['name'] as String? ?? '',
    type: j['type'] as String? ?? '',
    initiativeRoll: (j['initiative_roll'] as num?)?.toInt() ?? 0,
    initiativeMod: (j['initiative_mod'] as num?)?.toInt() ?? 0,
    hpMax: (j['hp_max'] as num?)?.toInt() ?? 0,
    hpCurrent: (j['hp_current'] as num?)?.toInt() ?? 0,
    ac: (j['ac'] as num?)?.toInt() ?? 0,
    isActive: (() {
      final v = j['is_active'];
      if (v is bool) return v;
      if (v is num) return v != 0;
      return false;
    })(),
    turnOrder: (j['turn_order'] as num?)?.toInt() ?? 0,
    conditionIds: j['condition_ids'] as String? ?? '',
    notes: j['notes'] as String? ?? '',
    actionUsed: j['action_used'] as bool? ?? false,
    bonusActionUsed: j['bonus_action_used'] as bool? ?? false,
    reactionUsed: j['reaction_used'] as bool? ?? false,
    movementUsed: (j['movement_used'] as num?)?.toInt() ?? 0,
    movementMax: (j['movement_max'] as num?)?.toInt() ?? 30,
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    if (characterId != null) 'character_id': characterId,
    if (campaignId != null) 'campaign_id': campaignId,
    'name': name,
    'type': type,
    'initiative_roll': initiativeRoll,
    'initiative_mod': initiativeMod,
    'hp_max': hpMax,
    'hp_current': hpCurrent,
    'ac': ac,
    'is_active': isActive,
    'turn_order': turnOrder,
    'condition_ids': conditionIds,
    'notes': notes,
    'action_used': actionUsed,
    'bonus_action_used': bonusActionUsed,
    'reaction_used': reactionUsed,
    'movement_used': movementUsed,
    'movement_max': movementMax,
  };
  @override
  bool operator ==(Object other) =>
      other is CombatEntry &&
      other.id == id &&
      other.characterId == characterId &&
      other.campaignId == campaignId &&
      other.name == name &&
      other.type == type &&
      other.initiativeRoll == initiativeRoll &&
      other.initiativeMod == initiativeMod &&
      other.hpMax == hpMax &&
      other.hpCurrent == hpCurrent &&
      other.ac == ac &&
      other.isActive == isActive &&
      other.turnOrder == turnOrder &&
      other.conditionIds == conditionIds &&
      other.notes == notes &&
      other.actionUsed == actionUsed &&
      other.bonusActionUsed == bonusActionUsed &&
      other.reactionUsed == reactionUsed &&
      other.movementUsed == movementUsed &&
      other.movementMax == movementMax;
  @override
  int get hashCode => Object.hashAll([
    id,
    characterId,
    campaignId,
    name,
    type,
    initiativeRoll,
    initiativeMod,
    hpMax,
    hpCurrent,
    ac,
    isActive,
    turnOrder,
    conditionIds,
    notes,
    actionUsed,
    bonusActionUsed,
    reactionUsed,
    movementUsed,
    movementMax,
  ]);
}
