import 'dart:convert';

import 'package:dio/dio.dart';

class RecordedRequest {
  final String method;
  final String path;
  final Map<String, dynamic> headers;
  final dynamic body;
  final Map<String, dynamic>? queryParams;
  RecordedRequest({
    required this.method,
    required this.path,
    required this.headers,
    this.body,
    this.queryParams,
  });
}

class FakeApi {
  final List<RecordedRequest> requests = [];
  final Map<String, Future<ResponseBody> Function(RequestOptions)> routes = {};
  int Function(RequestOptions)? fallbackStatus = (_) => 404;

  void on(
    String method,
    String pathPrefix,
    Future<ResponseBody> Function(RequestOptions) handler,
  ) {
    routes['${method.toUpperCase()} $pathPrefix'] = handler;
  }

  void onGet(String prefix, Future<ResponseBody> Function(RequestOptions) h) =>
      on('GET', prefix, h);
  void onPost(String prefix, Future<ResponseBody> Function(RequestOptions) h) =>
      on('POST', prefix, h);
  void onPut(String prefix, Future<ResponseBody> Function(RequestOptions) h) =>
      on('PUT', prefix, h);
  void onDelete(
    String prefix,
    Future<ResponseBody> Function(RequestOptions) h,
  ) => on('DELETE', prefix, h);
  void onPatch(
    String prefix,
    Future<ResponseBody> Function(RequestOptions) h,
  ) => on('PATCH', prefix, h);

  HttpClientAdapter buildAdapter() {
    return _RecordingAdapter(this);
  }

  Dio buildDio({String baseUrl = 'http://localhost:6270'}) {
    final dio = Dio(BaseOptions(baseUrl: baseUrl, validateStatus: (_) => true));
    dio.httpClientAdapter = buildAdapter();
    return dio;
  }

  Future<ResponseBody> Function(RequestOptions)? match(RequestOptions opts) {
    final key = '${opts.method.toUpperCase()} ${opts.path}';
    // exact then prefix match
    if (routes.containsKey(key)) return routes[key];
    for (final entry in routes.entries) {
      final parts = entry.key.split(' ');
      final m = parts[0];
      final prefix = parts.sublist(1).join(' ');
      if (m == opts.method.toUpperCase() && opts.path.startsWith(prefix)) {
        return entry.value;
      }
    }
    return null;
  }
}

