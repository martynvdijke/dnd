# Villum Mobile (`mobile/`)

Flutter Android pocket companion for Villum. Targets the existing Go `/api/*` + `/api/ws` endpoints.

> This file is the **frozen interface contract** between the mobile lanes. Field names, provider
> names, and directory ownership below are binding. Change them here first, then in code.

## Stack

Flutter 3.47.2 / Dart 3.13.2. `flutter_riverpod ^3.4.3` (Riverpod **3** — no codegen, no
`StateNotifier`), `dio ^5.11.1`, `web_socket_channel ^3.0.3`, `go_router ^18.1.0`,
`flutter_secure_storage ^11.2.0`, `crypto ^3.0.7`, `google_fonts ^9.0.0`.

Android: `applicationId = xyz.vandijke.dnd`, **`minSdk` must be 24** (`flutter_secure_storage` 11
requirement). `flutter build apk` needs JDK 17 and an Android SDK carrying the NDK pinned by
`flutter.ndkVersion` (`28.2.13676358`). Both are present in this container — JDK 17 at
`/usr/lib/jvm/java-17-openjdk` (the default JDK 11 is too old for Gradle 9.3.1 / AGP 9.1.0) and the
SDK at `/opt/android-sdk`. CI builds the same APK from a pristine checkout.

## Directory ownership

| Path | Owner | Contents |
|---|---|---|
| `lib/main.dart`, `lib/app/**`, `lib/features/**` | UI lane | App root, router, theme, screens/widgets |
| `lib/core/**`, `lib/data/**`, `test/core/**` | Core lane | ApiClient, auth, secure storage, WS, providers, models, repos |
| `test/**` (widget/integration) | Test lane | Widget + repository tests |
| `android/**`, `.github/workflows/flutter-android.yaml`, `.releaserc.yaml`, `.github/workflows/release.yaml` | CI lane | Build config, workflows, release wiring |

`pubspec.yaml` is owned by the Core lane. Ask before adding a dependency.

## Auth protocol (read this before writing the client)

Auth is **three-layer**. There is **no `/api/register`** — accounts are created by an admin, or by
the first-run setup path (`POST /api/login` with `setup: true`).

**Sign-in is OIDC-first.** The server retires password login by default
(`ALLOW_PASSWORD_FALLBACK=true` is break-glass recovery only), so the app's primary action is
"Sign in with Authelia", which runs the server's OIDC flow in an in-app WebView:

1. `GET /api/auth/oidc/status` (public) → `{"enabled": bool, "password_enabled"?: bool}`. Treat any
   failure as `enabled: false`. This decides whether SSO is offered (only once a server is set up;
   the first admin must still be created with the password setup path).
2. `GET /api/auth/oidc/login` inside a WebView; Authelia authenticates, the server's callback sets
   `Set-Cookie: session=…` (`HttpOnly`, so JavaScript cannot read it) and redirects to `/`.
3. Read the cookie from the **native** cookie store over the `xyz.vandijke.dnd/cookies` method
   channel (`MainActivity.kt`), then persist it with `AuthRepository.loginWithSession`
   (`GET /api/user/me` + `POST /api/tokens`). Cookies are cleared before starting so a stale
   session is never captured.

Password login (`POST /api/login`) still exists for break-glass and first-admin creation:

1. `GET /api/check-setup` (public) → `{"setup": bool}`. `false` means no admin exists yet → offer
   first-admin creation.
2. `POST /api/login` (public) `{"username","password","setup"?:bool}` → `200 {"session": string,
   "user": {"id": int64, "username": string, "role": string}}` and `Set-Cookie: session=…`.
   Store `session`, `user`, and the server URL in secure storage. A `403` means password auth is
   retired — point the user at SSO instead of reporting bad credentials.
3. `POST /api/tokens` (needs `Cookie: session` + `X-CSRF-Token`) `{"name","expires_in_days":0}` →
   `201 {…, "token": "vlt_…"}` (secret shown once). Store it; it is the bearer token.
4. Every **mutating** request (`POST/PUT/PATCH/DELETE`) must send **all three**:
   - `Cookie: session=<session>`
   - `X-CSRF-Token: sha256_hex(session + "-csrf")` — derivable locally, or via `GET /api/csrf-token`
   - `Authorization: Bearer vlt_…`
   `GET` requests need only the session cookie. Use `package:crypto` `sha256`.
5. `POST /api/logout` (mutation → all three headers).
6. Roles: `role` is `admin` | `dm` | `player`. `isDm` ⇒ `dm` or `admin`.

