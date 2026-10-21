import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:web_socket_channel/io.dart';

import 'auth_store.dart';

class WsEvent {
  final String type;
  final Map<String, dynamic> payload;
  const WsEvent({required this.type, required this.payload});
}

typedef WsChannelFactory = IOWebSocketChannel Function(
  Uri uri, {
  Map<String, dynamic>? headers,
});

class WsClient {
  final String baseUrl;
  final AuthStore store;
  final WsChannelFactory channelFactory;
  final StreamController<WsEvent> _controller =
      StreamController<WsEvent>.broadcast();
  IOWebSocketChannel? _channel;
  Timer? _reconnectTimer;
  bool _disposed = false;
  int _attempt = 0;

  WsClient({
    required this.baseUrl,
    required this.store,
    WsChannelFactory? channelFactory,
  }) : channelFactory =
           channelFactory ??
           ((uri, {headers}) => IOWebSocketChannel.connect(
             uri,
             headers: headers as Map<String, String>?,
           ));

  Stream<WsEvent> get events => _controller.stream;

  void connect() {
    if (_disposed) return;
    _connectOnce();
  }

  Future<void> _connectOnce() async {
    if (_disposed) return;
    // No server configured yet (first run): nothing to connect to.
    if (baseUrl.isEmpty) return;
    final auth = await store.read();
    if (auth == null) {
      _scheduleReconnect();
      return;
    }
    final wsUrl = _toWsUrl(baseUrl);
    try {
      final ch = channelFactory(
        Uri.parse(wsUrl),
        headers: {'Cookie': 'session=${auth.sessionId}'},
      );
      _channel = ch;
      _attempt = 0;
      ch.stream.listen(
        (msg) {
          try {
            final decoded = jsonDecode(msg as String);
            if (decoded is Map) {
              final type = decoded['type']?.toString() ?? '';
              final payload = decoded['payload'] is Map
                  ? Map<String, dynamic>.from(decoded['payload'] as Map)
                  : <String, dynamic>{};
              _controller.add(WsEvent(type: type, payload: payload));
            }
          } catch (_) {}
        },
        onDone: () {
          if (!_disposed) {
            _scheduleReconnect();
          }
        },
        onError: (_) {
          if (!_disposed) {
            _scheduleReconnect();
          }
        },
      );
    } catch (_) {
      _scheduleReconnect();
    }
  }

  String _toWsUrl(String base) {
    var u = base;
    if (u.startsWith('https://')) {
      u = u.replaceFirst('https://', 'wss://');
    } else if (u.startsWith('http://')) {
      u = u.replaceFirst('http://', 'ws://');
    } else if (!u.startsWith('ws://') && !u.startsWith('wss://')) {
      u = "ws://$u";
    }
    if (u.endsWith('/')) {
      u = u.substring(0, u.length - 1);
    }
    return '$u/api/ws';
  }

  void _scheduleReconnect() {
    if (_disposed) return;
    _channel = null;
    _attempt++;
    final baseDelay = min(30000, 1000 * pow(2, _attempt - 1).toInt());
    final jitter = Random().nextInt(500);
    final delay = Duration(milliseconds: baseDelay + jitter);
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(delay, () {
      _connectOnce();
    });
  }

  Duration nextBackoffForTest(int attempt) {
    final baseDelay = min(30000, 1000 * pow(2, attempt - 1).toInt());
    return Duration(milliseconds: baseDelay);
  }

  Future<void> dispose() async {
    _disposed = true;
    _reconnectTimer?.cancel();
    try {
      await _channel?.sink.close();
    } catch (_) {}
    await _controller.close();
  }
}
