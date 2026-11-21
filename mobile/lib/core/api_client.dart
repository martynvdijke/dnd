import 'dart:convert';

import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';

import 'auth_store.dart';

String? parseCookieHeader(String? header, String name) {
  if (header == null || header.isEmpty) return null;
  for (final part in header.split(';')) {
    final pair = part.trim();
    final equals = pair.indexOf('=');
    if (equals <= 0) continue;
    if (pair.substring(0, equals).trim() != name) continue;
    final value = pair.substring(equals + 1).trim();
    return value.isEmpty ? null : value;
  }
  return null;
}

class ApiException implements Exception {
  final int statusCode;
  final String message;
  ApiException(this.statusCode, this.message);
  @override
  String toString() => 'ApiException($statusCode): $message';
}

String deriveCsrfToken(String sessionId) {
  final bytes = utf8.encode('$sessionId-csrf');
  return sha256.convert(bytes).toString();
}

class ApiClient {
  final String baseUrl;
  final AuthStore store;
  final Dio dio;
  Future<void>? _tokenFuture;

  ApiClient({required this.baseUrl, required this.store, Dio? dio})
    : dio =
          dio ??
          Dio(BaseOptions(baseUrl: baseUrl, validateStatus: (_) => true)) {
    // Ensure baseUrl is set on dio
    this.dio.options.baseUrl = baseUrl;
    this.dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final auth = await store.read();
          if (auth != null) {
            options.headers['Cookie'] = 'session=${auth.sessionId}';
            final method = options.method.toUpperCase();
            if (method != 'GET' && method != 'HEAD' && method != 'OPTIONS') {
              options.headers['X-CSRF-Token'] = deriveCsrfToken(auth.sessionId);
              if (auth.apiToken != null && auth.apiToken!.isNotEmpty) {
                options.headers['Authorization'] = 'Bearer ${auth.apiToken}';
              }
            }
          }
          handler.next(options);
        },
      ),
    );
  }

  Future<void> ensureApiToken() {
    if (_tokenFuture != null) return _tokenFuture!;
    _tokenFuture = _ensureApiTokenInternal().whenComplete(() {
      _tokenFuture = null;
    });
    return _tokenFuture!;
  }

  Future<void> _ensureApiTokenInternal() async {
    final auth = await store.read();
    if (auth == null) return;
    if (auth.apiToken != null && auth.apiToken!.isNotEmpty) return;
    // A single retry: token provisioning is the gate for every mutation, so a
    // transient failure must not leave the session token-less.
    Response? last;
    for (var attempt = 0; attempt < 2; attempt++) {
      last = await dio.post(
        '/api/tokens',
        data: {'name': 'mobile', 'expires_in_days': 0},
      );
      final code = last.statusCode ?? 0;
      if (code >= 200 && code < 300) {
        final data = last.data;
        final token = data is Map ? data['token'] as String? : null;
        if (token != null && token.isNotEmpty) {
          await store.write(auth.copyWith(apiToken: token));
          return;
        }
        throw ApiException(
          code,
          'The server accepted the token request but returned no token. '
          'Sign in again.',
        );
      }
    }
    final code = last?.statusCode ?? 0;
    throw ApiException(code, _tokenFailureMessage(last, code));
  }

  /// Turns a failed `POST /api/tokens` into something the user can act on.
  String _tokenFailureMessage(Response? r, int code) {
    String detail = '';
    final data = r?.data;
    if (data is Map && data['error'] != null) {
      detail = data['error'].toString();
    } else if (data is String && data.isNotEmpty) {
      detail = data;
    }
    final status = code == 0 ? 'no response' : 'HTTP $code';
    final suffix = detail.isEmpty ? '' : ' — $detail';
    return 'Could not provision an API token ($status). Mutations will fail '
        'until this succeeds; check that you are signed in and retry.$suffix';
  }

  /// Fails fast when the user has not chosen a server yet instead of sending
  /// requests to some guessed address.
  void _ensureConfigured() {
    if (baseUrl.isEmpty) {
      throw ApiException(
        0,
        'No Villum server configured yet. Enter your server address.',
      );
    }
  }

  Never _throwIfError(Response r) {
    final code = r.statusCode ?? 0;
    if (code >= 200 && code < 300) throw StateError('not error');
    String msg = 'Request failed';
    final data = r.data;
    if (data is Map && data['error'] != null) {
      msg = data['error'].toString();
    } else if (data is String && data.isNotEmpty) {
      try {
        final m = jsonDecode(data);
        if (m is Map && m['error'] != null) msg = m['error'].toString();
      } catch (_) {
        msg = data;
      }
    }
    throw ApiException(code, msg);
  }

  Future<Response> get(
    String path, {
    Map<String, dynamic>? queryParameters,
  }) async {
    _ensureConfigured();
    final r = await dio.get(path, queryParameters: queryParameters);
    if (r.statusCode != null && r.statusCode! >= 200 && r.statusCode! < 300) {
      return r;
    }
    _throwIfError(r);
  }

  Future<Response> post(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
  }) async {
    _ensureConfigured();
    final r = await dio.post(
      path,
      data: data,
      queryParameters: queryParameters,
    );
    if (r.statusCode != null && r.statusCode! >= 200 && r.statusCode! < 300) {
      return r;
    }
    _throwIfError(r);
  }

  Future<Response> put(String path, {dynamic data}) async {
    _ensureConfigured();
    final r = await dio.put(path, data: data);
    if (r.statusCode != null && r.statusCode! >= 200 && r.statusCode! < 300) {
      return r;
    }
    _throwIfError(r);
  }

  Future<Response> patch(String path, {dynamic data}) async {
    _ensureConfigured();
    final r = await dio.patch(path, data: data);
    if (r.statusCode != null && r.statusCode! >= 200 && r.statusCode! < 300) {
      return r;
    }
    _throwIfError(r);
  }

  Future<Response> delete(String path, {dynamic data}) async {
    _ensureConfigured();
    final r = await dio.delete(path, data: data);
    if (r.statusCode != null && r.statusCode! >= 200 && r.statusCode! < 300) {
      return r;
    }
    _throwIfError(r);
  }
}
