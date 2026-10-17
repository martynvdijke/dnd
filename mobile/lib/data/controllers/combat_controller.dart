import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';
import '../api/combat_api.dart';
import '../models/combat_entry.dart';

class CombatController extends AsyncNotifier<List<CombatEntry>> {
  final int arg;
  CombatController(this.arg);
  CombatApi get _api => CombatApi(ref.read(apiClientProvider));
  @override
  Future<List<CombatEntry>> build() async {
    final list = await _api.list(arg);
    ref.listen(wsEventsProvider, (prev, next) {
      next.whenData((e) {
        if (!ref.mounted) return;
        if (e.type == 'combat_update') {
          final cid = e.payload['campaign_id'];
          if (cid == arg) {
            _api.list(arg).then((l) {
              if (ref.mounted) state = AsyncData(l);
            });
          }
        }
      });
    });
    return list;
  }

  Future<void> refresh() async {
    final l = await _api.list(arg);
    if (!ref.mounted) return;
    state = AsyncData(l);
  }
}

final combatProvider =
    AsyncNotifierProvider.family<CombatController, List<CombatEntry>, int>(
      (arg) => CombatController(arg),
    );
