import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/core/api_client.dart';
import 'package:dnd/core/auth_repository.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/core/providers.dart';

import '../helpers/fake_api.dart';

/// A store with a session already persisted, as if the app had been killed
/// after a previous sign-in.
class _SessionStore extends InMemoryAuthStore {
  _SessionStore(this.session);
  final StoredAuth? session;
  @override
  Future<StoredAuth?> read() async => session;
}

ProviderContainer _container(FakeApi fake, AuthStore store) {
  final client = ApiClient(
    baseUrl: 'http://villum.test',
    store: store,
    dio: fake.buildDio(baseUrl: 'http://villum.test'),
  );
  return ProviderContainer(
    overrides: [
      authStoreProvider.overrideWithValue(store),
      apiClientProvider.overrideWithValue(client),
      authRepositoryProvider.overrideWithValue(
        AuthRepository(store: store, clientFactory: (_) => client),
      ),
    ],
  );
}

const _stored = StoredAuth(
  serverUrl: 'http://villum.test',
  sessionId: 'sess_1',
  userId: 1,
  username: 'bob',
  role: 'player',
);

void main() {
  group('session restore at boot', () {
    test('a stored session starts the app authenticated', () async {
      final fake = FakeApi();
      final store = _SessionStore(_stored);
      fake.onGet(
        '/api/user/me',
        (o) async =>
            jsonBody({'id': 1, 'username': 'bob', 'role': 'player'}, 200),
      );
      final container = _container(fake, store);
      addTearDown(container.dispose);

      // Built synchronously as unauthenticated, then restored from storage.
      expect(
        container.read(authControllerProvider),
        isA<AuthUnauthenticated>(),
      );
      await pumpEventQueue();

      final state = container.read(authControllerProvider);
      expect(state, isA<AuthAuthenticated>());
      expect((state as AuthAuthenticated).serverUrl, _stored.serverUrl);
      expect(state.user.username, 'bob');
    });

    test(
      'no stored session stays unauthenticated without asking the server',
      () async {
        final fake = FakeApi();
        final store = _SessionStore(null);
        final container = _container(fake, store);
        addTearDown(container.dispose);

        await pumpEventQueue();

        expect(
          container.read(authControllerProvider),
          isA<AuthUnauthenticated>(),
        );
        expect(fake.requests, isEmpty);
      },
    );

    test('a rejected stored session falls back to onboarding', () async {
      final fake = FakeApi();
      final store = _SessionStore(_stored);
      fake.onGet('/api/user/me', (o) async => jsonBody({'error': 'nope'}, 401));
      final container = _container(fake, store);
      addTearDown(container.dispose);

      await pumpEventQueue();

      expect(
        container.read(authControllerProvider),
        isA<AuthUnauthenticated>(),
      );
    });
  });

  group('token provisioning', () {
    test('a failing token endpoint raises an actionable error', () async {
      final fake = FakeApi();
      final store = _SessionStore(_stored);
      var tokenCalls = 0;
      fake.onPost('/api/tokens', (o) async {
        tokenCalls++;
        return jsonBody({'error': 'server exploded'}, 500);
      });

      final client = ApiClient(
        baseUrl: _stored.serverUrl,
        store: store,
        dio: fake.buildDio(baseUrl: _stored.serverUrl),
      );

      await expectLater(
        client.ensureApiToken(),
        throwsA(
          isA<ApiException>()
              .having((e) => e.statusCode, 'statusCode', 500)
              .having(
                (e) => e.message,
                'message',
                allOf(
                  contains('Could not provision an API token'),
                  contains('HTTP 500'),
                  contains('server exploded'),
                  contains('retry'),
                ),
              ),
        ),
      );
      // One retry, then it surfaces instead of looping.
      expect(tokenCalls, 2);
      expect((await store.read())!.apiToken, isNull);
    });
  });
}
