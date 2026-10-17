import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/core/api_client.dart';
import 'package:dnd/core/auth_repository.dart';
import 'package:dnd/core/ws_client.dart';
import 'package:dnd/core/providers.dart';

import '../helpers/fake_api.dart';

void main() {
  test('BattlemapController moveToken optimistic and rolls back, sends normalised coords', () async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await store.write(
      const StoredAuth(
        serverUrl: 'http://localhost:6270',
        sessionId: 's',
        apiToken: 'vlt',
        userId: 1,
        username: 'u',
        role: 'player',
      ),
    );
    fake.onGet(
      '/api/campaigns/1/battlemap',
      (o) async => jsonBody(battlemapJson(campaignId: 1), 200),
    );
    fake.onPut(
      '/api/battlemap-tokens/10',
      (o) async => jsonBody({
        'id': 10,
        'campaign_id': 1,
        'name': 'Hero',
        'x': 0.7,
        'y': 0.8,
        'color': '#b8963e',
        'size': 1,
      }, 200),
    );
    final wsCtrl = StreamController<WsEvent>.broadcast();
    final container = ProviderContainer(
      overrides: [
        authStoreProvider.overrideWithValue(store),
        serverUrlProvider.overrideWithValue('http://localhost:6270'),
        apiClientProvider.overrideWithValue(
          ApiClient(
            baseUrl: 'http://localhost:6270',
            store: store,
            dio: fake.buildDio(),
          ),
        ),
        wsEventsProvider.overrideWith((ref) => wsCtrl.stream),
        authControllerProvider.overrideWith(() => _FakeAuth()),
      ],
    );
    addTearDown(() {
      container.dispose();
      wsCtrl.close();
    });
    await container.read(battlemapProvider(1).future);
    expect(container.read(battlemapProvider(1)).value!.tokens.first.x, 0.3);
    final ctrl = container.read(battlemapProvider(1).notifier);
    final fut = ctrl.moveToken(10, 0.7, 0.8);
    // optimistic
    expect(
      container.read(battlemapProvider(1)).value!.tokens.first.x,
      closeTo(0.7, 0.001),
    );
    expect(
      container.read(battlemapProvider(1)).value!.tokens.first.y,
      closeTo(0.8, 0.001),
    );
    await fut;
    // request body carries normalised 0..1
    final req = fake.requests
        .where((r) => r.path.contains('battlemap-tokens'))
        .first;
    expect((req.body as Map)['x'], closeTo(0.7, 0.001));
    expect((req.body as Map)['y'], closeTo(0.8, 0.001));
    expect((req.body as Map)['x'], greaterThanOrEqualTo(0));
    expect((req.body as Map)['x'], lessThanOrEqualTo(1));

    // failure rollback
    fake.routes.clear();
    fake.onPut(
      '/api/battlemap-tokens/10',
      (o) async => jsonBody({'error': 'fail'}, 500),
    );
    fake.requests.clear();
    // optimistic then fail
    try {
      await container
          .read(battlemapProvider(1).notifier)
          .moveToken(10, 0.1, 0.1);
    } catch (_) {}
    // rolled back to 0.7,0.8
    expect(
      container.read(battlemapProvider(1)).value!.tokens.first.x,
      closeTo(0.7, 0.001),
    );
  });

  test('WS battlemap_update triggers refetch', () async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await store.write(
      const StoredAuth(
        serverUrl: 'http://localhost:6270',
        sessionId: 's',
        apiToken: 'vlt',
        userId: 1,
        username: 'u',
        role: 'player',
      ),
    );
    int callCount = 0;
    fake.onGet('/api/campaigns/5/battlemap', (o) async {
      callCount++;
      final x = callCount == 1 ? 0.2 : 0.9;
      return jsonBody(
        battlemapJson(
          campaignId: 5,
          tokens: [
            {
              'id': 10,
              'campaign_id': 5,
              'name': 'Hero',
              'x': x,
              'y': 0.5,
              'color': '#b8963e',
              'size': 1,
            },
          ],
        ),
        200,
      );
    });
    final wsCtrl = StreamController<WsEvent>.broadcast();
    final container = ProviderContainer(
      overrides: [
        authStoreProvider.overrideWithValue(store),
        serverUrlProvider.overrideWithValue('http://localhost:6270'),
        apiClientProvider.overrideWithValue(
          ApiClient(
            baseUrl: 'http://localhost:6270',
            store: store,
            dio: fake.buildDio(),
          ),
        ),
        wsEventsProvider.overrideWith((ref) => wsCtrl.stream),
        authControllerProvider.overrideWith(() => _FakeAuth()),
      ],
    );
    addTearDown(() {
      container.dispose();
      wsCtrl.close();
    });
    // keep ws provider alive
    container.listen(wsEventsProvider, (prev, next) {});
    await container.read(battlemapProvider(5).future);
    expect(callCount, 1);
    expect(container.read(battlemapProvider(5)).value!.tokens.first.x, 0.2);
    wsCtrl.add(
      const WsEvent(type: 'battlemap_update', payload: {'campaign_id': 5}),
    );
    // allow async refetch
    await Future.delayed(const Duration(milliseconds: 200));
    // poll
    for (int i = 0; i < 10; i++) {
      await Future.delayed(const Duration(milliseconds: 50));
      if (callCount == 2) break;
    }
    expect(callCount, 2);
    expect(container.read(battlemapProvider(5)).value!.tokens.first.x, 0.9);
  });

  test('WS character_update triggers refetch for matching id', () async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await store.write(
      const StoredAuth(
        serverUrl: 'http://localhost:6270',
        sessionId: 's',
        apiToken: 'vlt',
        userId: 1,
        username: 'u',
        role: 'player',
      ),
    );
    int count = 0;
    fake.onGet('/api/characters/7', (o) async {
      count++;
      return jsonBody(characterJson(id: 7, name: count == 1 ? 'A' : 'B'), 200);
    });
    final wsCtrl = StreamController<WsEvent>.broadcast();
    final container = ProviderContainer(
      overrides: [
        authStoreProvider.overrideWithValue(store),
        serverUrlProvider.overrideWithValue('http://localhost:6270'),
        apiClientProvider.overrideWithValue(
          ApiClient(
            baseUrl: 'http://localhost:6270',
            store: store,
            dio: fake.buildDio(),
          ),
        ),
        wsEventsProvider.overrideWith((ref) => wsCtrl.stream),
        authControllerProvider.overrideWith(() => _FakeAuth()),
      ],
    );
    addTearDown(() {
      container.dispose();
      wsCtrl.close();
    });
    container.listen(wsEventsProvider, (prev, next) {});
    await container.read(characterProvider(7).future);
    expect(container.read(characterProvider(7)).value!.name, 'A');
    wsCtrl.add(
      const WsEvent(type: 'character_update', payload: {'character_id': 7}),
    );
    await Future.delayed(const Duration(milliseconds: 200));
    expect(count, 2);
    expect(container.read(characterProvider(7)).value!.name, 'B');
  });

  test('WS combat_update triggers combat refetch', () async {
    final fake = FakeApi();
    final store = InMemoryAuthStore();
    await store.write(
      const StoredAuth(
        serverUrl: 'http://localhost:6270',
        sessionId: 's',
        apiToken: 'vlt',
        userId: 1,
        username: 'u',
        role: 'player',
      ),
    );
    int count = 0;
    fake.onGet('/api/combat', (o) async {
      count++;
      return jsonBody([
        {
          'id': 1,
          'name': 'A',
          'type': 'pc',
          'initiative_roll': 10,
          'initiative_mod': 2,
          'hp_max': 10,
          'hp_current': 10,
          'ac': 12,
          'is_active': true,
          'turn_order': 0,
          'condition_ids': '',
          'notes': '',
          'action_used': false,
          'bonus_action_used': false,
          'reaction_used': false,
          'movement_used': 0,
          'movement_max': 30,
        },
      ], 200);
    });
    final wsCtrl = StreamController<WsEvent>.broadcast();
    final container = ProviderContainer(
      overrides: [
        authStoreProvider.overrideWithValue(store),
        serverUrlProvider.overrideWithValue('http://localhost:6270'),
        apiClientProvider.overrideWithValue(
          ApiClient(
            baseUrl: 'http://localhost:6270',
            store: store,
            dio: fake.buildDio(),
          ),
        ),
        wsEventsProvider.overrideWith((ref) => wsCtrl.stream),
        authControllerProvider.overrideWith(() => _FakeAuth()),
      ],
    );
    addTearDown(() {
      container.dispose();
      wsCtrl.close();
    });
    container.listen(wsEventsProvider, (prev, next) {});
    await container.read(combatProvider(9).future);
    expect(count, 1);
    wsCtrl.add(
      const WsEvent(type: 'combat_update', payload: {'campaign_id': 9}),
    );
    await Future.delayed(const Duration(milliseconds: 200));
    expect(count, 2);
  });
}

class _FakeAuth extends AuthController {
  @override
  AuthState build() => const AuthUnauthenticated();
}
