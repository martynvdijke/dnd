import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/auth_repository.dart';
import '../core/providers.dart';
import '../features/battlemap/battlemap_screen.dart';
import '../features/character_sheet/character_sheet_screen.dart';
import '../features/combat/combat_screen.dart';
import '../features/compendium/compendium_screen.dart';
import '../features/dm_tools/dm_tools_screen.dart';
import '../features/onboarding/onboarding_screen.dart';

// Refresh listenable that rebuilds GoRouter when auth state changes.
class AuthRefresh extends ChangeNotifier {
  final Ref ref;
  AuthRefresh(this.ref) {
    ref.listen<AuthState>(authControllerProvider, (prev, next) {
      if (prev != next) notifyListeners();
    });
  }
}

final goRouterProvider = Provider<GoRouter>((ref) {
  final refresh = AuthRefresh(ref);

  return GoRouter(
    initialLocation: '/sheet',
    refreshListenable: refresh,
    redirect: (context, state) {
      final auth = ref.read(authControllerProvider);

      final loc = state.matchedLocation;
      final isOnboarding = loc.startsWith('/onboarding');
      final isAuthRoute = isOnboarding;

      if (auth is AuthUnauthenticated) {
        return isAuthRoute ? null : '/onboarding';
      }
      if (auth is AuthNeedsSetup) {
        // Allow onboarding to show setup mode; don't bounce
        return isAuthRoute ? null : '/onboarding?setup=1';
      }
      if (auth is AuthAuthenticated) {
        if (isAuthRoute) return '/sheet';
        return null;
      }
      return null;
    },
    routes: [
      GoRoute(
        path: '/onboarding',
        builder: (ctx, state) {
          final setup =
              state.uri.queryParameters['setup'] == '1' ||
              state.uri.queryParameters['setup'] == 'true';
          return OnboardingScreen(forceSetup: setup);
        },
      ),
      StatefulShellRoute.indexedStack(
        builder: (context, state, shell) => VillumShell(shell: shell),
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/sheet',
                builder: (c, s) {
                  final idStr = s.uri.queryParameters['id'];
                  final id = int.tryParse(idStr ?? '');
                  return CharacterSheetScreen(initialCharacterId: id);
                },
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/combat',
                builder: (c, s) {
                  final cid = int.tryParse(
                    s.uri.queryParameters['campaign'] ?? '',
                  );
                  return CombatScreen(campaignId: cid);
                },
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/map',
                builder: (c, s) {
                  final cid = int.tryParse(
                    s.uri.queryParameters['campaign'] ?? '',
                  );
                  return BattlemapScreen(campaignId: cid);
                },
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/compendium',
                builder: (c, s) => const CompendiumScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/dm',
                builder: (c, s) {
                  final cid = int.tryParse(
                    s.uri.queryParameters['campaign'] ?? '',
                  );
                  return DmToolsScreen(campaignId: cid);
                },
              ),
            ],
          ),
        ],
      ),
    ],
    errorBuilder: (c, s) =>
        Scaffold(body: Center(child: Text('Route not found: ${s.uri}'))),
  );
});

class VillumShell extends StatelessWidget {
  final StatefulNavigationShell shell;
  const VillumShell({super.key, required this.shell});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: shell,
      bottomNavigationBar: NavigationBar(
        selectedIndex: shell.currentIndex,
        onDestinationSelected: shell.goBranch,
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.person_outline),
            selectedIcon: Icon(Icons.person),
            label: 'Sheet',
          ),
          NavigationDestination(
            icon: Icon(Icons.sports_kabaddi_outlined),
            selectedIcon: Icon(Icons.sports_kabaddi),
            label: 'Combat',
          ),
          NavigationDestination(
            icon: Icon(Icons.map_outlined),
            selectedIcon: Icon(Icons.map),
            label: 'Map',
          ),
          NavigationDestination(
            icon: Icon(Icons.menu_book_outlined),
            selectedIcon: Icon(Icons.menu_book),
            label: 'Compendium',
          ),
          NavigationDestination(
            icon: Icon(Icons.shield_outlined),
            selectedIcon: Icon(Icons.shield),
            label: 'DM',
          ),
        ],
      ),
    );
  }
}
