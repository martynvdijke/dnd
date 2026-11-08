import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app/router.dart';
import 'app/theme.dart';
import 'core/auth_repository.dart';
import 'core/config.dart';
import 'core/providers.dart';
import 'features/ink/ink_queue.dart';
import 'src/telemetry/telemetry.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const ProviderScope(child: TelemetryBootstrap(child: VillumApp())));
}

/// Starts the OTLP exporter once per resolved server URL, without blocking the
/// first frame and without ever surfacing network errors.
class TelemetryBootstrap extends ConsumerStatefulWidget {
  const TelemetryBootstrap({super.key, required this.child});

  final Widget child;

  @override
  ConsumerState<TelemetryBootstrap> createState() => _TelemetryBootstrapState();
}

class _TelemetryBootstrapState extends ConsumerState<TelemetryBootstrap> {
  String? _startedUrl;

  @override
  void initState() {
    super.initState();
    ref.listenManual<String>(
      serverUrlProvider,
      (_, url) => _maybeStart(url),
      fireImmediately: true,
    );
  }

  void _maybeStart(String url) {
    final resolved = url.trim().isEmpty ? defaultServerUrl : url.trim();
    if (_startedUrl == resolved) return;
    _startedUrl = resolved;
    unawaited(Telemetry.instance.startFromServer(resolved));
  }

  @override
  Widget build(BuildContext context) => widget.child;
}

class VillumApp extends ConsumerStatefulWidget {
  const VillumApp({super.key});

  @override
  ConsumerState<VillumApp> createState() => _VillumAppState();
}

class _VillumAppState extends ConsumerState<VillumApp>
    with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    WidgetsBinding.instance.addPostFrameCallback((_) => _drainSharedPages());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) _drainSharedPages();
  }

  /// Images shared into Villum are copied to app storage by MainActivity. Move
  /// them into the offline queue and open the ink screen once signed in.
  Future<void> _drainSharedPages() async {
    if (ref.read(authControllerProvider) is! AuthAuthenticated) return;
    try {
      final paths = await const InkChannel().takeSharedPaths();
      if (paths.isEmpty) return;
      await InkQueue(ChannelInkQueueStore()).addAll(paths);
      ref.read(goRouterProvider).go('/ink');
    } catch (_) {
      // No platform channel (tests/desktop) or no shares: nothing to drain.
    }
  }

  @override
  Widget build(BuildContext context) {
    final router = ref.watch(goRouterProvider);
    return MaterialApp.router(
      title: 'Villum',
      theme: VillumTheme.dark,
      darkTheme: VillumTheme.dark,
      themeMode: ThemeMode.dark,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
