import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/api_client.dart';
import '../../core/auth_repository.dart';
import '../../core/providers.dart';
import '../../core/server_url.dart';
import 'oidc_webview_screen.dart';

/// Opens the identity provider and resolves to the `session` cookie value, or
/// `null` when the user backed out. Overridable so widget tests can exercise
/// the flow without a platform WebView.
typedef SsoSessionLauncher =
    Future<String?> Function(BuildContext context, String serverUrl);

Future<String?> _pushOidcWebView(BuildContext context, String serverUrl) =>
    Navigator.of(context).push<String>(
      MaterialPageRoute(builder: (_) => OidcWebViewScreen(serverUrl: serverUrl)),
    );

class OnboardingScreen extends ConsumerStatefulWidget {
  final bool forceSetup;
  final SsoSessionLauncher? ssoLauncher;
  const OnboardingScreen({
    super.key,
    this.forceSetup = false,
    this.ssoLauncher,
  });

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  final _urlCtrl = TextEditingController();
  final _userCtrl = TextEditingController();
  final _passCtrl = TextEditingController();
  bool _obscure = true;
  bool _checking = false;
  bool _loggingIn = false;
  bool _ssoBusy = false;
  bool? _needsSetup; // null = unknown, false = login, true = first admin
  bool _oidcEnabled = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    if (widget.forceSetup) _needsSetup = true;
    // Try to preload last URL from store if available
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      try {
        final stored = await ref.read(authStoreProvider).read();
        if (stored != null && stored.serverUrl.isNotEmpty && mounted) {
          setState(() => _urlCtrl.text = stored.serverUrl);
          // A known-good server: show SSO without waiting for "Test connection".
          _checkOidc(stored.serverUrl);
        }
      } catch (_) {}
    });
  }

  @override
  void dispose() {
    _urlCtrl.dispose();
    _userCtrl.dispose();
    _passCtrl.dispose();
    super.dispose();
  }

  Future<void> _testConnection() async {
    final invalid = serverUrlError(_urlCtrl.text);
    if (invalid != null) {
      setState(() => _error = invalid);
      return;
    }
    final url = normalizeServerUrl(_urlCtrl.text);
    if (url != _urlCtrl.text) setState(() => _urlCtrl.text = url);
    setState(() {
      _checking = true;
      _error = null;
    });
    try {
      final repo = ref.read(authRepositoryProvider);
      final setup = await repo.checkSetup(url);
      if (!mounted) return;
      // setup==true => already set up => login mode
      // setup==false => needs first admin
      setState(() {
        _needsSetup = !setup;
      });
      _checkOidc(url);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              _needsSetup!
                  ? 'No admin found — create the first account.'
                  : 'Server reachable — please log in.',
            ),
          ),
        );
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = _friendlyError(e));
    } finally {
      if (mounted) setState(() => _checking = false);
    }
  }

  Future<void> _checkOidc(String url) async {
    try {
      final enabled = await ref.read(authRepositoryProvider).oidcEnabled(url);
      if (mounted) setState(() => _oidcEnabled = enabled);
    } catch (_) {
      if (mounted) setState(() => _oidcEnabled = false);
    }
  }

  Future<void> _oidcLogin() async {
    final invalid = serverUrlError(_urlCtrl.text);
    if (invalid != null) {
      setState(() => _error = invalid);
      return;
    }
    final url = normalizeServerUrl(_urlCtrl.text);
    setState(() {
      _ssoBusy = true;
      _error = null;
    });
    try {
      final launcher = widget.ssoLauncher ?? _pushOidcWebView;
      final session = await launcher(context, url);
      if (session == null || session.isEmpty) return;
      await ref.read(authControllerProvider.notifier).loginWithSession(
            serverUrl: url,
            sessionId: session,
          );
      if (!mounted) return;
      HapticFeedback.mediumImpact();
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = _friendlyError(e));
    } finally {
      if (mounted) setState(() => _ssoBusy = false);
    }
  }

  Future<void> _login() async {
    final invalid = serverUrlError(_urlCtrl.text);
    if (invalid != null) {
      setState(() => _error = invalid);
      return;
    }
    final url = normalizeServerUrl(_urlCtrl.text);
    if (url != _urlCtrl.text) setState(() => _urlCtrl.text = url);
    final user = _userCtrl.text.trim();
    final pass = _passCtrl.text;
    if (user.isEmpty) {
      setState(() => _error = 'Username is required.');
      return;
    }
    if (_needsSetup == true && pass.length < 8) {
      setState(
        () => _error =
            'Password must be at least 8 characters for the first admin.',
      );
      return;
    }
    if (pass.isEmpty) {
      setState(() => _error = 'Password is required.');
      return;
    }
    setState(() {
      _loggingIn = true;
      _error = null;
    });
    try {
      final repo = ref.read(authRepositoryProvider);
      final isSetup = _needsSetup == true;
      await repo.login(
        serverUrl: url,
        username: user,
        password: pass,
        setup: isSetup,
      );
      // Drive controller state so router redirects
      final ctrl = ref.read(authControllerProvider.notifier);
      // If controller has login method, call it; otherwise manually set state
      try {
        await (ctrl as dynamic).login(
          serverUrl: url,
          username: user,
          password: pass,
          setup: isSetup,
        );
      } catch (_) {
        // fallback: try restore
        try {
          await (ctrl as dynamic).restore();
        } catch (_) {}
      }
      if (!mounted) return;
      HapticFeedback.mediumImpact();
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = _friendlyError(e));
      // Password auth is retired by default: point the user at SSO instead of
      // leaving them stuck on a bare 403.
      if (e is ApiException && e.statusCode == 403) {
        await _checkOidc(url);
      }
    } finally {
      if (mounted) setState(() => _loggingIn = false);
    }
  }

  String _friendlyError(Object e) {
    if (e is ApiException) {
      switch (e.statusCode) {
        case 0:
          return 'No Villum server configured yet. Enter your server address.';
        case 401:
          return 'Wrong username or password. Try again.';
        case 403:
          return 'Password sign-in is turned off on this server. Use Sign in with Authelia.';
        case 404:
          return 'That address does not look like a Villum server.';
      }
    }
    final s = e.toString();
    if (s.contains('SocketException') ||
        s.contains('Failed host lookup') ||
        s.contains('Connection refused')) {
      return 'Cannot reach that server. Check the URL and that the Villum server is running.';
    }
    if (s.contains('HandshakeException') ||
        s.contains('CERTIFICATE_VERIFY_FAILED')) {
      return 'The server certificate could not be verified.';
    }
    // Strip the ApiException / DioException decoration for anything else.
    return s
        .replaceFirst(RegExp(r'^ApiException\(\d+\):\s*'), '')
        .replaceFirst(RegExp(r'^DioException \[[^\]]*\]:\s*'), '')
        .trim();
  }

  @override
  Widget build(BuildContext context) {
    final isSetup = _needsSetup == true;
    final busy = _loggingIn || _ssoBusy;
    // OIDC cannot create the first admin, so only offer SSO once setup is done.
    final ssoAvailable = _oidcEnabled && !isSetup;
    final authState = ref.watch(authControllerProvider);
    // Show already-authenticated notice
    final alreadyIn = authState is AuthAuthenticated;

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(20, 24, 20, 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const SizedBox(height: 8),
              // Crest / header
              Center(
                child: Column(
                  children: [
                    Container(
                      width: 72,
                      height: 72,
                      decoration: BoxDecoration(
                        color: VillumTheme.gold.withValues(alpha: 0.14),
                        borderRadius: BorderRadius.circular(18),
                        border: Border.all(
                          color: VillumTheme.gold.withValues(alpha: 0.35),
                        ),
                      ),
                      child: const Icon(
                        Icons.shield_rounded,
                        size: 36,
                        color: VillumTheme.gold,
                      ),
                    ),
                    const SizedBox(height: 14),
                    Text(
                      'VILLUM',
                      style: GoogleFonts.cinzel(
                        fontSize: 28,
                        fontWeight: FontWeight.w900,
                        letterSpacing: 3,
                        color: VillumTheme.parchment,
                      ),
                    ),
                    Text(
                      'Pocket companion for the table',
                      style: GoogleFonts.lora(
                        fontSize: 13,
                        color: VillumTheme.parchmentMuted,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Container(
                      height: 1,
                      width: 120,
                      color: VillumTheme.gold.withValues(alpha: 0.35),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              if (alreadyIn)
                Card(
                  color: VillumTheme.hpHealthy.withValues(alpha: 0.12),
                  child: Padding(
                    padding: const EdgeInsets.all(14),
                    child: Row(
                      children: [
                        const Icon(
                          Icons.check_circle,
                          color: VillumTheme.hpHealthy,
                        ),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Text(
                            // ignore: unnecessary_cast
                            'Signed in as ${alreadyIn ? (authState as AuthAuthenticated).user.username : ''}',
                            style: GoogleFonts.lora(
                              color: VillumTheme.parchment,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              // Server URL
              Text(
                'Server URL',
                style: GoogleFonts.cinzel(
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 0.8,
                  color: VillumTheme.gold,
                ),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _urlCtrl,
                keyboardType: TextInputType.url,
                autocorrect: false,
                decoration: InputDecoration(
                  hintText: 'http://192.168.1.10:6270',
                  prefixIcon: const Icon(Icons.link),
                  suffixIcon: _checking
                      ? const Padding(
                          padding: EdgeInsets.all(12),
                          child: SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        )
                      : null,
                ),
                onSubmitted: (_) => _testConnection(),
              ),
              const SizedBox(height: 10),
              SizedBox(
                height: 48,
                child: OutlinedButton.icon(
                  onPressed: _checking ? null : _testConnection,
                  icon: const Icon(Icons.wifi_tethering, size: 18),
                  label: Text(_checking ? 'Testing…' : 'Test connection'),
                ),
              ),
              if (_needsSetup != null) ...[
                const SizedBox(height: 10),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 10,
                  ),
                  decoration: BoxDecoration(
                    color: isSetup
                        ? VillumTheme.gold.withValues(alpha: 0.12)
                        : VillumTheme.surfaceCard,
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                      color: isSetup
                          ? VillumTheme.gold.withValues(alpha: 0.3)
                          : VillumTheme.inkLight,
                    ),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        isSetup ? Icons.person_add_alt_1 : Icons.login,
                        size: 18,
                        color: VillumTheme.gold,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          isSetup
                              ? 'No admin exists yet — you will create the first admin account.'
                              : 'Server is ready — log in with your Villum account.',
                          style: GoogleFonts.lora(
                            fontSize: 12,
                            color: VillumTheme.parchmentMuted,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: 18),
              const Divider(),
              const SizedBox(height: 14),
              if (ssoAvailable) ...[
                SizedBox(
                  height: 52,
                  child: FilledButton.icon(
                    key: const Key('sso-sign-in'),
                    onPressed: _ssoBusy ? null : _oidcLogin,
                    icon: _ssoBusy
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: VillumTheme.ink,
                            ),
                          )
                        : const Icon(Icons.shield_outlined, size: 20),
                    label: Text(
                      _ssoBusy
                          ? 'Waiting for Authelia…'
                          : 'Sign in with Authelia',
                    ),
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Opens your identity provider in a secure sign-in window.',
                  textAlign: TextAlign.center,
                  style: GoogleFonts.lora(
                    fontSize: 11,
                    color: VillumTheme.parchmentMuted,
                  ),
                ),
                const SizedBox(height: 18),
              ],
              Text(
                ssoAvailable
                    ? 'Password sign-in (break-glass)'
                    : (isSetup ? 'Create admin account' : 'Log in'),
                style: GoogleFonts.cinzel(
                  fontSize: 13,
                  fontWeight: FontWeight.w700,
                  color: VillumTheme.parchment,
                ),
              ),
              const SizedBox(height: 4),
              if (isSetup)
                Text(
                  'There is no self-service registration. This creates the server\'s first admin.',
                  style: GoogleFonts.lora(
                    fontSize: 12,
                    color: VillumTheme.parchmentMuted,
                  ),
                )
              else if (ssoAvailable)
                Text(
                  'Recovery only — most servers have password login turned off.',
                  style: GoogleFonts.lora(
                    fontSize: 12,
                    color: VillumTheme.parchmentMuted,
                  ),
                ),
              const SizedBox(height: 12),
              TextField(
                controller: _userCtrl,
                textInputAction: TextInputAction.next,
                autocorrect: false,
                decoration: const InputDecoration(
                  labelText: 'Username',
                  prefixIcon: Icon(Icons.person_outline),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _passCtrl,
                obscureText: _obscure,
                decoration: InputDecoration(
                  labelText: isSetup
                      ? 'Password (min 8 characters)'
                      : 'Password',
                  prefixIcon: const Icon(Icons.lock_outline),
                  suffixIcon: IconButton(
                    icon: Icon(
                      _obscure ? Icons.visibility_off : Icons.visibility,
                    ),
                    onPressed: () => setState(() => _obscure = !_obscure),
                    tooltip: _obscure ? 'Show password' : 'Hide password',
                  ),
                ),
                onSubmitted: (_) => _login(),
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: VillumTheme.hpDying.withValues(alpha: 0.12),
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                      color: VillumTheme.hpDying.withValues(alpha: 0.35),
                    ),
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Icon(
                        Icons.error_outline,
                        size: 18,
                        color: VillumTheme.hpDying,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _error!,
                          style: GoogleFonts.lora(
                            fontSize: 13,
                            color: VillumTheme.parchment,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: 18),
              SizedBox(
                height: ssoAvailable ? 48 : 52,
                child: ssoAvailable
                    ? OutlinedButton(
                        onPressed: busy ? null : _login,
                        child: _loggingIn
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: VillumTheme.gold,
                                ),
                              )
                            : const Text('Log in with password'),
                      )
                    : FilledButton(
                        onPressed: busy ? null : _login,
                        child: _loggingIn
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: VillumTheme.ink,
                                ),
                              )
                            : Text(isSetup ? 'Create admin & enter' : 'Log in'),
                      ),
              ),
              if (!ssoAvailable && _oidcEnabled) ...[
                const SizedBox(height: 10),
                SizedBox(
                  height: 48,
                  child: OutlinedButton.icon(
                    onPressed: busy ? null : _oidcLogin,
                    icon: const Icon(Icons.shield_outlined, size: 18),
                    label: const Text('Sign in with Authelia'),
                  ),
                ),
              ],
              const SizedBox(height: 12),
              Text(
                'Use the same address you open Villum with in your browser, '
                'for example https://villum.example.com.',
                style: GoogleFonts.lora(
                  fontSize: 11,
                  color: VillumTheme.parchmentMuted.withValues(alpha: 0.8),
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
