import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'auth_store.dart';
import 'api_client.dart';
import 'auth_repository.dart';
import 'ws_client.dart';
import '../data/models/character_summary.dart';
import '../data/models/campaign.dart';
import '../data/api/character_api.dart';
import '../data/api/campaign_api.dart';
export '../data/controllers/character_controller.dart' show characterProvider;
export '../data/controllers/combat_controller.dart' show combatProvider;
export '../data/controllers/battlemap_controller.dart' show battlemapProvider;
export '../data/controllers/dm_screen_controller.dart' show dmScreenProvider;
export '../data/models/battlemap.dart'
    show BattlemapState, BattlemapToken, BattlemapMeta;
export '../data/models/compendium.dart' show CompendiumSearchResult;
export '../data/models/dm_screen.dart' show DmScreen;
export '../data/models/draft.dart' show DraftSession;

final authStoreProvider = Provider<AuthStore>((ref) => SecureAuthStore());
final serverUrlProvider = Provider<String>((ref) => '');
final apiClientProvider = Provider<ApiClient>((ref) {
  final store = ref.watch(authStoreProvider);
  final url = ref.watch(serverUrlProvider);
  return ApiClient(
    baseUrl: url.isEmpty ? 'http://localhost:6270' : url,
    store: store,
  );
});
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final store = ref.watch(authStoreProvider);
  return AuthRepository(
    store: store,
    clientFactory: (url) => ApiClient(baseUrl: url, store: store),
  );
});

class AuthController extends Notifier<AuthState> {
  @override
  AuthState build() => const AuthUnauthenticated();
  Future<void> checkSetup(String serverUrl) async {
    final repo = ref.read(authRepositoryProvider);
    final ok = await repo.checkSetup(serverUrl);
    if (!ref.mounted) return;
    if (!ok) {
      state = AuthNeedsSetup(serverUrl);
    } else {
      state = const AuthUnauthenticated();
    }
  }

  Future<void> login({
    required String serverUrl,
    required String username,
    required String password,
    bool setup = false,
  }) async {
    final repo = ref.read(authRepositoryProvider);
    final user = await repo.login(
      serverUrl: serverUrl,
      username: username,
      password: password,
      setup: setup,
    );
    if (!ref.mounted) return;
    state = AuthAuthenticated(user, serverUrl);
  }

  Future<void> restore() async {
    final repo = ref.read(authRepositoryProvider);
    final s = await repo.restore();
    if (!ref.mounted) return;
    state = s;
  }

  Future<void> logout() async {
    await ref.read(authRepositoryProvider).logout();
    if (!ref.mounted) return;
    state = const AuthUnauthenticated();
  }
}

final authControllerProvider = NotifierProvider<AuthController, AuthState>(
  AuthController.new,
);

final wsClientProvider = Provider<WsClient>((ref) {
  final store = ref.watch(authStoreProvider);
  final url = ref.watch(serverUrlProvider);
  final c = WsClient(
    baseUrl: url.isEmpty ? 'http://localhost:6270' : url,
    store: store,
  );
  ref.onDispose(() => c.dispose());
  return c;
});

final wsEventsProvider = StreamProvider<WsEvent>((ref) {
  final ws = ref.watch(wsClientProvider);
  ws.connect();
  return ws.events;
});

class CharactersListController extends AsyncNotifier<List<CharacterSummary>> {
  @override
  Future<List<CharacterSummary>> build() async {
    final api = CharacterApi(ref.read(apiClientProvider));
    return api.list();
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    final api = CharacterApi(ref.read(apiClientProvider));
    final l = await api.list();
    if (!ref.mounted) return;
    state = AsyncData(l);
  }
}

final charactersListProvider =
    AsyncNotifierProvider<CharactersListController, List<CharacterSummary>>(
      CharactersListController.new,
    );

class CampaignsListController extends AsyncNotifier<List<Campaign>> {
  @override
  Future<List<Campaign>> build() async {
    final api = CampaignApi(ref.read(apiClientProvider));
    return api.list();
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    final api = CampaignApi(ref.read(apiClientProvider));
    final l = await api.list();
    if (!ref.mounted) return;
    state = AsyncData(l);
  }
}

final campaignsListProvider =
    AsyncNotifierProvider<CampaignsListController, List<Campaign>>(
      CampaignsListController.new,
    );
