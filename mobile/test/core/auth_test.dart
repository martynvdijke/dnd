import 'dart:convert';

import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/api_client.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/core/auth_repository.dart';
import 'package:dnd/core/ws_client.dart';
import 'package:dnd/data/models/character.dart';
import 'package:dnd/data/models/combat_entry.dart';
import 'package:dnd/data/models/battlemap.dart';
import 'package:dnd/data/models/hp_change_result.dart';

class FakeAdapter implements HttpClientAdapter {
  final Future<ResponseBody> Function(RequestOptions) handler;
  RequestOptions? lastRequest;
  FakeAdapter(this.handler);
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? s,
    Future<void>? c,
  ) async {
    lastRequest = options;
    return handler(options);
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

void main() {
  test('csrf derivation', () {
    final sess = 'abc123';
    final expected = sha256.convert(utf8.encode('$sess-csrf')).toString();
    expect(deriveCsrfToken(sess), expected);
  });

  test('Character fromJson roundtrip', () {
    final m = {
      'id': 1,
      'user_id': 2,
      'name': 'Gandalf',
      'race': 'Human',
      'class': 'Wizard',
      'subclass': '',
      'level': 5,
      'xp': 100,
      'background': 'Sage',
      'alignment': 'NG',
      'str': 10,
      'dex': 12,
      'con': 14,
      'int': 18,
      'wis': 14,
      'cha': 12,
      'ac': 15,
      'initiative': 2,
      'speed': 30,
      'hp_max': 30,
      'hp_current': 20,
      'temp_hp': 0,
      'hit_dice': '5d6',
      'hit_dice_current': 5,
      'proficiency_bonus': 3,
      'inspiration': false,
      'passive_perception': 12,
      'death_saves_successes': 0,
      'death_saves_failures': 0,
      'concentrating_on': '',
      'portrait_url': '',
      'character_type': 'pc',
      'can_edit': true,
    };
    final c = Character.fromJson(m);
    final j = c.toJson();
    final c2 = Character.fromJson(j);
    expect(c, c2);
  });

  test('CombatEntry roundtrip', () {
    final m = {
      'id': 1,
      'name': 'Goblin',
      'type': 'monster',
      'initiative_roll': 15,
      'initiative_mod': 2,
      'hp_max': 10,
      'hp_current': 10,
      'ac': 12,
      'is_active': true,
      'turn_order': 1,
      'condition_ids': '',
      'notes': '',
      'action_used': false,
      'bonus_action_used': false,
      'reaction_used': false,
      'movement_used': 0,
      'movement_max': 30,
    };
    final e = CombatEntry.fromJson(m);
    final j = e.toJson();
    final e2 = CombatEntry.fromJson(j);
    expect(e, e2);
  });

  test('BattlemapToken roundtrip', () {
    final m = {
      'id': 1,
      'campaign_id': 2,
      'name': 'Token',
      'x': 0.5,
      'y': 0.5,
      'color': 'red',
      'size': 1,
      'aura_radius': 0,
      'vision_radius': 0,
    };
    final t = BattlemapToken.fromJson(m);
    final j = t.toJson();
    final t2 = BattlemapToken.fromJson(j);
    expect(t, t2);
  });

  test('HpChangeResult roundtrip', () {
    final m = {
      'hp_current': 10,
      'temp_hp': 0,
      'hp_max': 20,
      'death_saves_successes': 0,
      'death_saves_failures': 0,
    };
    final r = HpChangeResult.fromJson(m);
    final j = r.toJson();
    final r2 = HpChangeResult.fromJson(j);
    expect(r, r2);
  });

  test(
    'ApiClient attaches cookie/csrf/bearer on POST but only cookie on GET',
    () async {
      final store = InMemoryAuthStore();
      await store.write(
        const StoredAuth(
          serverUrl: 'http://localhost:6270',
          sessionId: 'sess123',
          apiToken: 'vlt_abc',
          userId: 1,
          username: 'u',
          role: 'player',
        ),
      );
      final dio = Dio(
        BaseOptions(
          baseUrl: 'http://localhost:6270',
          validateStatus: (_) => true,
        ),
      );
      RequestOptions? getReq;
      RequestOptions? postReq;
      dio.httpClientAdapter = FakeAdapter((opts) async {
        if (opts.method == 'GET') getReq = opts;
        if (opts.method == 'POST') postReq = opts;
        return jsonBody({}, 200);
      });
      final client = ApiClient(
        baseUrl: 'http://localhost:6270',
        store: store,
        dio: dio,
      );
      await client.get('/api/characters');
      await client.post('/api/characters/1/hp', data: {'delta': -5});
      expect(getReq!.headers['Cookie'], 'session=sess123');
      expect(getReq!.headers.containsKey('X-CSRF-Token'), false);
      expect(getReq!.headers.containsKey('Authorization'), false);
      expect(postReq!.headers['Cookie'], 'session=sess123');
      expect(postReq!.headers['X-CSRF-Token'], deriveCsrfToken('sess123'));
      expect(postReq!.headers['Authorization'], 'Bearer vlt_abc');
    },
  );

  test('AuthRepository login happy path', () async {
    final store = InMemoryAuthStore();
    final dioLogin = Dio(BaseOptions(validateStatus: (_) => true));
    dioLogin.httpClientAdapter = FakeAdapter((opts) async {
      if (opts.path.contains('/api/login')) {
        return jsonBody({
          'session': 's1',
          'user': {'id': 1, 'username': 'bob', 'role': 'player'},
        }, 200);
      }
      if (opts.path.contains('/api/tokens')) {
        return jsonBody({'token': 'vlt_xyz'}, 201);
      }
      return jsonBody({}, 404);
    });
    final repo = AuthRepository(
      store: store,
      clientFactory: (url) =>
          ApiClient(baseUrl: url, store: store, dio: dioLogin),
    );
    final user = await repo.login(
      serverUrl: 'http://localhost:6270',
      username: 'bob',
      password: 'pw',
    );
    expect(user.username, 'bob');
    final stored = await store.read();
    expect(stored!.sessionId, 's1');
    expect(stored.apiToken, 'vlt_xyz');
  });

  test('AuthRepository login failure maps to ApiException', () async {
    final store = InMemoryAuthStore();
    final dioFail = Dio(BaseOptions(validateStatus: (_) => true));
    dioFail.httpClientAdapter = FakeAdapter(
      (opts) async => jsonBody({'error': 'invalid credentials'}, 401),
    );
    final repo = AuthRepository(
      store: store,
      clientFactory: (url) =>
          ApiClient(baseUrl: url, store: store, dio: dioFail),
    );
    expect(
      () => repo.login(
        serverUrl: 'http://localhost:6270',
        username: 'bob',
        password: 'wrong',
      ),
      throwsA(isA<ApiException>()),
    );
  });

  test('WsClient backoff cap', () {
    final store = InMemoryAuthStore();
    final ws = WsClient(baseUrl: 'http://localhost:6270', store: store);
    expect(ws.nextBackoffForTest(1).inMilliseconds, 1000);
    expect(ws.nextBackoffForTest(6).inMilliseconds, 30000);
    expect(ws.nextBackoffForTest(10).inMilliseconds, 30000);
  });

  test('ensureApiToken concurrent calls result in single POST', () async {
    final store = InMemoryAuthStore();
    await store.write(
      const StoredAuth(
        serverUrl: 'http://localhost:6270',
        sessionId: 'sess1',
        userId: 1,
        username: 'u',
        role: 'player',
      ),
    );
    int tokenCalls = 0;
    final dio = Dio(
      BaseOptions(
        baseUrl: 'http://localhost:6270',
        validateStatus: (_) => true,
      ),
    );
    dio.httpClientAdapter = FakeAdapter((opts) async {
      if (opts.path.contains('/api/tokens')) {
        tokenCalls++;
        // delay to allow concurrent overlap
        await Future.delayed(const Duration(milliseconds: 50));
        return jsonBody({'token': 'vlt_concurrent'}, 201);
      }
      return jsonBody({}, 200);
    });
    final client = ApiClient(
      baseUrl: 'http://localhost:6270',
      store: store,
      dio: dio,
    );
    await Future.wait([client.ensureApiToken(), client.ensureApiToken()]);
    expect(tokenCalls, 1);
    final stored = await store.read();
    expect(stored!.apiToken, 'vlt_concurrent');
  });
}
