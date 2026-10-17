import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
// ignore: implementation_imports, depend_on_referenced_packages
import 'package:riverpod/src/framework.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:dnd/app/theme.dart';
import 'package:dnd/core/api_client.dart';
import 'package:dnd/core/auth_repository.dart';
import 'package:dnd/core/auth_store.dart';
import 'package:dnd/core/providers.dart';
import 'package:dnd/core/ws_client.dart';

import 'fake_api.dart';

/// Hermeticity: google_fonts attempts runtime fetch which is network-dependent.
/// In tests we disable it and provide a fallback theme that doesn't require bundled font assets.
/// If allowRuntimeFetching=false throws (fonts not bundled), we gracefully keep default fetching off
/// by using a test theme that avoids GoogleFonts calls.
ThemeData testTheme() {
  try {
    GoogleFonts.config.allowRuntimeFetching = false;
  } catch (_) {}
  try {
    return VillumTheme.dark;
  } catch (_) {
    return ThemeData.dark(useMaterial3: true).copyWith(
      colorScheme: const ColorScheme.dark(
        primary: Color(0xFFB8963E),
        onPrimary: Color(0xFF1A1410),
        surface: Color(0xFF211C17),
        onSurface: Color(0xFFF5ECD7),
        error: Color(0xFFE53935),
      ),
      scaffoldBackgroundColor: const Color(0xFF1A1410),
    );
  }
}

void disableGoogleFonts() {
  try {
    GoogleFonts.config.allowRuntimeFetching = false;
  } catch (_) {}
}

/// Creates an ApiClient wired to FakeApi + InMemoryAuthStore
ApiClient makeApiClient(
  FakeApi fake,
  InMemoryAuthStore store, {
  String baseUrl = 'http://localhost:6270',
}) {
  final dio = fake.buildDio(baseUrl: baseUrl);
  return ApiClient(baseUrl: baseUrl, store: store, dio: dio);
}

List<Override> baseOverrides({
  required FakeApi fake,
  required InMemoryAuthStore store,
  StreamController<WsEvent>? wsController,
  String serverUrl = 'http://localhost:6270',
  AuthState? authState,
}) {
  final dio = fake.buildDio(baseUrl: serverUrl);
  final apiClient = ApiClient(baseUrl: serverUrl, store: store, dio: dio);
  final overrides = <Override>[
    authStoreProvider.overrideWithValue(store),
    serverUrlProvider.overrideWithValue(serverUrl),
    apiClientProvider.overrideWithValue(apiClient),
    authRepositoryProvider.overrideWithValue(
      AuthRepository(
        store: store,
        clientFactory: (url) => ApiClient(
          baseUrl: url,
          store: store,
          dio: fake.buildDio(baseUrl: url),
        ),
      ),
    ),
  ];
  if (wsController != null) {
    // Provide a never-connecting WsClient and override wsEventsProvider
    final wsClient = WsClient(baseUrl: serverUrl, store: store);
    overrides.addAll([
      wsClientProvider.overrideWithValue(wsClient),
      wsEventsProvider.overrideWith((ref) => wsController.stream),
    ]);
  } else {
    final emptyCtrl = StreamController<WsEvent>.broadcast();
    overrides.addAll([
      wsEventsProvider.overrideWith((ref) => emptyCtrl.stream),
    ]);
  }
  if (authState != null) {
    overrides.add(
      authControllerProvider.overrideWith(() {
        final c = _FixedAuthController(authState);
        return c;
      }),
    );
  }
  // Provide empty campaigns/characters by default via fake routes; screens read providers directly
  return overrides;
}

class _FixedAuthController extends AuthController {
  final AuthState fixed;
  _FixedAuthController(this.fixed);
  @override
  AuthState build() => fixed;
}

Future<void> pumpScreen(
  WidgetTester tester,
  Widget screen, {
  required FakeApi fake,
  required InMemoryAuthStore store,
  StreamController<WsEvent>? wsController,
  AuthState? authState,
  List<Override> extraOverrides = const [],
}) async {
  disableGoogleFonts();
  final overrides = baseOverrides(
    fake: fake,
    store: store,
    wsController: wsController,
    authState: authState,
  )..addAll(extraOverrides);
  await tester.pumpWidget(
    ProviderScope(
      overrides: overrides,
      child: MaterialApp(theme: testTheme(), home: screen),
    ),
  );
  await tester.pump();
}

Future<void> pumpVillumApp(
  WidgetTester tester, {
  required FakeApi fake,
  required InMemoryAuthStore store,
  StreamController<WsEvent>? wsController,
  AuthState? authState,
  List<Override> extraOverrides = const [],
}) async {
  disableGoogleFonts();
  // Not used for full app router tests currently; pumpScreen is preferred
  await pumpScreen(
    tester,
    const Scaffold(body: Text('app')),
    fake: fake,
    store: store,
    wsController: wsController,
    authState: authState,
    extraOverrides: extraOverrides,
  );
}
