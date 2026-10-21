import 'api_client.dart';
import 'auth_store.dart';

class UserSession {
  final int id;
  final String username;
  final String role;
  const UserSession({
    required this.id,
    required this.username,
    required this.role,
  });
  bool get isDm => role == 'dm' || role == 'admin';
  @override
  bool operator ==(Object other) =>
      other is UserSession &&
      other.id == id &&
      other.username == username &&
      other.role == role;
  @override
  int get hashCode => Object.hash(id, username, role);
}

sealed class AuthState {
  const AuthState();
}

class AuthUnauthenticated extends AuthState {
  const AuthUnauthenticated();
}

class AuthNeedsSetup extends AuthState {
  final String serverUrl;
  const AuthNeedsSetup(this.serverUrl);
}

class AuthAuthenticated extends AuthState {
  final UserSession user;
  final String serverUrl;
  const AuthAuthenticated(this.user, this.serverUrl);
}

class AuthRepository {
  final AuthStore store;
  final ApiClient Function(String baseUrl) clientFactory;

  AuthRepository({required this.store, required this.clientFactory});

  Future<bool> checkSetup(String serverUrl) async {
    final client = clientFactory(serverUrl);
    final resp = await client.dio.get('/api/check-setup');
    if (resp.statusCode == 200) {
      final data = resp.data;
      if (data is Map) return data['setup'] == true;
      return true;
    }
    throw ApiException(resp.statusCode ?? 0, 'checkSetup failed');
  }

  /// Whether the server offers OIDC login. A failure means "unknown", which is
  /// treated as disabled so the button never dangles.
  Future<bool> oidcEnabled(String serverUrl) async {
    try {
      final client = clientFactory(serverUrl);
      final resp = await client.dio.get('/api/auth/oidc/status');
      if (resp.statusCode == 200 && resp.data is Map) {
        return resp.data['enabled'] == true;
      }
    } catch (_) {
      // Ignore and report disabled.
    }
    return false;
  }

  /// Completes an OIDC login that was performed in an in-app WebView.
  ///
  /// [sessionId] is the value of the `HttpOnly` `session` cookie captured from
  /// the native cookie store. The server-side session already exists, so this
  /// only persists it and resolves the authenticated user from `/api/user/me`.
  Future<UserSession> loginWithSession({
    required String serverUrl,
    required String sessionId,
  }) async {
    await store.write(
      StoredAuth(
        serverUrl: serverUrl,
        sessionId: sessionId,
        apiToken: null,
        userId: 0,
        username: '',
        role: '',
      ),
    );
    final client = clientFactory(serverUrl);
    final resp = await client.dio.get('/api/user/me');
    if (resp.statusCode == null ||
        resp.statusCode! < 200 ||
        resp.statusCode! >= 300) {
      await store.clear();
      String msg = 'OIDC login failed';
      final data = resp.data;
      if (data is Map && data['error'] != null) msg = data['error'].toString();
      throw ApiException(resp.statusCode ?? 0, msg);
    }
    final data = resp.data as Map;
    final userMap = data['user'] is Map ? data['user'] as Map : data;
    final user = UserSession(
      id: (userMap['id'] as num).toInt(),
      username: userMap['username'] as String,
      role: userMap['role'] as String,
    );
    await store.write(
      StoredAuth(
        serverUrl: serverUrl,
        sessionId: sessionId,
        apiToken: null,
        userId: user.id,
        username: user.username,
        role: user.role,
      ),
    );
    // Ensure the bearer token the rest of the app uses.
    final authedClient = clientFactory(serverUrl);
    await authedClient.ensureApiToken();
    return user;
  }


  Future<UserSession> login({
    required String serverUrl,
    required String username,
    required String password,
    bool setup = false,
  }) async {
    final client = clientFactory(serverUrl);
    final resp = await client.dio.post(
      '/api/login',
      data: {
        'username': username,
        'password': password,
        if (setup) 'setup': true,
      },
    );
    if (resp.statusCode == null ||
        resp.statusCode! < 200 ||
        resp.statusCode! >= 300) {
      String msg = 'Login failed';
      final data = resp.data;
      if (data is Map && data['error'] != null) msg = data['error'].toString();
      throw ApiException(resp.statusCode ?? 0, msg);
    }
    final data = resp.data as Map;
    final session = data['session'] as String;
    final userMap = data['user'] as Map;
    final user = UserSession(
      id: userMap['id'] as int,
      username: userMap['username'] as String,
      role: userMap['role'] as String,
    );
    // persist
    await store.write(
      StoredAuth(
        serverUrl: serverUrl,
        sessionId: session,
        userId: user.id,
        username: user.username,
        role: user.role,
      ),
    );
    // ensure token using a client that reads new store
    final authedClient = clientFactory(serverUrl);
    await authedClient.ensureApiToken();
    return user;
  }

  Future<AuthState> restore() async {
    final stored = await store.read();
    if (stored == null) {
      return const AuthUnauthenticated();
    }
    if (stored.serverUrl.isEmpty) {
      return const AuthUnauthenticated();
    }
    // Validate session: GET /api/user/me or check-setup? Use /api/user/me if available, fallback to csrf-token check
    final client = clientFactory(stored.serverUrl);
    try {
      final resp = await client.dio.get('/api/user/me');
      if (resp.statusCode == 200) {
        final data = resp.data;
        UserSession user;
        if (data is Map && data['user'] != null) {
          final u = data['user'] as Map;
          user = UserSession(
            id: u['id'] as int,
            username: u['username'] as String,
            role: u['role'] as String,
          );
        } else if (data is Map && data['id'] != null) {
          user = UserSession(
            id: data['id'] as int,
            username: data['username'] as String,
            role: data['role'] as String,
          );
        } else {
          user = UserSession(
            id: stored.userId,
            username: stored.username,
            role: stored.role,
          );
        }
        return AuthAuthenticated(user, stored.serverUrl);
      } else if (resp.statusCode == 401) {
        return const AuthUnauthenticated();
      } else {
        // treat as unauthenticated
        return const AuthUnauthenticated();
      }
    } catch (_) {
      return const AuthUnauthenticated();
    }
  }

  Future<void> logout() async {
    final stored = await store.read();
    if (stored != null) {
      try {
        final client = clientFactory(stored.serverUrl);
        await client.dio.post('/api/logout');
      } catch (_) {}
    }
    await store.clear();
  }
}