Auth state machine: `AuthState` = `AuthUnauthenticated` | `AuthNeedsSetup(serverUrl)` |
`AuthAuthenticated(user, serverUrl)`.

**There is no default server address.** The app talks only to a server the user names: the URL field
starts empty (prefilled from secure storage when a previous value exists), input is validated as an
absolute `http`/`https` URL, and `ApiClient`/`WsClient` refuse to fire while `serverUrlProvider` is
empty.

## Frozen Dart interface (Core lane implements; UI lane consumes)

```dart
// lib/core/auth_store.dart
class StoredAuth { final String serverUrl, sessionId; final String? apiToken; final int userId;
                   final String username, role; }
abstract class AuthStore {
  Future<StoredAuth?> read();
  Future<void> write(StoredAuth auth);
  Future<void> clear();
}

// lib/core/api_client.dart
class ApiException implements Exception { final int statusCode; final String message; }
class ApiClient {
  ApiClient({required String baseUrl, required AuthStore store});
  String get baseUrl;
  Dio get dio;                                  // configured: session cookie + CSRF + bearer
  Future<void> ensureApiToken();                // idempotent POST /api/tokens
}

// lib/core/auth_repository.dart
class UserSession { final int id; final String username, role; bool get isDm; }
sealed class AuthState { const AuthState(); }
class AuthUnauthenticated extends AuthState { const AuthUnauthenticated(); }
class AuthNeedsSetup extends AuthState { final String serverUrl; }
class AuthAuthenticated extends AuthState { final UserSession user; final String serverUrl; }

class AuthRepository {
  Future<bool> checkSetup(String serverUrl);
  Future<bool> oidcEnabled(String serverUrl);              // GET /api/auth/oidc/status, failures => false
  Future<UserSession> login({required String serverUrl, required String username,
                             required String password, bool setup = false});
  Future<UserSession> loginWithSession({required String serverUrl, required String sessionId});
  Future<AuthState> restore();
  Future<void> logout();
}

// lib/core/server_url.dart
String normalizeServerUrl(String raw);                    // trim + strip trailing slashes
String? serverUrlError(String raw);                       // null when valid, else a user-facing message

// lib/core/api_client.dart (also)
String? parseCookieHeader(String? header, String name);   // reads `session` out of a Set-Cookie dump

// lib/core/ws_client.dart
class WsEvent { final String type; final Map<String, dynamic> payload; }
class WsClient {
  WsClient({required String baseUrl, required AuthStore store});
  Stream<WsEvent> get events;   // broadcast
  void connect();
  Future<void> dispose();
}

// lib/core/providers.dart  — provider NAMES are the public API
final authStoreProvider        = Provider<AuthStore>(...);
final apiClientProvider        = Provider<ApiClient>(...);
final authRepositoryProvider   = Provider<AuthRepository>(...);
final authControllerProvider   = NotifierProvider<AuthController, AuthState>(...); // AsyncNotifierProvider is acceptable if restore() must be async at boot
final wsClientProvider         = Provider<WsClient>(...);
final wsEventsProvider         = StreamProvider<WsEvent>(...);
final serverUrlProvider        = Provider<String>(...);
final characterProvider        = AsyncNotifierProvider.family<CharacterController, Character, int>(...); // by character id
final combatProvider           = AsyncNotifierProvider.family<CombatController, List<CombatEntry>, int>(...); // by campaign id
final battlemapProvider        = AsyncNotifierProvider.family<BattlemapController, BattlemapState, int>(...); // by campaign id
final dmScreenProvider         = AsyncNotifierProvider.family<DmScreenController, DmScreen, int>(...); // by campaign id
```

Riverpod 3 notes: single non-generic `Ref`; after every `await` guard with `if (!ref.mounted)
return;`; providers compare with `==` (models need value equality or `updateShouldNotify`
overrides); automatic retry with backoff is on by default — opt out for auth calls.

## Frozen model field names (`lib/data/models/`)

Models are hand-written immutable classes with `fromJson`/`toJson` and `==`/`hashCode`. JSON keys
and DB shapes come from `handlers/`; snake_case in JSON, camelCase in Dart.

- `Character` — `id, userId, name, race, class_, subclass, level, xp, background, alignment, str,
  dex, con, int_, wis, cha, ac, initiative, speed, hpMax, hpCurrent, tempHp, hitDice,
  hitDiceCurrent, proficiencyBonus, inspiration, passivePerception, deathSavesSuccesses,
  deathSavesFailures, concentratingOn, portraitUrl, characterType, canEdit` + computed
  `strMod..chaMod, spellSaveDc, spellAttackBonus`; sub-lists `features, spellcasting, spells,
  inventory, currency, classes, campaigns`.
