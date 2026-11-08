import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:dnd/src/telemetry/telemetry.dart';

class _FakeAdapter implements HttpClientAdapter {
  _FakeAdapter(this.onFetch);

  final Future<ResponseBody> Function(RequestOptions options) onFetch;
  final List<RequestOptions> requests = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) {
    requests.add(options);
    return onFetch(options);
  }

  @override
  void close({bool force = false}) {}
}

ResponseBody _jsonBody(Object body, int status) => ResponseBody.fromString(
  jsonEncode(body),
  status,
  headers: {
    Headers.contentTypeHeader: [Headers.jsonContentType],
  },
);

Dio _dioWith(HttpClientAdapter adapter) {
  final dio = Dio(BaseOptions(validateStatus: (_) => true));
  dio.httpClientAdapter = adapter;
  return dio;
}

void main() {
  group('normalizeEndpoint', () {
    test('appends /v1/traces', () {
      expect(
        Telemetry.normalizeEndpoint('http://otel:4318'),
        'http://otel:4318/v1/traces',
      );
    });

    test('strips trailing slashes before appending', () {
      expect(
        Telemetry.normalizeEndpoint('https://otel.example.com///'),
        'https://otel.example.com/v1/traces',
      );
    });

    test('keeps an existing traces path', () {
      expect(
        Telemetry.normalizeEndpoint('https://otel.example.com/v1/traces/'),
        'https://otel.example.com/v1/traces',
      );
    });

    test('empty input stays empty', () {
      expect(Telemetry.normalizeEndpoint('   '), '');
    });
  });

  group('parseOtelHeaders', () {
    test('parses comma separated key=value pairs', () {
      expect(
        Telemetry.parseOtelHeaders(
          'Authorization=Bearer%20abc123,X-Scope=traces',
        ),
        {'Authorization': 'Bearer abc123', 'X-Scope': 'traces'},
      );
    });

    test('ignores malformed entries and empty values', () {
      expect(Telemetry.parseOtelHeaders(' ,novalue,=x,Good=1,Bad='), {
        'Good': '1',
      });
    });

    test('empty string yields empty map', () {
      expect(Telemetry.parseOtelHeaders(''), isEmpty);
    });
  });

  group('buildOtlpPayload', () {
    test('emits resource, scope, span and status shape', () {
      final span = TraceSpan(
        name: 'http.request',
        attributes: {'http.request.method': 'GET'},
        traceId: 'a' * 32,
        spanId: 'b' * 16,
        startNanos: 1000,
      );

      final payload = buildOtlpPayload(
        serviceName: 'dnd-android',
        serviceVersion: '9.9.9',
        span: span,
        attributes: {
          'http.request.method': 'GET',
          'custom.count': 3,
          'custom.ok': true,
        },
        failed: false,
        endNanos: 2000,
      );

      final resourceSpans = payload['resourceSpans'] as List;
      final resource = resourceSpans.single as Map;
      final resourceAttrs = resource['resource']['attributes'] as List;
      final byKey = {
        for (final attr in resourceAttrs)
          (attr as Map)['key'] as String: attr['value'],
      };

      expect(byKey['service.name'], {'stringValue': 'dnd-android'});
      expect(byKey['service.version'], {'stringValue': '9.9.9'});
      expect(byKey['telemetry.sdk.name'], {'stringValue': 'dnd-mobile'});
      expect(byKey['telemetry.sdk.language'], {'stringValue': 'dart'});

      final scopeSpans = resource['scopeSpans'] as List;
      final scope = scopeSpans.single as Map;
      expect(scope['scope'], {'name': 'dnd.mobile', 'version': '9.9.9'});

      final outSpan = (scope['spans'] as List).single as Map;
      expect(outSpan['traceId'], 'a' * 32);
      expect(outSpan['spanId'], 'b' * 16);
      expect(outSpan['name'], 'http.request');
      expect(outSpan['startTimeUnixNano'], '1000');
      expect(outSpan['endTimeUnixNano'], '2000');
      expect(outSpan['status'], {'code': 1});

      final attrs = {
        for (final attr in outSpan['attributes'] as List)
          (attr as Map)['key'] as String: attr['value'],
      };
      expect(attrs['custom.count'], {'intValue': '3'});
      expect(attrs['custom.ok'], {'boolValue': true});
    });

    test('marks failures with status code 2', () {
      final span = TraceSpan(
        name: 'http.request',
        attributes: const {},
        traceId: 'a' * 32,
        spanId: 'b' * 16,
        startNanos: 1,
      );
      final payload = buildOtlpPayload(
        serviceName: 'dnd-android',
        serviceVersion: '1.0.0',
        span: span,
        attributes: const {},
        failed: true,
        endNanos: 2,
      );
      final spanJson =
          ((payload['resourceSpans'] as List).single
                  as Map)['scopeSpans'][0]['spans'][0]
              as Map;
      expect(spanJson['status'], {'code': 2});
    });
  });

  group('OtelSettings.fromJson', () {
    test('reads canonical keys', () {
      final settings = OtelSettings.fromJson({
        'endpoint': 'http://otel:4318',
        'enabled': true,
        'traces_enabled': true,
        'metrics_enabled': false,
        'logs_enabled': false,
      });
      expect(settings.endpoint, 'http://otel:4318');
      expect(settings.enabled, isTrue);
      expect(settings.tracesEnabled, isTrue);
      expect(settings.metricsEnabled, isFalse);
      expect(settings.logsEnabled, isFalse);
    });

    test('accepts otel_ prefixed keys and string booleans', () {
      final settings = OtelSettings.fromJson({
        'otel_endpoint': ' https://otel.example.com ',
        'otel_enabled': 'true',
        'traces_enabled': '0',
      });
      expect(settings.endpoint, 'https://otel.example.com');
      expect(settings.enabled, isTrue);
      expect(settings.tracesEnabled, isFalse);
    });

    test('traces defaults to enabled', () {
      final settings = OtelSettings.fromJson({
        'endpoint': 'http://otel:4318',
        'enabled': true,
      });
      expect(settings.tracesEnabled, isTrue);
      expect(settings.metricsEnabled, isFalse);
      expect(settings.logsEnabled, isFalse);
    });

    test('empty endpoint disables everything', () {
      final settings = OtelSettings.fromJson({
        'endpoint': '',
        'enabled': true,
        'traces_enabled': true,
        'metrics_enabled': true,
        'logs_enabled': true,
      });
      expect(settings.enabled, isFalse);
      expect(settings.tracesEnabled, isFalse);
      expect(settings.metricsEnabled, isFalse);
      expect(settings.logsEnabled, isFalse);
    });

    test('missing endpoint is treated as disabled', () {
      final settings = OtelSettings.fromJson({});
      expect(settings.endpoint, '');
      expect(settings.enabled, isFalse);
      expect(settings.tracesEnabled, isFalse);
    });
  });

  group('fetchOtelSettings', () {
    test('returns parsed settings on 200', () async {
      final adapter = _FakeAdapter(
        (_) async => _jsonBody({
          'endpoint': 'http://otel:4318',
          'enabled': true,
          'traces_enabled': true,
          'metrics_enabled': false,
          'logs_enabled': false,
        }, 200),
      );

      final settings = await fetchOtelSettings(
        'https://dnd.example.com/',
        dio: _dioWith(adapter),
      );

      expect(settings, isNotNull);
      expect(settings!.endpoint, 'http://otel:4318');
      expect(settings.tracesEnabled, isTrue);
      expect(
        adapter.requests.single.path,
        'https://dnd.example.com/api/otel-settings',
      );
    });

    test('returns null on non-200', () async {
      final adapter = _FakeAdapter(
        (_) async => _jsonBody({'error': 'nope'}, 503),
      );
      final settings = await fetchOtelSettings(
        'https://dnd.example.com',
        dio: _dioWith(adapter),
      );
      expect(settings, isNull);
    });

    test('returns null when the request throws', () async {
      final adapter = _FakeAdapter(
        (options) async =>
            throw DioException(requestOptions: options, message: 'offline'),
      );
      final settings = await fetchOtelSettings(
        'https://dnd.example.com',
        dio: _dioWith(adapter),
      );
      expect(settings, isNull);
    });

    test('returns null on timeout', () async {
      final adapter = _FakeAdapter((_) async {
        await Future<void>.delayed(const Duration(milliseconds: 200));
        return _jsonBody({'endpoint': 'http://late:4318'}, 200);
      });
      final settings = await fetchOtelSettings(
        'https://dnd.example.com',
        dio: _dioWith(adapter),
        timeout: const Duration(milliseconds: 20),
      );
      expect(settings, isNull);
    });

    test('returns null on an empty base URL without a request', () async {
      final adapter = _FakeAdapter((_) async => _jsonBody({}, 200));
      final settings = await fetchOtelSettings('', dio: _dioWith(adapter));
      expect(settings, isNull);
      expect(adapter.requests, isEmpty);
    });
  });
}
