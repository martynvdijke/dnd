class DmScreenCampaign {
  final int id;
  final String name;
  final String partyName;
  final String dmNotes;
  const DmScreenCampaign({
    required this.id,
    required this.name,
    required this.partyName,
    required this.dmNotes,
  });
  factory DmScreenCampaign.fromJson(Map<String, dynamic> j) => DmScreenCampaign(
    id: (j['id'] as num).toInt(),
    name: j['name'] as String? ?? '',
    partyName: j['party_name'] as String? ?? j['partyName'] as String? ?? '',
    dmNotes: j['dm_notes'] as String? ?? j['dmNotes'] as String? ?? '',
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'party_name': partyName,
    'dm_notes': dmNotes,
  };
  @override
  bool operator ==(Object other) => other is DmScreenCampaign && other.id == id;
  @override
  int get hashCode => id.hashCode;
}

class DmScreenCombat {
  final bool active;
  final String currentTurn;
  final List<Map<String, dynamic>> entries;
  const DmScreenCombat({
    required this.active,
    required this.currentTurn,
    required this.entries,
  });
  factory DmScreenCombat.fromJson(Map<String, dynamic> j) => DmScreenCombat(
    active: j['active'] as bool? ?? false,
    currentTurn:
        j['current_turn'] as String? ?? j['currentTurn'] as String? ?? '',
    entries: (j['entries'] as List? ?? [])
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList(),
  );
  Map<String, dynamic> toJson() => {
    'active': active,
    'current_turn': currentTurn,
    'entries': entries,
  };
}

class DmScreenParty {
  final int id;
  final String name;
  final int hpCurrent;
  final int hpMax;
  final int ac;
  final List<String> conditions;
  const DmScreenParty({
    required this.id,
    required this.name,
    required this.hpCurrent,
    required this.hpMax,
    required this.ac,
    required this.conditions,
  });
  factory DmScreenParty.fromJson(Map<String, dynamic> j) => DmScreenParty(
    id: (j['id'] as num).toInt(),
    name: j['name'] as String? ?? '',
    hpCurrent: (j['hp_current'] as num).toInt(),
    hpMax: (j['hp_max'] as num).toInt(),
    ac: (j['ac'] as num).toInt(),
    conditions: (j['conditions'] as List? ?? [])
        .map((e) => e.toString())
        .toList(),
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'hp_current': hpCurrent,
    'hp_max': hpMax,
    'ac': ac,
    'conditions': conditions,
  };
}

class DmScreenAmbience {
  final String track;
  final String action;
  const DmScreenAmbience({required this.track, required this.action});
  factory DmScreenAmbience.fromJson(Map<String, dynamic> j) => DmScreenAmbience(
    track: j['track'] as String? ?? '',
    action: j['action'] as String? ?? '',
  );
  Map<String, dynamic> toJson() => {'track': track, 'action': action};
}

class DmScreen {
  final DmScreenCampaign campaign;
  final Map<String, dynamic>? session;
  final DmScreenCombat combat;
  final List<DmScreenParty> party;
  final DmScreenAmbience ambience;
  const DmScreen({
    required this.campaign,
    this.session,
    required this.combat,
    required this.party,
    required this.ambience,
  });
  factory DmScreen.fromJson(Map<String, dynamic> j) => DmScreen(
    campaign: DmScreenCampaign.fromJson(j['campaign'] as Map<String, dynamic>),
    session: j['session'] as Map<String, dynamic>?,
    combat: DmScreenCombat.fromJson(j['combat'] as Map<String, dynamic>),
    party: (j['party'] as List)
        .map((e) => DmScreenParty.fromJson(e as Map<String, dynamic>))
        .toList(),
    ambience: DmScreenAmbience.fromJson(j['ambience'] as Map<String, dynamic>),
  );
  Map<String, dynamic> toJson() => {
    'campaign': campaign.toJson(),
    if (session != null) 'session': session,
    'combat': combat.toJson(),
    'party': party.map((e) => e.toJson()).toList(),
    'ambience': ambience.toJson(),
  };
}