class _RecordingAdapter implements HttpClientAdapter {
  final FakeApi api;
  _RecordingAdapter(this.api);
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? s,
    Future<void>? c,
  ) async {
    api.requests.add(
      RecordedRequest(
        method: options.method,
        path: options.path,
        headers: Map<String, dynamic>.from(options.headers),
        body: options.data,
        queryParams: options.queryParameters.isEmpty
            ? null
            : Map<String, dynamic>.from(options.queryParameters),
      ),
    );
    final handler = api.match(options);
    if (handler != null) return handler(options);
    return ResponseBody.fromString(
      jsonEncode({'error': 'not mocked: ${options.method} ${options.path}'}),
      404,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

ResponseBody jsonBody(Object obj, int status) => ResponseBody.fromString(
  jsonEncode(obj),
  status,
  headers: {
    Headers.contentTypeHeader: [Headers.jsonContentType],
  },
);

/// Sentinel so [characterJson] can tell "argument omitted" (use the default
/// payload) apart from "explicitly null" (the server sent no sub-resource).
const Object _unset = Object();

/// The default `spellcasting` payload: level 1 and 2 slots available.
Map<String, dynamic> spellcastingJson({
  String ability = 'int',
  int saveDc = 13,
  int attackBonus = 5,
  int slots1Max = 4,
  int slots1Used = 0,
  int slots2Max = 2,
  int slots2Used = 0,
}) => {
  'ability': ability,
  'save_dc': saveDc,
  'attack_bonus': attackBonus,
  'slots_1_max': slots1Max,
  'slots_1_used': slots1Used,
  'slots_2_max': slots2Max,
  'slots_2_used': slots2Used,
  'slots_3_max': 0,
  'slots_3_used': 0,
  'slots_4_max': 0,
  'slots_4_used': 0,
  'slots_5_max': 0,
  'slots_5_used': 0,
  'slots_6_max': 0,
  'slots_6_used': 0,
  'slots_7_max': 0,
  'slots_7_used': 0,
  'slots_8_max': 0,
  'slots_8_used': 0,
  'slots_9_max': 0,
  'slots_9_used': 0,
};

Map<String, dynamic> characterJson({
  int id = 1,
  int hpCurrent = 20,
  int hpMax = 30,
  int tempHp = 0,
  String name = 'TestHero',
  int exhaustionLevel = 0,
  int deathSavesSuccesses = 0,
  int deathSavesFailures = 0,
  bool canEdit = true,
  Object? spellcasting = _unset,
  List<Map<String, dynamic>>? spells,
  List<Map<String, dynamic>>? inventory,
  Object? currency = _unset,
}) {
  return {
    'id': id,
    'user_id': 1,
    'name': name,
    'race': 'Human',
    'class': 'Fighter',
    'subclass': '',
    'level': 3,
    'xp': 900,
    'background': 'Soldier',
    'alignment': 'LG',
    'str': 16,
    'dex': 12,
    'con': 14,
    'int': 10,
    'wis': 12,
    'cha': 10,
    'ac': 16,
    'initiative': 1,
    'speed': 30,
    'hp_max': hpMax,
    'hp_current': hpCurrent,
    'temp_hp': tempHp,
    'hit_dice': '3d10',
    'hit_dice_current': 3,
    'proficiency_bonus': 2,
    'inspiration': 0,
    'passive_perception': 11,
    'death_saves_successes': deathSavesSuccesses,
    'death_saves_failures': deathSavesFailures,
    'exhaustion_level': exhaustionLevel,
    'concentrating_on': '',
    'portrait_url': '',
    'character_type': 'pc',
    'can_edit': canEdit,
    'spellcasting': identical(spellcasting, _unset)
        ? spellcastingJson()
        : spellcasting,
    'spells':
        spells ??
        [
          {
            'id': 1,
            'name': 'Magic Missile',
            'level': 1,
            'school': 'Evocation',
            'casting_time': '1 action',
            'range': '120 ft',
            'components': 'V,S',
            'duration': 'Instant',
            'description': ' pew',
            'prepared': true,
            'always_prepared': false,
            'source': 'class',
            'notes': '',
          },
        ],
    'inventory':
        inventory ??
        [
          {
            'id': 1,
            'name': 'Longsword',
            'quantity': 1,
            'weight': 3,
            'category': 'weapon',
            'damage_dice': '1d8',
            'damage_type': 'slashing',
            'weapon_properties': '',
            'ac_bonus': 0,
            'armor_type': '',
            'description': '',
            'is_equipped': true,
            'is_magical': false,
            'attunement': false,
            'is_identified': true,
            'notes': '',
            'attack_ability': 'str',
            'attack_bonus': 5,
          },
        ],
    'currency': identical(currency, _unset)
        ? {'cp': 0, 'sp': 0, 'ep': 0, 'gp': 50, 'pp': 0}
        : currency,
    'features': [
      {
        'id': 1,
        'name': 'Second Wind',
        'description': 'Recover HP',
        'source': 'Fighter',
        'level_gained': 1,
      },
    ],
    'classes': [],
    'campaigns': [],
  };
}

/// The server's `inventory` row shape, so tests can build their own list
/// without repeating every field.
Map<String, dynamic> inventoryItemJson({
  int id = 1,
  String name = 'Longsword',
  int quantity = 1,
  num weight = 3,
  String category = 'weapon',
  String damageDice = '1d8',
  String damageType = 'slashing',
  bool isEquipped = true,
}) => {
  'id': id,
  'name': name,
  'quantity': quantity,
  'weight': weight,
  'category': category,
  'damage_dice': damageDice,
  'damage_type': damageType,
  'weapon_properties': '',
  'ac_bonus': 0,
  'armor_type': '',
  'description': '',
  'is_equipped': isEquipped,
  'is_magical': false,
  'attunement': false,
  'is_identified': true,
  'notes': '',
  'attack_ability': 'str',
  'attack_bonus': 5,
};

Map<String, dynamic> battlemapJson({
  int campaignId = 1,
  List<Map<String, dynamic>>? tokens,
}) => {
  'map': {
    'id': 1,
    'name': 'Map1',
    'image_url': '',
    'width': 800,
    'height': 600,
    'grid_size': 50,
    'grid_units': 'ft',
    'walls': null,
  },
  'tokens':
      tokens ??
      [
        {
          'id': 10,
          'campaign_id': campaignId,
          'name': 'Hero',
          'x': 0.3,
          'y': 0.4,
          'color': '#b8963e',
          'size': 1,
          'hp_current': 20,
          'hp_max': 30,
          'ac': 15,
          'aura_radius': 0,
          'vision_radius': 0,
        },
      ],
};
