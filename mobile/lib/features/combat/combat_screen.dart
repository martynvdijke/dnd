import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/api_client.dart';
import '../../core/providers.dart';
import '../../data/models/attack.dart';
import '../../data/models/character_summary.dart';
import '../../data/models/combat_entry.dart';
import '../shared/states.dart';
import '../shared/stat_block.dart';
import '../shared/section_header.dart';

class CombatScreen extends ConsumerStatefulWidget {
  final int? campaignId;

  /// The character the roller should default to attacking with (for example the
  /// sheet the user just came from). Falls back to the first character.
  final int? initialCharacterId;

  const CombatScreen({super.key, this.campaignId, this.initialCharacterId});
  @override
  ConsumerState<CombatScreen> createState() => _CombatScreenState();
}

/// A server-side attack target: either a combat entry (`/api/combat`) or a
/// character. Combat entries win when initiative is running because they carry
/// the AC the server actually uses.
class _AttackTarget {
  final String type; // 'character' | 'combat'
  final int id;
  final String name;
  final String detail;
  final bool friendly;
  const _AttackTarget({
    required this.type,
    required this.id,
    required this.name,
    required this.detail,
    required this.friendly,
  });
  String get key => '$type:$id';
}

class _CombatScreenState extends ConsumerState<CombatScreen> {
  int? _campaignId;
  String _adv = 'normal';
  bool _apply = true;
  AttackResult? _lastResult;
  bool _rolling = false;
  int? _attackerId;
  String? _targetKey;
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
    final entries = _campaignId == null
        ? const <CombatEntry>[]
        : ref.watch(combatProvider(_campaignId!)).value ?? const <CombatEntry>[];
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
            final attacker = _resolveAttacker(list);
            final targets = _targetsFor(list, entries, attacker);
            final target = _resolveTarget(targets);
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // The attacker is a real choice now: it defaults to the sheet
                // the user came from, never to `characters.first`.
                DropdownButtonFormField<int>(
                  key: ValueKey('attacker-$attacker'),
                  initialValue: attacker,
                  isExpanded: true,
                  decoration: const InputDecoration(
                    labelText: 'Attacker',
                    prefixIcon: Icon(Icons.person_outline),
                  ),
                  items: list
                      .map(
                        (c) => DropdownMenuItem<int>(
                          value: c.id,
                          child: Text(
                            c.name.isEmpty ? 'Unnamed' : c.name,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      )
                      .toList(),
                  onChanged: _rolling
                      ? null
                      : (v) => setState(() => _attackerId = v),
                ),
                const SizedBox(height: 10),
                if (targets.isEmpty)
                  Text(
                    'No attack targets yet — add combatants to the initiative '
                    'tracker or another character.',
                    style: GoogleFonts.lora(
                      fontSize: 12,
                      color: VillumTheme.parchmentMuted,
                    ),
                  )
                else
                  DropdownButtonFormField<String>(
                    key: ValueKey('target-${target?.key}'),
                    initialValue: target?.key,
                    isExpanded: true,
                    decoration: const InputDecoration(
                      labelText: 'Target',
                      prefixIcon: Icon(Icons.gps_fixed),
                    ),
                    items: targets
                        .map(
                          (t) => DropdownMenuItem<String>(
                            value: t.key,
                            child: Text(
                              t.detail.isEmpty
                                  ? t.name
                                  : '${t.name} · ${t.detail}',
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        )
                        .toList(),
                    onChanged: _rolling
                        ? null
                        : (v) => setState(() => _targetKey = v),
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
                        onPressed: (_rolling || target == null)
                            ? null
                            : () => _roll(attacker, target),
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

  /// Default attacker: the character this screen was opened for, else the first
  /// character (an explicit dropdown choice always wins).
  int _resolveAttacker(List<CharacterSummary> chars) {
    final preferred = _attackerId ?? widget.initialCharacterId;
    if (preferred != null) {
      for (final c in chars) {
        if (c.id == preferred) return c.id;
      }
    }
    return chars.first.id;
  }

  /// Combat entries are the real targets while initiative is tracked; without
  /// initiative the party's characters are the only targets available.
  List<_AttackTarget> _targetsFor(
    List<CharacterSummary> chars,
    List<CombatEntry> entries,
    int attackerId,
  ) {
    if (entries.isNotEmpty) {
      return entries
          .where((e) => e.characterId != attackerId)
          .map(
            (e) => _AttackTarget(
              type: 'combat',
              id: e.id,
              name: e.name,
              detail: 'HP ${e.hpCurrent}/${e.hpMax} · AC ${e.ac}',
              friendly: e.type == 'pc',
            ),
          )
          .toList();
    }
    return chars
        .map(
          (c) => _AttackTarget(
            type: 'character',
            id: c.id,
            name: c.name,
            detail: 'Lv ${c.level} · ${c.class_}',
            friendly: true,
          ),
        )
        .toList();
  }

  _AttackTarget? _resolveTarget(List<_AttackTarget> targets) {
    if (targets.isEmpty) return null;
    final chosen = _targetKey;
    if (chosen != null) {
      for (final t in targets) {
        if (t.key == chosen) return t;
      }
    }
    for (final t in targets) {
      if (!t.friendly) return t;
    }
    return targets.first;
  }

  Future<void> _roll(int attackerId, _AttackTarget target) async {
    setState(() => _rolling = true);
    try {
      final api = ref.read(apiClientProvider);
      String? adv;
      if (_adv == 'advantage') adv = 'advantage';
      if (_adv == 'disadvantage') adv = 'disadvantage';
      final req = AttackRequest(
        attackerType: 'character',
        attackerId: attackerId,
        targetType: target.type,
        targetId: target.id,
        advantage: adv,
        apply: _apply,
        campaignId: _campaignId,
      );
      // Only the server's answer is ever shown; a failed call surfaces an error
      // instead of inventing a roll.
      final resp = await api.post('/api/combat/attack', data: req.toJson());
      final data = resp.data;
      if (data is! Map) {
        throw ApiException(
          resp.statusCode ?? 0,
          'The server returned an unexpected attack response.',
        );
      }
      final result = AttackResult.fromJson(Map<String, dynamic>.from(data));
      if (!mounted) return;
      setState(() => _lastResult = result);
      HapticFeedback.heavyImpact();
    } catch (e) {
      if (!mounted) return;
      final message = e is ApiException
          ? (e.statusCode == 0 ? e.message : 'HTTP ${e.statusCode} — ${e.message}')
          : e.toString();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Attack failed: $message'),
          action: SnackBarAction(
            label: 'Retry',
            onPressed: () => _roll(attackerId, target),
          ),
        ),
      );
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
