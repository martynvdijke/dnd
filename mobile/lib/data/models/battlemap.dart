class BattlemapMeta {
  final int id;
  final String name;
  final String imageUrl;
  final int width;
  final int height;
  final int gridSize;
  final String gridUnits;
  final dynamic walls;
  const BattlemapMeta({
    required this.id,
    required this.name,
    required this.imageUrl,
    required this.width,
    required this.height,
    required this.gridSize,
    required this.gridUnits,
    this.walls,
  });
  factory BattlemapMeta.fromJson(Map<String, dynamic> j) => BattlemapMeta(
    id: (j['id'] as num).toInt(),
    name: j['name'] as String? ?? '',
    imageUrl: j['image_url'] as String? ?? j['imageUrl'] as String? ?? '',
    width: (j['width'] as num?)?.toInt() ?? 0,
    height: (j['height'] as num?)?.toInt() ?? 0,
    gridSize: (j['grid_size'] as num?)?.toInt() ?? 0,
    gridUnits: j['grid_units'] as String? ?? j['gridUnits'] as String? ?? 'ft',
    walls: j['walls'],
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'image_url': imageUrl,
    'width': width,
    'height': height,
    'grid_size': gridSize,
    'grid_units': gridUnits,
    'walls': walls,
  };
  @override
  bool operator ==(Object other) =>
      other is BattlemapMeta && other.id == id && other.name == name;
  @override
  int get hashCode => Object.hash(id, name);
}

class BattlemapToken {
  final int id;
  final int campaignId;
  final int? combatEntryId;
  final String name;
  final double x;
  final double y;
  final String color;
  final double size;
  final int? hpCurrent;
  final int? hpMax;
  final int? ac;
  final String? conditions;
  final double auraRadius;
  final double visionRadius;
  const BattlemapToken({
    required this.id,
    required this.campaignId,
    this.combatEntryId,
    required this.name,
    required this.x,
    required this.y,
    required this.color,
    required this.size,
    this.hpCurrent,
    this.hpMax,
    this.ac,
    this.conditions,
    this.auraRadius = 0,
    this.visionRadius = 0,
  });
  factory BattlemapToken.fromJson(Map<String, dynamic> j) => BattlemapToken(
    id: (j['id'] as num).toInt(),
    campaignId: (j['campaign_id'] as num).toInt(),
    combatEntryId: (j['combat_entry_id'] as num?)?.toInt(),
    name: j['name'] as String? ?? '',
    x: (j['x'] as num).toDouble(),
    y: (j['y'] as num).toDouble(),
    color: j['color'] as String? ?? '',
    size: (j['size'] as num?)?.toDouble() ?? 1,
    hpCurrent: (j['hp_current'] as num?)?.toInt(),
    hpMax: (j['hp_max'] as num?)?.toInt(),
    ac: (j['ac'] as num?)?.toInt(),
    conditions: j['conditions'] as String?,
    auraRadius:
        (j['aura_radius'] as num?)?.toDouble() ??
        (j['auraRadius'] as num?)?.toDouble() ??
        0,
    visionRadius:
        (j['vision_radius'] as num?)?.toDouble() ??
        (j['visionRadius'] as num?)?.toDouble() ??
        0,
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'campaign_id': campaignId,
    if (combatEntryId != null) 'combat_entry_id': combatEntryId,
    'name': name,
    'x': x,
    'y': y,
    'color': color,
    'size': size,
    if (hpCurrent != null) 'hp_current': hpCurrent,
    if (hpMax != null) 'hp_max': hpMax,
    if (ac != null) 'ac': ac,
    if (conditions != null) 'conditions': conditions,
    'aura_radius': auraRadius,
    'vision_radius': visionRadius,
  };
  @override
  bool operator ==(Object other) =>
      other is BattlemapToken && other.id == id && other.x == x && other.y == y;
  @override
  int get hashCode => Object.hashAll([
    id,
    campaignId,
    combatEntryId,
    name,
    x,
    y,
    color,
    size,
    hpCurrent,
    hpMax,
    ac,
    conditions,
    auraRadius,
    visionRadius,
  ]);
}

class BattlemapState {
  final BattlemapMeta? map;
  final List<BattlemapToken> tokens;
  const BattlemapState({this.map, required this.tokens});
  factory BattlemapState.fromJson(Map<String, dynamic> j) => BattlemapState(
    map: j['map'] == null
        ? null
        : BattlemapMeta.fromJson(j['map'] as Map<String, dynamic>),
    tokens: (j['tokens'] as List? ?? [])
        .map((e) => BattlemapToken.fromJson(e as Map<String, dynamic>))
        .toList(),
  );
  Map<String, dynamic> toJson() => {
    'map': map?.toJson(),
    'tokens': tokens.map((e) => e.toJson()).toList(),
  };
  @override
  bool operator ==(Object other) =>
      other is BattlemapState &&
      other.map == map &&
      _listEq(other.tokens, tokens);
  @override
  int get hashCode => Object.hash(map, Object.hashAll(tokens));
  static bool _listEq(List<BattlemapToken> a, List<BattlemapToken> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }
}
