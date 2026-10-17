import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class StoredAuth {
  final String serverUrl;
  final String sessionId;
  final String? apiToken;
  final int userId;
  final String username;
  final String role;
  const StoredAuth({
    required this.serverUrl,
    required this.sessionId,
    this.apiToken,
    required this.userId,
    required this.username,
    required this.role,
  });
  StoredAuth copyWith({
    String? serverUrl,
    String? sessionId,
    String? apiToken,
    int? userId,
    String? username,
    String? role,
  }) {
    return StoredAuth(
      serverUrl: serverUrl ?? this.serverUrl,
      sessionId: sessionId ?? this.sessionId,
      apiToken: apiToken ?? this.apiToken,
      userId: userId ?? this.userId,
      username: username ?? this.username,
      role: role ?? this.role,
    );
  }

  Map<String, dynamic> toJson() => {
    'serverUrl': serverUrl,
    'sessionId': sessionId,
    'apiToken': apiToken,
    'userId': userId,
    'username': username,
    'role': role,
  };
  factory StoredAuth.fromJson(Map<String, dynamic> j) => StoredAuth(
    serverUrl: j['serverUrl'] as String,
    sessionId: j['sessionId'] as String,
    apiToken: j['apiToken'] as String?,
    userId: j['userId'] as int,
    username: j['username'] as String,
    role: j['role'] as String,
  );
  @override
  bool operator ==(Object other) =>
      other is StoredAuth &&
      other.serverUrl == serverUrl &&
      other.sessionId == sessionId &&
      other.apiToken == apiToken &&
      other.userId == userId &&
      other.username == username &&
      other.role == role;
  @override
  int get hashCode =>
      Object.hash(serverUrl, sessionId, apiToken, userId, username, role);
}

abstract class AuthStore {
  Future<StoredAuth?> read();
  Future<void> write(StoredAuth auth);
  Future<void> clear();
}

class SecureAuthStore implements AuthStore {
  final FlutterSecureStorage _storage;
  static const _kUrl = 'villum_server_url';
  static const _kSession = 'villum_session';
  static const _kToken = 'villum_api_token';
  static const _kUserId = 'villum_user_id';
  static const _kUsername = 'villum_username';
  static const _kRole = 'villum_role';

  SecureAuthStore({FlutterSecureStorage? storage})
    : _storage =
          storage ??
          const FlutterSecureStorage(
            aOptions: AndroidOptions(storageNamespace: 'villum'),
          );

  @override
  Future<StoredAuth?> read() async {
    final url = await _storage.read(key: _kUrl);
    final sess = await _storage.read(key: _kSession);
    if (url == null || sess == null) return null;
    final token = await _storage.read(key: _kToken);
    final uidStr = await _storage.read(key: _kUserId);
    final username = await _storage.read(key: _kUsername) ?? '';
    final role = await _storage.read(key: _kRole) ?? '';
    final uid = int.tryParse(uidStr ?? '') ?? 0;
    return StoredAuth(
      serverUrl: url,
      sessionId: sess,
      apiToken: token,
      userId: uid,
      username: username,
      role: role,
    );
  }

  @override
  Future<void> write(StoredAuth auth) async {
    await _storage.write(key: _kUrl, value: auth.serverUrl);
    await _storage.write(key: _kSession, value: auth.sessionId);
    if (auth.apiToken != null) {
      await _storage.write(key: _kToken, value: auth.apiToken);
    }
    await _storage.write(key: _kUserId, value: auth.userId.toString());
    await _storage.write(key: _kUsername, value: auth.username);
    await _storage.write(key: _kRole, value: auth.role);
  }

  @override
  Future<void> clear() async {
    await _storage.delete(key: _kUrl);
    await _storage.delete(key: _kSession);
    await _storage.delete(key: _kToken);
    await _storage.delete(key: _kUserId);
    await _storage.delete(key: _kUsername);
    await _storage.delete(key: _kRole);
  }
}

class InMemoryAuthStore implements AuthStore {
  StoredAuth? _value;
  @override
  Future<StoredAuth?> read() async => _value;
  @override
  Future<void> write(StoredAuth auth) async {
    _value = auth;
  }

  @override
  Future<void> clear() async {
    _value = null;
  }
}
