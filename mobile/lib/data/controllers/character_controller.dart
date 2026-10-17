import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/character.dart';
import '../models/currency.dart';
import '../../core/providers.dart';
import '../api/character_api.dart';

class CharacterController extends AsyncNotifier<Character> {
  final int arg;
  CharacterController(this.arg);
  CharacterApi get _api => CharacterApi(ref.read(apiClientProvider));
  @override
  Future<Character> build() async {
    final c = await _api.get(arg);
    ref.listen(wsEventsProvider, (prev, next) {
      next.whenData((event) {
        if (!ref.mounted) return;
        if (event.type == 'character_update' || event.type == 'party_update') {
          final cid = event.payload['character_id'];
          if (event.type == 'party_update' || (cid != null && cid == arg)) {
            _api.get(arg).then((ch) {
              if (ref.mounted) state = AsyncData(ch);
            });
          }
        }
      });
    });
    return c;
  }

  Future<void> changeHp(
    int delta, {
    String type = '',
    String source = '',
  }) async {
    final prev = state.value;
    if (prev == null) return;
    final optimistic = prev.copyWith(
      hpCurrent: (prev.hpCurrent + delta).clamp(0, prev.hpMax),
      tempHp: prev.tempHp,
    );
    state = AsyncData(optimistic);
    try {
      final result = await _api.changeHp(arg, delta, type, source);
      if (!ref.mounted) return;
      state = AsyncData(
        prev.copyWith(hpCurrent: result.hpCurrent, tempHp: result.tempHp),
      );
    } catch (e) {
      if (!ref.mounted) return;
      state = AsyncData(prev);
      rethrow;
    }
  }

  Future<void> updateCurrency(Currency c) async {
    final prev = state.value;
    if (prev == null) return;
    state = AsyncData(prev.copyWith(currency: c));
    try {
      await _api.updateCurrency(arg, c);
      if (!ref.mounted) return;
    } catch (e) {
      if (!ref.mounted) return;
      state = AsyncData(prev);
      rethrow;
    }
  }
}

final characterProvider =
    AsyncNotifierProvider.family<CharacterController, Character, int>(
      (arg) => CharacterController(arg),
    );
