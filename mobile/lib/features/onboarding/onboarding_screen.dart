import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/auth_repository.dart';
import '../../core/providers.dart';

class OnboardingScreen extends ConsumerStatefulWidget {
  final bool forceSetup;
  const OnboardingScreen({super.key, this.forceSetup = false});

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  final _urlCtrl = TextEditingController(text: 'http://10.0.2.2:6270');
  final _userCtrl = TextEditingController();
  final _passCtrl = TextEditingController();
  bool _obscure = true;
  bool _checking = false;
  bool _loggingIn = false;
  bool? _needsSetup; // null = unknown, false = login, true = first admin
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
    final url = _urlCtrl.text.trim();
    if (url.isEmpty ||
        !(url.startsWith('http://') || url.startsWith('https://'))) {
      setState(
        () => _error = 'Enter a valid URL starting with http:// or https://',
      );
      return;
    }
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

  Future<void> _login() async {
    final url = _urlCtrl.text.trim();
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
    } finally {
      if (mounted) setState(() => _loggingIn = false);
    }
  }

  String _friendlyError(Object e) {
    final s = e.toString();
    if (s.contains('401') ||
        s.toLowerCase().contains('unauthorized') ||
        s.toLowerCase().contains('invalid')) {
      return 'Wrong username or password. Try again.';
    }
    if (s.contains('SocketException') ||
        s.contains('Failed host lookup') ||
        s.contains('Connection refused')) {
      return 'Cannot reach that server. Check the URL and that the Villum server is running.';
    }
    // Strip ApiException prefix
    return s
        .replaceFirst('ApiException', '')
        .replaceFirst(RegExp(r'\(\d+\):'), '')
        .trim();
  }

  @override
  Widget build(BuildContext context) {
    final isSetup = _needsSetup == true;
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
              Text(
                isSetup ? 'Create admin account' : 'Log in',
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
                height: 52,
                child: FilledButton(
                  onPressed: _loggingIn ? null : _login,
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
              const SizedBox(height: 10),
              Text(
                'Tip: on Android emulator use 10.0.2.2; on a real device use your PC\'s LAN IP. Default port is 6270 (Docker: 6280).',
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
