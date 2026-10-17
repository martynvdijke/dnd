import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';
import '../api/campaign_api.dart';

class DmScreenController extends AsyncNotifier<DmScreen> {
  final int arg;
  DmScreenController(this.arg);
  CampaignApi get _api => CampaignApi(ref.read(apiClientProvider));
  @override
  Future<DmScreen> build() async {
    final d = await _api.dmScreen(arg);
    ref.listen(wsEventsProvider, (prev, next) {
      next.whenData((e) {
        if (!ref.mounted) return;
        if (e.type == 'combat_update' ||
            e.type == 'character_update' ||
            e.type == 'battlemap_update') {
          _api.dmScreen(arg).then((dm) {
            if (ref.mounted) state = AsyncData(dm);
          });
        }
      });
    });
    return d;
  }
}

final dmScreenProvider =
    AsyncNotifierProvider.family<DmScreenController, DmScreen, int>(
      (arg) => DmScreenController(arg),
    );