- `Currency` — `cp, sp, ep, gp, pp` (all `int`).
- `Spellcasting` — `ability, saveDc, attackBonus, slotsMax` (`List<int>` len 9), `slotsUsed`
  (`List<int>` len 9). JSON uses `slots_1_max..slots_9_max` / `slots_1_used..slots_9_used`.
- `Spell` — `id, name, level, school, castingTime, range, components, duration, description,
  prepared, alwaysPrepared, source, notes`.
- `InventoryItem` — `id, name, quantity, weight, category, damageDice, damageType,
  weaponProperties, acBonus, armorType, description, isEquipped, isMagical, attunement,
  isIdentified, notes, attackAbility, attackBonus`.
- `Feature` — `id, name, description, source, levelGained`.
- `CharClass` — `id, class_, subclass, level, hitDice`.
- `HpChangeResult` — `hpCurrent, tempHp, hpMax, deathSavesSuccesses, deathSavesFailures,
  concentration?, defenseApplied?, effectiveDamage?, hpMaxEffective?`.
- `AttackResult` — `attackRoll, attackRolls, attackBonus, attackTotal, targetAc, coverBonus, hit,
  critical, fumble, cannotTarget, situationalAdvantage, damage, damageType, damageBreakdown,
  conditionApplied, logId, applied?, targetHp?`.
- `AttackRequest` — `attackerType, attackerId, targetType, targetId, itemId?, attackBonus?,
  damageDice?, damageType?, advantage?, apply, condition?, conditionDuration?, campaignId?, cover?,
  flanking?, distance?, normalRange?, longRange?`.
- `CombatEntry` — `id, characterId?, campaignId?, name, type, initiativeRoll, initiativeMod, hpMax,
  hpCurrent, ac, isActive, turnOrder, conditionIds, notes, actionUsed, bonusActionUsed,
  reactionUsed, movementUsed, movementMax`.
- `BattlemapMeta` — `id, name, imageUrl, width, height, gridSize, gridUnits, walls`.
- `BattlemapToken` — `id, campaignId, combatEntryId?, name, x, y, color, size, hpCurrent?, hpMax?,
  ac?, conditions?, auraRadius, visionRadius`. **`x`/`y` are normalised 0..1**, not pixels.
- `BattlemapState` — `map` (`BattlemapMeta?`), `tokens` (`List<BattlemapToken>`).
- `CompendiumSearchResult` — `type, id, name, subtype?, cr?, level?, hitDie?, primaryAbility?`.
- `DraftSession` — `id, entityType, status, message, draft?, messages` (`List<DraftMessage>`,
  `DraftMessage` = `role, content`).
- `DmScreen` — `campaign, session?, combat, party, ambience` with sub-types
  `DmScreenCampaign(id, name, partyName, dmNotes)`, `DmScreenCombat(active, currentTurn, entries)`,
  `DmScreenParty(id, name, hpCurrent, hpMax, ac, conditions)`, `DmScreenAmbience(track, action)`.

## Real endpoint paths (artifacts sometimes name these loosely)

Server default port `6270`; Docker `6280` — but the app has **no default host**: the user supplies
the address. WebSocket is **`/api/ws`** (not `/ws`).

