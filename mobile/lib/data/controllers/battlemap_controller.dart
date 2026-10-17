import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';
import '../api/battlemap_api.dart';

class BattlemapController extends AsyncNotifier<BattlemapState> {
  final int arg;
  BattlemapController(this.arg);
  BattlemapApi get _api => BattlemapApi(ref.read(apiClientProvider));
  @override
  Future<BattlemapState> build() async {
    final s = await _api.get(arg);
    ref.listen(wsEventsProvider, (prev, next) {
      next.whenData((e) {
        if (!ref.mounted) return;
        if (e.type == 'battlemap_update') {
          final cid = e.payload['campaign_id'];
          if (cid == arg) {
            _api.get(arg).then((st) {
              if (ref.mounted) state = AsyncData(st);
            });
          }
        }
      });
    });
    return s;
  }

  Future<void> moveToken(int tokenId, double x, double y) async {
    final prev = state.value;
    if (prev == null) return;
    final idx = prev.tokens.indexWhere((t) => t.id == tokenId);
    if (idx == -1) return;
    final old = prev.tokens[idx];
    final optimistic = BattlemapToken(
      id: old.id,
      campaignId: old.campaignId,
      combatEntryId: old.combatEntryId,
      name: old.name,
      x: x,
      y: y,
      color: old.color,
      size: old.size,
      hpCurrent: old.hpCurrent,
      hpMax: old.hpMax,
      ac: old.ac,
      conditions: old.conditions,
      auraRadius: old.auraRadius,
      visionRadius: old.visionRadius,
    );
    final newTokens = List<BattlemapToken>.from(prev.tokens);
    newTokens[idx] = optimistic;
    state = AsyncData(BattlemapState(map: prev.map, tokens: newTokens));
    try {
      await _api.updateToken(tokenId, {'x': x, 'y': y});
      if (!ref.mounted) return;
    } catch (e) {
      if (!ref.mounted) return;
      state = AsyncData(prev);
      rethrow;
    }
  }

  Future<void> refresh() async {
    final s = await _api.get(arg);
    if (!ref.mounted) return;
    state = AsyncData(s);
  }
}

final battlemapProvider =
    AsyncNotifierProvider.family<BattlemapController, BattlemapState, int>(
      (arg) => BattlemapController(arg),
    );
