import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:dnd/core/api_client.dart';
import 'package:dnd/core/auth_repository.dart';
import 'package:dnd/core/auth_store.dart';

class _Adapter implements HttpClientAdapter {
  _Adapter(this.handler);
  final Future<ResponseBody> Function(RequestOptions) handler;
  final List<RequestOptions> requests = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? s,
    Future<void>? c,
  ) async {
    requests.add(options);
    return handler(options);
  }

  @override
  void close({bool force = false}) {}
}

ResponseBody _json(Object obj, int status) => ResponseBody.fromString(
  jsonEncode(obj),
  status,
  headers: {
    Headers.contentTypeHeader: [Headers.jsonContentType],
  },
);

AuthRepository _repo(
  InMemoryAuthStore store,
  Future<ResponseBody> Function(RequestOptions) handler,
) {
  final adapter = _Adapter(handler);
  return AuthRepository(
    store: store,
    clientFactory: (url) {
      final dio = Dio(
        BaseOptions(baseUrl: url, validateStatus: (_) => true),
      )..httpClientAdapter = adapter;
      return ApiClient(baseUrl: url, store: store, dio: dio);
    },
  );
}

void main() {
  test('loginWithSession persists the cookie and resolves the user', () async {
    final store = InMemoryAuthStore();
    final repo = _repo(store, (options) async {
      if (options.path == '/api/user/me') {
        return _json(
          {
            'id': 7,
            'username': 'martyn',
            'role': 'dm',
          },
          200,
        );
      }
      if (options.path == '/api/tokens') {
        return _json({'token': 'vlt_mobile'}, 201);
      }
      return _json({'error': 'unexpected'}, 500);
    });

    final user = await repo.loginWithSession(
      serverUrl: 'https://vin.dev',
      sessionId: 'oidc-session',
    );

    expect(user.id, 7);
    expect(user.username, 'martyn');
    expect(user.isDm, isTrue);

    final stored = await store.read();
    expect(stored, isNotNull);
    expect(stored!.serverUrl, 'https://vin.dev');
    expect(stored.sessionId, 'oidc-session');
    expect(stored.apiToken, 'vlt_mobile');
  });

  test('loginWithSession clears the store when the session is rejected', () async {
    final store = InMemoryAuthStore();
    final repo = _repo(
      store,
      (options) async => _json({'error': 'unauthenticated'}, 401),
    );

    await expectLater(
      repo.loginWithSession(
        serverUrl: 'https://vin.dev',
        sessionId: 'stale',
      ),
      throwsA(
        isA<ApiException>().having((e) => e.statusCode, 'statusCode', 401),
      ),
    );
    expect(await store.read(), isNull);
  });

  test('oidcEnabled reports false when the endpoint is missing', () async {
    final store = InMemoryAuthStore();
    final repo = _repo(
      store,
      (options) async => _json({'error': 'not found'}, 404),
    );

    expect(await repo.oidcEnabled('https://vin.dev'), isFalse);
  });

  test('oidcEnabled reports true when the server advertises OIDC', () async {
    final store = InMemoryAuthStore();
    final repo = _repo(
      store,
      (options) async => _json({'enabled': true}, 200),
    );

    expect(await repo.oidcEnabled('https://vin.dev'), isTrue);
  });

  test('an unconfigured ApiClient refuses to issue requests', () async {
    final store = InMemoryAuthStore();
    final adapter = _Adapter(
      (options) async => _json({'ok': true}, 200),
    );
    final client = ApiClient(
      baseUrl: '',
      store: store,
      dio: Dio(BaseOptions(validateStatus: (_) => true))
        ..httpClientAdapter = adapter,
    );

    await expectLater(
      client.get('/api/user/me'),
      throwsA(
        isA<ApiException>().having(
          (e) => e.statusCode,
          'statusCode',
          0,
        ),
      ),
    );
    expect(adapter.requests, isEmpty);
  });
}