| Purpose | Method + path |
|---|---|
| Setup probe | `GET /api/check-setup` |
| OIDC status | `GET /api/auth/oidc/status` |
| OIDC flow (WebView) | `GET /api/auth/oidc/login` → `/api/auth/oidc/callback` |
| Login / first-admin (break-glass) | `POST /api/login` |
| Logout | `POST /api/logout` |
| CSRF token | `GET /api/csrf-token` |
| Provision bearer token | `POST /api/tokens` |
| Current user | `GET /api/user/me` |
| Characters | `GET /api/characters`, `GET/PUT/DELETE /api/characters/:id` |
| HP change | `POST /api/characters/:id/hp` `{delta, type, source}` |
| Rest | `POST /api/characters/:id/rest` `{rest_type: "short"\|"long", hit_dice_count?}` |
| Death save | `POST /api/characters/:id/death-save` |
| Exhaustion | `PATCH /api/characters/:id/exhaustion` `{level}` |
| Cast spell (expend slot) | `POST /api/characters/:id/cast-spell` `{spell_id, slot_level?}` |
| Spellcasting config | `PUT /api/characters/:id/spellcasting` |
| Currency | `PUT /api/characters/:id/currency` |
| Inventory | `POST /api/characters/:id/inventory`, `PUT/DELETE /api/inventory/:iid` |
| Spells | `POST /api/characters/:id/spells`, `PUT/DELETE /api/spells/:sid`, `PUT /api/characters/:id/spells/prepare` |
| Features | `POST /api/characters/:id/features`, `PUT/DELETE /api/features/:fid` |
| Classes | `POST /api/characters/:id/classes`, `PUT/DELETE /api/classes/:ccid` |
| Attack roll | `POST /api/combat/attack` |
| Initiative | `GET /api/combat?campaign_id=`, `POST /api/combat` |
| Battlemap | `GET /api/campaigns/:id/battlemap`, `POST /api/campaigns/:id/battlemap/tokens` (DM), `PUT /api/battlemap-tokens/:id` (any member), `DELETE /api/battlemap-tokens/:id` (DM), `POST /api/campaigns/:id/battlemap/sync` (DM) |
| Compendium | `GET /api/compendium/search?q=&type=` (`spell\|equipment\|monster\|race\|feat\|background\|class`) |
| Generators | `GET /api/generate/{npc,name,encounter,loot,treasure,character,adventure-hook,dungeon-dressing,tavern,urban-encounter,road-encounter,weather}` |
| AI draft (DM only) | `POST /api/ai/draft`, `GET /api/ai/draft`, `GET /api/ai/draft/:id`, `POST /api/ai/draft/:id/turn`, `POST /api/ai/draft/:id/commit`, `DELETE /api/ai/draft/:id` |
| DM screen (DM only) | `GET /api/campaigns/:id/dm-screen` |
| Realtime | `GET /api/ws` — `Cookie: session=…` on the handshake, **no** CSRF/bearer, `IOWebSocketChannel.connect(headers: {...})` |

### WebSocket events

Server → client envelope: `{"type": string, "payload": object}`. Client → server messages are
**discarded** by the server. Types: `character_update {character_id}`, `party_update {}`,
`dice_roll {...}`, `combat_update {campaign_id}`, `knowledge_reveal {...}`, `scene_effect
{scene_id, effect}`, `spell_cast {...}`, `ambience {track, action, volume}`, `battlemap_update
{campaign_id}`. `combat_update` and `battlemap_update` carry **no diff** — refetch
`GET /api/combat?campaign_id=` / `GET /api/campaigns/:id/battlemap` on receipt.

### Player vs DM

Player-reachable: characters (own/party), combat read, battlemap read + token move, compendium,
generators, dice, WS. **DM/admin only**: `/api/ai/*` (draft assistant), `/api/campaigns/:id/dm-screen`,
battlemap token create/delete + `/sync`. Gate those screens on `UserSession.isDm` and show a
read-only/only-for-DM affordance otherwise.

## Local verification

```bash
cd mobile
flutter analyze          # must be clean
flutter test             # unit + widget tests

# Release APK: needs JDK 17 + Android SDK (both present in this container)
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk
export ANDROID_HOME=/opt/android-sdk ANDROID_SDK_ROOT=/opt/android-sdk
export PATH="$JAVA_HOME/bin:$PATH"
flutter build apk --release
```

A clean build produces `build/app/outputs/flutter-apk/app-release.apk` at **56,306,799 bytes**
containing exactly three ABIs — `arm64-v8a`, `armeabi-v7a`, `x86_64`. CI publishes that same APK as
`dnd-<version>-android.apk`, so the size and ABI list are a cheap parity check. The workflow runs
`flutter analyze`, `flutter test` and this build with `continue-on-error: true`.

### Troubleshooting: `armeabi is no longer supported`

Gradle resolves its native ABI set from a cache inside the **shared** pub cache
(`~/.pub-cache/hosted/pub.dev/jni-1.1.0/android/.cxx`; `jni` arrives via `flutter_secure_storage`).
If another checkout sharing that pub cache leaves a stale model behind, AGP configures the `:jni`
library for the legacy `armeabi` ABI and the build dies in CMake:

```
CMake Error at .../ndk/28.2.13676358/build/cmake/android-legacy.toolchain.cmake:165 (message):
  armeabi is no longer supported.  Use armeabi-v7a.
```

Nothing in this repo asks for `armeabi` (the Flutter Gradle plugin only ever sets the three modern
ABIs, and the app module declares no `abiFilters`), and CI is unaffected because every CI build
starts pristine. Clear the stale state and rebuild:

```bash
flutter clean
rm -rf build .dart_tool /root/.pub-cache/hosted/pub.dev/jni-1.1.0/android/.cxx
flutter pub get && flutter build apk --release
```
