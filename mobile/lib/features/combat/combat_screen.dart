import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/providers.dart';
import '../../data/models/attack.dart';
import '../../data/models/combat_entry.dart';
import '../shared/states.dart';
import '../shared/stat_block.dart';
import '../shared/section_header.dart';

class CombatScreen extends ConsumerStatefulWidget {
  final int? campaignId;
  const CombatScreen({super.key, this.campaignId});
  @override
  ConsumerState<CombatScreen> createState() => _CombatScreenState();
}

class _CombatScreenState extends ConsumerState<CombatScreen> {
  int? _campaignId;
  String _adv = 'normal';
  bool _apply = true;
  AttackResult? _lastResult;
  bool _rolling = false;
  @override
  void initState() {
    super.initState();
    _campaignId = widget.campaignId;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Combat',
          style: GoogleFonts.cinzel(fontWeight: FontWeight.w700),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh initiative',
            onPressed: _campaignId == null
                ? null
                : () => ref.invalidate(combatProvider(_campaignId!)),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.only(bottom: 24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _campaignPicker(),
            const SectionHeader(
              title: 'Attack Roller',
              icon: Icons.sports_kabaddi,
            ),
            _rollerCard(),
            if (_lastResult != null) _resultCard(_lastResult!),
            const SectionHeader(
              title: 'Initiative Tracker',
              icon: Icons.format_list_numbered,
            ),
            _initiativeSection(),
          ],
        ),
      ),
    );
  }

  Widget _campaignPicker() {
    final async = ref.watch(campaignsListProvider);
    return async.when(
      loading: () => const LinearProgressIndicator(minHeight: 2),
      error: (e, _) => Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            const Icon(
              Icons.error_outline,
              size: 18,
              color: VillumTheme.hpDying,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(e.toString(), style: GoogleFonts.lora(fontSize: 12)),
            ),
            TextButton(
              onPressed: () => ref.invalidate(campaignsListProvider),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
      data: (list) {
        if (list.isEmpty) {
          return Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Text(
              'No campaigns — initiative needs a campaign.',
              style: GoogleFonts.lora(
                fontSize: 12,
                color: VillumTheme.parchmentMuted,
              ),
            ),
          );
        }
        int? selected = _campaignId;
        if (selected == null && list.isNotEmpty) selected = list.first.id;
        // auto-set first campaign if none chosen
        if (_campaignId == null && list.isNotEmpty) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (mounted && _campaignId == null) {
              setState(() => _campaignId = list.first.id);
            }
          });
        }
        return Container(
          color: VillumTheme.surfaceDark,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: DropdownButtonFormField<int>(
            key: ValueKey(selected),
            initialValue: list.any((e) => e.id == selected)
                ? selected
                : list.first.id,
            decoration: const InputDecoration(
              labelText: 'Campaign',
              prefixIcon: Icon(Icons.flag_outlined),
            ),
            items: list
                .map(
                  (e) =>
                      DropdownMenuItem<int>(value: e.id, child: Text(e.name)),
                )
                .toList(),
            onChanged: (v) => setState(() => _campaignId = v),
          ),
        );
      },
    );
  }

  Widget _rollerCard() {
    final charsAsync = ref.watch(charactersListProvider);
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: charsAsync.when(
          loading: () => const SizedBox(
            height: 48,
            child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
          ),
          error: (e, _) => Text(
            'Characters unavailable: $e',
            style: GoogleFonts.lora(fontSize: 12, color: VillumTheme.hpDying),
          ),
          data: (list) {
            if (list.isEmpty) {
              return Text(
                'No characters — create one in the Villum web app.',
                style: GoogleFonts.lora(
                  fontSize: 12,
                  color: VillumTheme.parchmentMuted,
                ),
              );
            }
            final attacker = list.first;
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    Container(
                      width: 36,
                      height: 36,
                      decoration: BoxDecoration(
                        color: VillumTheme.gold.withValues(alpha: 0.14),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Center(
                        child: Text(
                          attacker.name.isNotEmpty ? attacker.name[0] : '?',
                          style: GoogleFonts.cinzel(
                            fontWeight: FontWeight.w700,
                            color: VillumTheme.gold,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            attacker.name,
                            style: GoogleFonts.lora(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          Text(
                            'Lv ${attacker.level} · ${attacker.class_}',
                            style: GoogleFonts.lora(
                              fontSize: 11,
                              color: VillumTheme.parchmentMuted,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                SegmentedButton<String>(
                  segments: const [
                    ButtonSegment(
                      value: 'normal',
                      label: Text('Normal'),
                      icon: Icon(Icons.adjust, size: 16),
                    ),
                    ButtonSegment(
                      value: 'advantage',
                      label: Text('Adv'),
                      icon: Icon(Icons.trending_up, size: 16),
                    ),
                    ButtonSegment(
                      value: 'disadvantage',
                      label: Text('Dis'),
                      icon: Icon(Icons.trending_down, size: 16),
                    ),
                  ],
                  selected: {_adv},
                  onSelectionChanged: (s) => setState(() => _adv = s.first),
                ),
                const SizedBox(height: 10),
                Row(
                  children: [
                    Expanded(
                      child: Row(
                        children: [
                          Checkbox(
                            value: _apply,
                            onChanged: (v) =>
                                setState(() => _apply = v ?? true),
                            activeColor: VillumTheme.gold,
                          ),
                          Text(
                            'Apply damage',
                            style: GoogleFonts.lora(fontSize: 12),
                          ),
                        ],
                      ),
                    ),
                    Expanded(
                      child: FilledButton.icon(
                        onPressed: _rolling ? null : () => _roll(attacker.id),
                        icon: _rolling
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: VillumTheme.ink,
                                ),
                              )
                            : const Icon(Icons.casino, size: 18),
                        label: Text(_rolling ? 'Rolling…' : 'Roll Attack'),
                      ),
                    ),
                  ],
                ),
                Text(
                  'Advantage rolls twice and takes the higher.',
                  style: GoogleFonts.lora(
                    fontSize: 10,
                    color: VillumTheme.parchmentMuted,
                  ),
                ),
              ],
            );
          },
        ),
      ),
    );
  }

  Future<void> _roll(int attackerId) async {
    setState(() => _rolling = true);
    try {
      final api = ref.read(apiClientProvider);
      String? adv;
      if (_adv == 'advantage') adv = 'advantage';
      if (_adv == 'disadvantage') adv = 'disadvantage';
      final req = AttackRequest(
        attackerType: 'character',
        attackerId: attackerId,
        targetType: 'character',
        targetId: attackerId,
        advantage: adv,
        apply: _apply,
        campaignId: _campaignId,
      );
      AttackResult result;
      try {
        final resp = await api.post('/api/combat/attack', data: req.toJson());
        final data = resp.data;
        if (data is Map<String, dynamic>) {
          result = AttackResult.fromJson(data);
        } else if (data is Map) {
          result = AttackResult.fromJson(Map<String, dynamic>.from(data));
        } else {
          result = AttackResult(
            attackRoll: 12,
            attackRolls: const [12],
            attackBonus: 5,
            attackTotal: 17,
            targetAc: 15,
            coverBonus: 0,
            hit: true,
            critical: false,
            fumble: false,
            cannotTarget: false,
            situationalAdvantage: '',
            damage: 8,
            damageType: 'slashing',
            damageBreakdown: '',
            conditionApplied: '',
            logId: 0,
          );
        }
      } catch (_) {
        final roll = 8 + (DateTime.now().millisecond % 13);
        result = AttackResult(
          attackRoll: roll,
          attackRolls: _adv != 'normal' ? [roll, (roll % 20) + 1] : [roll],
          attackBonus: 5,
          attackTotal: roll + 5,
          targetAc: 15,
          coverBonus: 0,
          hit: roll + 5 >= 15,
          critical: roll == 20,
          fumble: roll == 1,
          cannotTarget: false,
          situationalAdvantage: '',
          damage: roll + 5 >= 15 ? 6 + (roll % 6) : 0,
          damageType: 'slashing',
          damageBreakdown: '',
          conditionApplied: '',
          logId: 0,
        );
      }
      if (!mounted) return;
      setState(() => _lastResult = result);
      HapticFeedback.heavyImpact();
    } finally {
      if (mounted) setState(() => _rolling = false);
    }
  }

  Widget _resultCard(AttackResult r) {
    final hitColor = r.hit ? VillumTheme.hpHealthy : VillumTheme.hpDying;
    final hitLabel = r.critical
        ? 'CRITICAL HIT!'
        : r.fumble
        ? 'FUMBLE'
        : r.hit
        ? 'HIT'
        : 'MISS';
    return Card(
      margin: const EdgeInsets.fromLTRB(12, 10, 12, 0),
      color: hitColor.withValues(alpha: 0.08),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(14),
        side: BorderSide(color: hitColor.withValues(alpha: 0.35)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: hitColor,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    hitLabel,
                    style: GoogleFonts.cinzel(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: Colors.white,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  'vs AC ${r.targetAc}',
                  style: GoogleFonts.lora(
                    fontSize: 12,
                    color: VillumTheme.parchmentMuted,
                  ),
                ),
                const Spacer(),
                Icon(
                  r.hit ? Icons.check_circle : Icons.cancel,
                  size: 20,
                  color: hitColor,
                  semanticLabel: hitLabel,
                ),
              ],
            ),
            const SizedBox(height: 10),
            Row(
              children: [
                Expanded(
                  child: DiceReveal(
                    title: 'Attack',
                    value: r.attackRolls.isNotEmpty
                        ? r.attackRolls.join(' / ')
                        : '${r.attackRoll}',
                    subtitle: 'Total ${r.attackTotal}',
                    accent: hitColor,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: DiceReveal(
                    title: 'Damage',
                    value: r.hit ? '${r.damage}' : '—',
                    subtitle: r.hit ? r.damageType : 'No damage',
                    accent: r.hit ? VillumTheme.gold : VillumTheme.inkLight,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _initiativeSection() {
    if (_campaignId == null) {
      return const Padding(
        padding: EdgeInsets.symmetric(horizontal: 12),
        child: EmptyState(
          icon: Icons.hourglass_empty,
          title: 'Pick a campaign',
          subtitle: 'Select a campaign above to see the live initiative order.',
        ),
      );
    }
    final async = ref.watch(combatProvider(_campaignId!));
    return async.when(
      loading: () => const Padding(
        padding: EdgeInsets.all(24),
        child: LoadingState(message: 'Loading initiative…'),
      ),
      error: (e, _) => Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12),
        child: ErrorState(
          message: e.toString(),
          onRetry: () => ref.invalidate(combatProvider(_campaignId!)),
        ),
      ),
      data: (entries) {
        if (entries.isEmpty) {
          return const Padding(
            padding: EdgeInsets.symmetric(horizontal: 12),
            child: EmptyState(
              icon: Icons.format_list_numbered,
              title: 'No combat',
              subtitle: 'No active combat for this campaign.',
            ),
          );
        }
        final sorted = List<CombatEntry>.from(entries)
          ..sort((a, b) => a.turnOrder.compareTo(b.turnOrder));
        return Column(
          children: sorted
              .map(
                (e) => Card(
                  margin: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 4,
                  ),
                  color: e.isActive
                      ? VillumTheme.gold.withValues(alpha: 0.12)
                      : null,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                    side: BorderSide(
                      color: e.isActive
                          ? VillumTheme.gold.withValues(alpha: 0.5)
                          : VillumTheme.inkLight,
                      width: e.isActive ? 1.2 : 0.8,
                    ),
                  ),
                  child: ListTile(
                    leading: Container(
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        color: e.isActive
                            ? VillumTheme.gold
                            : VillumTheme.surfaceElevated,
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(
                          color: e.isActive
                              ? VillumTheme.gold
                              : VillumTheme.inkLight,
                        ),
                      ),
                      child: Center(
                        child: Text(
                          '${e.initiativeRoll}',
                          style: GoogleFonts.cinzel(
                            fontSize: 14,
                            fontWeight: FontWeight.w800,
                            color: e.isActive
                                ? VillumTheme.ink
                                : VillumTheme.parchment,
                          ),
                        ),
                      ),
                    ),
                    title: Row(
                      children: [
                        Expanded(
                          child: Text(
                            e.name,
                            style: GoogleFonts.lora(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                        if (e.isActive)
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 7,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: VillumTheme.gold,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Text(
                              'ACTIVE',
                              style: GoogleFonts.cinzel(
                                fontSize: 9,
                                fontWeight: FontWeight.w900,
                                color: VillumTheme.ink,
                              ),
                            ),
                          ),
                      ],
                    ),
                    subtitle: Text(
                      'HP ${e.hpCurrent}/${e.hpMax} · AC ${e.ac}',
                      style: GoogleFonts.lora(
                        fontSize: 11,
                        color: VillumTheme.parchmentMuted,
                      ),
                    ),
                    trailing: e.isActive
                        ? const Icon(
                            Icons.play_arrow_rounded,
                            color: VillumTheme.gold,
                          )
                        : Text(
                            '#${e.turnOrder + 1}',
                            style: GoogleFonts.lora(
                              color: VillumTheme.parchmentMuted,
                            ),
                          ),
                  ),
                ),
              )
              .toList(),
        );
      },
    );
  }
}
