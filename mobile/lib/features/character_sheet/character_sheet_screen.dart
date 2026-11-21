import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/providers.dart';
import '../../data/api/character_api.dart';
import '../../data/models/character.dart';
import '../../data/models/currency.dart';
import '../../data/models/inventory_item.dart';
import '../../data/models/spell.dart';
import '../shared/hp_bar.dart';
import '../shared/section_header.dart';
import '../shared/stat_block.dart';
import '../ink/ink_screen.dart';
import '../shared/states.dart';

class CharacterSheetScreen extends ConsumerStatefulWidget {
  final int? initialCharacterId;
  const CharacterSheetScreen({super.key, this.initialCharacterId});

  @override
  ConsumerState<CharacterSheetScreen> createState() =>
      _CharacterSheetScreenState();
}

class _CharacterSheetScreenState extends ConsumerState<CharacterSheetScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tab;
  int? _selectedId;

  @override
  void initState() {
    super.initState();
    _tab = TabController(length: 5, vsync: this);
    _selectedId = widget.initialCharacterId;
  }

  @override
  void dispose() {
    _tab.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final charsAsync = ref.watch(charactersListProvider);
    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Character Sheet',
          style: GoogleFonts.cinzel(fontWeight: FontWeight.w700),
        ),
        bottom: TabBar(
          controller: _tab,
          isScrollable: true,
          labelColor: VillumTheme.gold,
          unselectedLabelColor: VillumTheme.parchmentMuted,
          indicatorColor: VillumTheme.gold,
          labelStyle: GoogleFonts.cinzel(
            fontSize: 11,
            fontWeight: FontWeight.w800,
            letterSpacing: 0.6,
          ),
          tabs: const [
            Tab(text: 'CORE'),
            Tab(text: 'COMBAT'),
            Tab(text: 'SPELLS'),
            Tab(text: 'INVENTORY'),
            Tab(text: 'FEATURES'),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.gesture),
            tooltip: 'Ink notes',
            onPressed: () =>
                Navigator.of(context)
                    .push(MaterialPageRoute(builder: (_) => const InkScreen())),
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () {
              ref.invalidate(charactersListProvider);
              if (_selectedId != null) {
                ref.invalidate(characterProvider(_selectedId!));
              }
            },
          ),
        ],
      ),
      body: Column(
        children: [
          charsAsync.when(
            loading: () => const LinearProgressIndicator(minHeight: 2),
            error: (e, _) => _PickerError(
              message: e.toString(),
              onRetry: () => ref.invalidate(charactersListProvider),
            ),
            data: (list) {
              if (list.isEmpty) {
                return const Padding(
                  padding: EdgeInsets.all(12),
                  child: EmptyState(
                    icon: Icons.person_search,
                    title: 'No characters',
                    subtitle: 'Create a character in the Villum web app first.',
                  ),
                );
              }
              if (_selectedId == null) {
                WidgetsBinding.instance.addPostFrameCallback((_) {
                  if (mounted && _selectedId == null) {
                    setState(() => _selectedId = list.first.id);
                  }
                });
              }
              final selected = _selectedId;
              return Container(
                color: VillumTheme.surfaceDark,
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 8,
                ),
                child: DropdownButtonFormField<int>(
                  key: ValueKey(selected),
                  initialValue: list.any((c) => c.id == selected)
                      ? selected
                      : list.first.id,
                  decoration: const InputDecoration(
                    labelText: 'Character',
                    prefixIcon: Icon(Icons.person),
                    contentPadding: EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 10,
                    ),
                  ),
                  items: list
                      .map(
                        (c) => DropdownMenuItem<int>(
                          value: c.id,
                          child: Text(
                            '${c.name} · Lv${c.level} ${c.class_}',
                            style: GoogleFonts.lora(
                              fontSize: 14,
                              color: VillumTheme.parchment,
                            ),
                          ),
                        ),
                      )
                      .toList(),
                  onChanged: (v) => setState(() => _selectedId = v),
                ),
              );
            },
          ),
          Expanded(
            child: _selectedId == null
                ? const LoadingState(message: 'Select a character')
                : _CharacterTabs(characterId: _selectedId!, tab: _tab),
          ),
        ],
      ),
    );
  }
}

class _PickerError extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;
  const _PickerError({required this.message, required this.onRetry});
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(12),
    child: Row(
      children: [
        const Icon(Icons.error_outline, color: VillumTheme.hpDying, size: 18),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            message,
            style: GoogleFonts.lora(
              fontSize: 12,
              color: VillumTheme.parchmentMuted,
            ),
          ),
        ),
        TextButton(onPressed: onRetry, child: const Text('Retry')),
      ],
    ),
  );
}

class _CharacterTabs extends ConsumerWidget {
  final int characterId;
  final TabController tab;
  const _CharacterTabs({required this.characterId, required this.tab});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(characterProvider(characterId));
    return async.when(
      loading: () => const LoadingState(message: 'Loading character…'),
      error: (e, _) => ErrorState(
        message: e.toString(),
        onRetry: () => ref.invalidate(characterProvider(characterId)),
      ),
      data: (c) => TabBarView(
        controller: tab,
        children: [
          _CoreTab(c: c),
          _CombatTab(c: c, characterId: characterId),
          _SpellsTab(c: c, characterId: characterId),
          _InventoryTab(c: c, characterId: characterId),
          _FeaturesTab(c: c),
        ],
      ),
    );
  }
}

class _CoreTab extends StatelessWidget {
  final Character c;
  const _CoreTab({required this.c});
  @override
  Widget build(BuildContext context) => SingleChildScrollView(
    padding: const EdgeInsets.only(bottom: 24),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Card(
          margin: const EdgeInsets.fromLTRB(12, 12, 12, 6),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Container(
                  width: 56,
                  height: 56,
                  decoration: BoxDecoration(
                    color: VillumTheme.gold.withValues(alpha: 0.14),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: VillumTheme.gold.withValues(alpha: 0.28),
                    ),
                  ),
                  child: Center(
                    child: Text(
                      c.name.isNotEmpty ? c.name[0].toUpperCase() : '?',
                      style: GoogleFonts.cinzel(
                        fontSize: 22,
                        fontWeight: FontWeight.w800,
                        color: VillumTheme.gold,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        c.name,
                        style: GoogleFonts.cinzel(
                          fontSize: 18,
                          fontWeight: FontWeight.w700,
                          color: VillumTheme.parchment,
                        ),
                      ),
                      Text(
                        '${c.race} · ${c.class_} ${c.subclass.isNotEmpty ? '(${c.subclass})' : ''} · Lv ${c.level}',
                        style: GoogleFonts.lora(
                          fontSize: 12,
                          color: VillumTheme.parchmentMuted,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Wrap(
                        spacing: 6,
                        children: [
                          _badge('AC ${c.ac}', Icons.shield),
                          _badge(
                            'PB +${c.proficiencyBonus}',
                            Icons.star_border,
                          ),
                          _badge('${c.speed} ft', Icons.directions_walk),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        const SectionHeader(title: 'Ability Scores', icon: Icons.grid_view),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          child: GridView.count(
            crossAxisCount: 3,
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            mainAxisSpacing: 8,
            crossAxisSpacing: 8,
            childAspectRatio: 1.35,
            children: [
              StatTile(label: 'STR', score: c.str, modifier: c.strMod),
              StatTile(label: 'DEX', score: c.dex, modifier: c.dexMod),
              StatTile(label: 'CON', score: c.con, modifier: c.conMod),
              StatTile(label: 'INT', score: c.int_, modifier: c.intMod),
              StatTile(label: 'WIS', score: c.wis, modifier: c.wisMod),
              StatTile(label: 'CHA', score: c.cha, modifier: c.chaMod),
            ],
          ),
        ),
        const SectionHeader(
          title: 'Saving Throws & Skills',
          subtitle: 'Ability modifiers only',
          icon: Icons.checklist,
        ),
        Card(
          margin: const EdgeInsets.symmetric(horizontal: 12),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              children: [
                // The server payload has no per-save proficiency data, so the
                // sheet shows the base ability modifier and says so rather than
                // inventing proficiency dots or a bonus.
                _saveRow('STR save', c.strMod),
                _saveRow('DEX save', c.dexMod),
                _saveRow('CON save', c.conMod),
                _saveRow('INT save', c.intMod),
                _saveRow('WIS save', c.wisMod),
                _saveRow('CHA save', c.chaMod),
                const Divider(),
                Row(
                  children: [
                    const Icon(
                      Icons.visibility_outlined,
                      size: 16,
                      color: VillumTheme.gold,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'Passive Perception ${c.passivePerception}',
                      style: GoogleFonts.lora(
                        fontSize: 13,
                        color: VillumTheme.parchment,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 6, 16, 0),
          child: Text(
            'Base ability modifiers — save proficiencies are managed in the Villum web app.',
            style: GoogleFonts.lora(
              fontSize: 11,
              color: VillumTheme.parchmentMuted,
            ),
          ),
        ),
      ],
    ),
  );
  Widget _badge(String label, IconData icon) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
    decoration: BoxDecoration(
      color: VillumTheme.surfaceElevated,
      borderRadius: BorderRadius.circular(8),
      border: Border.all(color: VillumTheme.inkLight),
    ),
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 11, color: VillumTheme.gold),
        const SizedBox(width: 4),
        Text(
          label,
          style: GoogleFonts.lora(
            fontSize: 11,
            fontWeight: FontWeight.w600,
            color: VillumTheme.parchment,
          ),
        ),
      ],
    ),
  );
  Widget _saveRow(String label, int mod) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 4),
    child: Row(
      children: [
        const SizedBox(width: 18),
        Expanded(
          child: Text(
            label,
            style: GoogleFonts.lora(fontSize: 13, color: VillumTheme.parchment),
          ),
        ),
        Text(
          mod >= 0 ? '+$mod' : '$mod',
          style: GoogleFonts.lora(
            fontSize: 13,
            fontWeight: FontWeight.w700,
            color: VillumTheme.parchment,
          ),
        ),
      ],
    ),
  );
}

class _CombatTab extends ConsumerWidget {
  final Character c;
  final int characterId;
  const _CombatTab({required this.c, required this.characterId});
  Future<void> _adjustHp(WidgetRef ref, BuildContext context, int delta) async {
    final ctrl = ref.read(characterProvider(characterId).notifier);
    try {
      await ctrl.changeHp(delta);
      HapticFeedback.mediumImpact();
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('HP update failed: $e')));
      }
    }
  }

  Future<void> _rest(WidgetRef ref, BuildContext context, String type) async {
    try {
      final api = ref.read(apiClientProvider);
      await api.post(
        '/api/characters/$characterId/rest',
        data: {'rest_type': type},
      );
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('${type == 'long' ? 'Long' : 'Short'} rest complete'),
          ),
        );
      }
      ref.invalidate(characterProvider(characterId));
      HapticFeedback.heavyImpact();
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Rest failed: $e')));
      }
    }
  }

  /// The server rolls the save; the sheet only shows what it returned.
  Future<void> _rollDeathSave(WidgetRef ref, BuildContext context) async {
    try {
      final api = CharacterApi(ref.read(apiClientProvider));
      final result = await api.deathSave(characterId);
      ref.invalidate(characterProvider(characterId));
      HapticFeedback.mediumImpact();
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(result.summary)));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Death save failed: $e')));
      }
    }
  }

  Future<void> _setExhaustion(
    WidgetRef ref,
    BuildContext context,
    int level,
  ) async {
    final target = level.clamp(0, 6);
    try {
      final api = CharacterApi(ref.read(apiClientProvider));
      await api.setExhaustion(characterId, target);
      ref.invalidate(characterProvider(characterId));
      HapticFeedback.lightImpact();
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Exhaustion set to $target')));
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Exhaustion update failed: $e')));
      }
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final atZero = c.hpCurrent <= 0;
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Card(
            margin: const EdgeInsets.fromLTRB(12, 12, 12, 8),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Row(
                    children: [
                      Text(
                        'Hit Points',
                        style: GoogleFonts.cinzel(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: VillumTheme.parchment,
                        ),
                      ),
                      const Spacer(),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 3,
                        ),
                        decoration: BoxDecoration(
                          color: VillumTheme.surfaceElevated,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          'AC ${c.ac} · Init ${c.initiative >= 0 ? '+${c.initiative}' : '${c.initiative}'}',
                          style: GoogleFonts.lora(
                            fontSize: 11,
                            color: VillumTheme.parchmentMuted,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  HpBar(current: c.hpCurrent, max: c.hpMax, tempHp: c.tempHp),
                  const SizedBox(height: 14),
                  HpStepper(
                    onAdjust: (d, {bool temp = false}) =>
                        _adjustHp(ref, context, d),
                  ),
                  if (atZero) ...[
                    const SizedBox(height: 12),
                    DeathSaveTracker(
                      successes: c.deathSavesSuccesses,
                      failures: c.deathSavesFailures,
                    ),
                    const SizedBox(height: 8),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton.icon(
                        onPressed: () => _rollDeathSave(ref, context),
                        icon: const Icon(Icons.casino_outlined, size: 18),
                        label: const Text('Roll death save'),
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'The server rolls the d20 and records the result.',
                      style: GoogleFonts.lora(
                        fontSize: 11,
                        color: VillumTheme.parchmentMuted,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
          const SectionHeader(title: 'Rest', icon: Icons.bedtime_outlined),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => _rest(ref, context, 'short'),
                    icon: const Icon(Icons.coffee, size: 18),
                    label: const Text('Short Rest'),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: FilledButton.icon(
                    onPressed: () => _rest(ref, context, 'long'),
                    icon: const Icon(Icons.nights_stay, size: 18),
                    label: const Text('Long Rest'),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 4),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              'Short rest recovers some features; long rest restores HP, slots & hit dice.',
              style: GoogleFonts.lora(
                fontSize: 11,
                color: VillumTheme.parchmentMuted,
              ),
            ),
          ),
          const SectionHeader(
            title: 'Conditions',
            icon: Icons.heart_broken_outlined,
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                if (c.concentratingOn.isNotEmpty)
                  ConditionChip(label: 'Concentrating: ${c.concentratingOn}'),
                if (c.hpCurrent == 0) const ConditionChip(label: 'Unconscious'),
                if (c.hpCurrent == 0) const ConditionChip(label: 'Dying'),
                if (c.concentratingOn.isEmpty && c.hpCurrent != 0)
                  Text(
                    'No conditions reported by the server.',
                    style: GoogleFonts.lora(
                      fontSize: 12,
                      color: VillumTheme.parchmentMuted,
                    ),
                  ),
              ],
            ),
          ),
          const SectionHeader(
            title: 'Exhaustion',
            icon: Icons.battery_2_bar_outlined,
          ),
          Card(
            margin: const EdgeInsets.symmetric(horizontal: 12),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Row(
                children: [
                  Text(
                    'Level ${c.exhaustionLevel} / 6',
                    style: GoogleFonts.lora(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: VillumTheme.parchment,
                    ),
                  ),
                  const Spacer(),
                  IconButton(
                    tooltip: 'Lower exhaustion',
                    onPressed: c.exhaustionLevel > 0
                        ? () => _setExhaustion(
                            ref,
                            context,
                            c.exhaustionLevel - 1,
                          )
                        : null,
                    icon: const Icon(Icons.remove_circle_outline),
                  ),
                  IconButton(
                    tooltip: 'Raise exhaustion',
                    onPressed: c.exhaustionLevel < 6
                        ? () => _setExhaustion(
                            ref,
                            context,
                            c.exhaustionLevel + 1,
                          )
                        : null,
                    icon: const Icon(Icons.add_circle_outline),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SpellsTab extends ConsumerWidget {
  final Character c;
  final int characterId;
  const _SpellsTab({required this.c, required this.characterId});

  /// Casts the spell at [spell]'s level, or at [slotLevel] when upcasting.
  /// The server consumes the slot at the spell's own level and answers with
  /// `slot_consumed`; the sheet refetches so the pips show server truth.
  Future<void> _cast(
    WidgetRef ref,
    BuildContext context,
    Spell spell, {
    int? slotLevel,
  }) async {
    try {
      final api = CharacterApi(ref.read(apiClientProvider));
      final result = await api.castSpell(
        characterId,
        spell.id,
        slotLevel: slotLevel,
      );
      ref.invalidate(characterProvider(characterId));
      HapticFeedback.lightImpact();
      if (context.mounted) {
        final consumed = result['slot_consumed'] == true;
        final castLevel = slotLevel ?? spell.level;
        final at = castLevel == 0 ? 'cantrip' : 'level $castLevel';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              consumed
                  ? '${spell.name} cast at $at — slot expended'
                  : '${spell.name} cast at $at',
            ),
          ),
        );
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Cast failed: $e')));
      }
    }
  }

  /// Slot-row action: expends that slot with a spell of the matching level.
  /// Picks a prepared spell first, then any known spell of that level; if the
  /// character knows none it says so instead of posting a spell id of 0.
  Future<void> _castAtSlotLevel(
    WidgetRef ref,
    BuildContext context,
    int level,
  ) async {
    final atLevel = c.spells.where((s) => s.level == level).toList();
    final prepared = atLevel
        .where((s) => s.prepared || s.alwaysPrepared)
        .toList();
    final spell = prepared.isNotEmpty
        ? prepared.first
        : (atLevel.isNotEmpty ? atLevel.first : null);
    if (spell == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'No spell of level $level on this character — add one in the Villum web app.',
          ),
        ),
      );
      return;
    }
    await _cast(ref, context, spell, slotLevel: level);
  }

  /// Lets the player pick the slot to spend: the spell's own level plus any
  /// higher level with a slot still free (upcasting).
  Future<void> _chooseSlotLevel(
    WidgetRef ref,
    BuildContext context,
    Spell spell,
  ) async {
    if (spell.level == 0) {
      await _cast(ref, context, spell);
      return;
    }
    final sc = c.spellcasting;
    final options = <int>[spell.level];
    if (sc != null) {
      for (var lvl = spell.level + 1; lvl <= 9; lvl++) {
        final max = lvl <= sc.slotsMax.length ? sc.slotsMax[lvl - 1] : 0;
        final used = lvl <= sc.slotsUsed.length ? sc.slotsUsed[lvl - 1] : 0;
        if (max > used) options.add(lvl);
      }
    }
    if (options.length == 1) {
      await _cast(ref, context, spell, slotLevel: spell.level);
      return;
    }
    final chosen = await showDialog<int>(
      context: context,
      builder: (dCtx) => SimpleDialog(
        title: Text('Cast ${spell.name}'),
        children: [
          for (final lvl in options)
            SimpleDialogOption(
              onPressed: () => Navigator.pop(dCtx, lvl),
              child: Text(
                lvl == spell.level
                    ? 'Level $lvl (spell level)'
                    : 'Level $lvl (upcast)',
              ),
            ),
        ],
      ),
    );
    if (chosen == null || !context.mounted) return;
    await _cast(ref, context, spell, slotLevel: chosen);
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sc = c.spellcasting;
    final preparedSpells = c.spells
        .where((s) => s.prepared || s.alwaysPrepared)
        .toList();
    final knownSpells = c.spells
        .where((s) => !(s.prepared || s.alwaysPrepared))
        .toList();
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (sc != null)
            Card(
              margin: const EdgeInsets.fromLTRB(12, 12, 12, 6),
              child: Padding(
                padding: const EdgeInsets.all(14),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Row(
                      children: [
                        const Icon(
                          Icons.auto_fix_high,
                          size: 18,
                          color: VillumTheme.gold,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Spellcasting · ${sc.ability.toUpperCase()}',
                          style: GoogleFonts.cinzel(
                            fontSize: 13,
                            fontWeight: FontWeight.w700,
                            color: VillumTheme.parchment,
                          ),
                        ),
                        const Spacer(),
                        Text(
                          'DC ${sc.saveDc} · +${sc.attackBonus} to hit',
                          style: GoogleFonts.lora(
                            fontSize: 11,
                            color: VillumTheme.parchmentMuted,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    for (int lvl = 1; lvl <= 9; lvl++)
                      if (lvl <= sc.slotsMax.length && sc.slotsMax[lvl - 1] > 0)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 10),
                          child: _SlotRow(
                            level: lvl,
                            max: sc.slotsMax[lvl - 1],
                            used: lvl <= sc.slotsUsed.length
                                ? sc.slotsUsed[lvl - 1]
                                : 0,
                            onCast: (l) => _castAtSlotLevel(ref, context, l),
                          ),
                        ),
                    if (sc.slotsMax.every((e) => e == 0))
                      Text(
                        'No spell slots configured.',
                        style: GoogleFonts.lora(
                          fontSize: 12,
                          color: VillumTheme.parchmentMuted,
                        ),
                      ),
                  ],
                ),
              ),
            )
          else
            Padding(
              padding: const EdgeInsets.all(16),
              child: EmptyState(
                icon: Icons.auto_fix_high,
                title: 'No spellcasting',
                subtitle:
                    'This character has no spellcasting class on the server. '
                    'Configure one in the Villum web app.',
              ),
            ),
          if (preparedSpells.isNotEmpty) ...[
            SectionHeader(
              title: 'Prepared Spells · ${preparedSpells.length}',
              icon: Icons.bolt_outlined,
            ),
            ...preparedSpells.map(
              (s) =>
                  _spellTile(s, onTap: () => _chooseSlotLevel(ref, context, s)),
            ),
          ],
          SectionHeader(
            title:
                'Known Spells${knownSpells.isEmpty ? '' : ' · ${knownSpells.length}'}',
            icon: Icons.menu_book_outlined,
          ),
          if (c.spells.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 12),
              child: EmptyState(
                icon: Icons.auto_stories_outlined,
                title: 'No spells',
                subtitle: 'Add spells in the Villum web app.',
              ),
            )
          else if (knownSpells.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Text(
                'All known spells are prepared.',
                style: GoogleFonts.lora(
                  fontSize: 12,
                  color: VillumTheme.parchmentMuted,
                ),
              ),
            )
          else
            ...knownSpells.map(
              (s) =>
                  _spellTile(s, onTap: () => _chooseSlotLevel(ref, context, s)),
            ),
        ],
      ),
    );
  }

  Widget _spellTile(Spell s, {required VoidCallback onTap}) => Card(
    margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
    child: ListTile(
      onTap: onTap,
      leading: Container(
        width: 36,
        height: 36,
        decoration: BoxDecoration(
          color: VillumTheme.gold.withValues(alpha: 0.14),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Center(
          child: Text(
            s.level == 0 ? 'C' : '${s.level}',
            style: GoogleFonts.cinzel(
              fontSize: 13,
              fontWeight: FontWeight.w700,
              color: VillumTheme.gold,
            ),
          ),
        ),
      ),
      title: Text(s.name, style: GoogleFonts.lora(fontWeight: FontWeight.w600)),
      subtitle: Text(
        [s.school, s.castingTime].where((e) => e.isNotEmpty).join(' · '),
        style: GoogleFonts.lora(
          fontSize: 11,
          color: VillumTheme.parchmentMuted,
        ),
      ),
      trailing: s.prepared || s.alwaysPrepared
          ? Container(
              padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
              decoration: BoxDecoration(
                color: VillumTheme.hpHealthy.withValues(alpha: 0.16),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                s.alwaysPrepared && !s.prepared ? 'Always' : 'Prepared',
                style: GoogleFonts.lora(
                  fontSize: 10,
                  fontWeight: FontWeight.w700,
                  color: VillumTheme.hpHealthy,
                ),
              ),
            )
          : const Icon(Icons.chevron_right, size: 18),
    ),
  );
}

/// One spell level's slots: used/total, the server's pip state, and the cast
/// action that expends it.
class _SlotRow extends StatelessWidget {
  final int level;
  final int max;
  final int used;
  final ValueChanged<int> onCast;
  const _SlotRow({
    required this.level,
    required this.max,
    required this.used,
    required this.onCast,
  });

  @override
  Widget build(BuildContext context) {
    final remaining = (max - used).clamp(0, max);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: VillumTheme.gold.withValues(alpha: 0.14),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: VillumTheme.gold.withValues(alpha: 0.25),
                ),
              ),
              child: Text(
                'LVL $level',
                style: GoogleFonts.cinzel(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  color: VillumTheme.gold,
                ),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              '$used / $max used',
              style: GoogleFonts.lora(
                fontSize: 12,
                color: VillumTheme.parchmentMuted,
              ),
            ),
            const Spacer(),
            if (remaining > 0)
              SizedBox(
                height: 32,
                child: OutlinedButton(
                  onPressed: () => onCast(level),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 10),
                    minimumSize: const Size(48, 32),
                    textStyle: GoogleFonts.lora(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  child: const Text('Cast'),
                ),
              ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: List.generate(max, (i) {
            final isUsed = i < used;
            return Semantics(
              label:
                  'Spell slot level $level, ${isUsed ? 'expended' : 'available'}',
              child: Container(
                width: 18,
                height: 18,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isUsed ? Colors.transparent : VillumTheme.gold,
                  border: Border.all(
                    color: isUsed ? VillumTheme.inkLight : VillumTheme.gold,
                    width: 1.4,
                  ),
                ),
                child: isUsed
                    ? const Icon(
                        Icons.close,
                        size: 10,
                        color: VillumTheme.parchmentMuted,
                      )
                    : null,
              ),
            );
          }),
        ),
      ],
    );
  }
}

class _InventoryTab extends ConsumerWidget {
  final Character c;
  final int characterId;
  const _InventoryTab({required this.c, required this.characterId});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final items = c.inventory;
    final cur = c.currency;
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SectionHeader(
            title: 'Currency',
            icon: Icons.monetization_on_outlined,
          ),
          Card(
            margin: const EdgeInsets.symmetric(horizontal: 12),
            child: Padding(
              padding: const EdgeInsets.all(14),
              child: Column(
                children: [
                  if (cur == null)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 6),
                      child: Text(
                        'The server returned no coin for this character.',
                        style: GoogleFonts.lora(
                          fontSize: 12,
                          color: VillumTheme.parchmentMuted,
                        ),
                      ),
                    )
                  else
                    Row(
                      children: [
                        _coin('CP', cur.cp, const Color(0xFFB87333)),
                        _coin('SP', cur.sp, const Color(0xFFC0C0C0)),
                        _coin('EP', cur.ep, const Color(0xFF8CA68E)),
                        _coin('GP', cur.gp, VillumTheme.gold),
                        _coin('PP', cur.pp, const Color(0xFFE5E4E2)),
                      ],
                    ),
                  const SizedBox(height: 10),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: cur == null
                          ? null
                          : () => _editCurrency(context, ref),
                      icon: const Icon(Icons.edit, size: 16),
                      label: const Text('Edit coin'),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SectionHeader(
            title: 'Inventory',
            icon: Icons.backpack_outlined,
          ),
          if (items.isNotEmpty)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 4),
              child: Text(
                '${items.length} item${items.length == 1 ? '' : 's'} · '
                '${_totalWeight(items).toStringAsFixed(1)} lb carried',
                style: GoogleFonts.lora(
                  fontSize: 11,
                  color: VillumTheme.parchmentMuted,
                ),
              ),
            ),
          if (items.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 12),
              child: EmptyState(
                icon: Icons.inventory_2_outlined,
                title: 'Empty pack',
                subtitle: 'No items on the server yet — add one below.',
              ),
            ),
          ...items.map((it) => _itemTile(context, ref, it)),
          Padding(
            padding: const EdgeInsets.all(12),
            child: OutlinedButton.icon(
              onPressed: () => _editItem(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('Add item'),
            ),
          ),
        ],
      ),
    );
  }

  double _totalWeight(List<InventoryItem> items) =>
      items.fold(0, (sum, it) => sum + it.weight * it.quantity);

  Widget _itemTile(
    BuildContext context,
    WidgetRef ref,
    InventoryItem it,
  ) => Card(
    margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
    child: ListTile(
      onTap: () => _editItem(context, ref, item: it),
      leading: Container(
        width: 36,
        height: 36,
        decoration: BoxDecoration(
          color: it.isEquipped
              ? VillumTheme.gold.withValues(alpha: 0.18)
              : VillumTheme.surfaceElevated,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: it.isEquipped
                ? VillumTheme.gold.withValues(alpha: 0.4)
                : VillumTheme.inkLight,
          ),
        ),
        child: Icon(
          it.isEquipped ? Icons.check : Icons.inventory_2_outlined,
          size: 18,
          color: it.isEquipped ? VillumTheme.gold : VillumTheme.parchmentMuted,
        ),
      ),
      title: Text(
        '${it.name} ×${it.quantity}',
        style: GoogleFonts.lora(fontWeight: FontWeight.w600),
      ),
      subtitle: Text(
        [
          if (it.isEquipped) 'Equipped',
          it.category,
          it.description,
        ].where((e) => e.isNotEmpty).join(' · '),
        maxLines: 2,
        overflow: TextOverflow.ellipsis,
        style: GoogleFonts.lora(
          fontSize: 11,
          color: VillumTheme.parchmentMuted,
        ),
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '${it.weight} lb',
            style: GoogleFonts.lora(
              fontSize: 11,
              color: VillumTheme.parchmentMuted,
            ),
          ),
          IconButton(
            tooltip: it.isEquipped ? 'Unequip' : 'Equip',
            onPressed: () =>
                _setEquipped(context, ref, it, equipped: !it.isEquipped),
            icon: Icon(
              it.isEquipped ? Icons.check_circle : Icons.radio_button_unchecked,
              color: it.isEquipped
                  ? VillumTheme.gold
                  : VillumTheme.parchmentMuted,
            ),
          ),
        ],
      ),
    ),
  );

  /// Equip/unequip goes through the full-item PUT, so the server's own copy is
  /// flipped and every other field is carried over untouched.
  Future<void> _setEquipped(
    BuildContext context,
    WidgetRef ref,
    InventoryItem it, {
    required bool equipped,
  }) async {
    final api = CharacterApi(ref.read(apiClientProvider));
    try {
      await api.updateInventoryItem(
        it.id,
        InventoryItem(
          id: it.id,
          name: it.name,
          quantity: it.quantity,
          weight: it.weight,
          category: it.category,
          damageDice: it.damageDice,
          damageType: it.damageType,
          weaponProperties: it.weaponProperties,
          acBonus: it.acBonus,
          armorType: it.armorType,
          description: it.description,
          isEquipped: equipped,
          isMagical: it.isMagical,
          attunement: it.attunement,
          isIdentified: it.isIdentified,
          notes: it.notes,
          attackAbility: it.attackAbility,
          attackBonus: it.attackBonus,
        ),
      );
      ref.invalidate(characterProvider(characterId));
      HapticFeedback.lightImpact();
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Could not update item: $e')));
      }
    }
  }

  /// Add ([item] null) or edit an item. Persists through the inventory
  /// endpoints and then refetches, so quantity/equipped/encumbrance all come
  /// from the server.
  Future<void> _editItem(
    BuildContext context,
    WidgetRef ref, {
    InventoryItem? item,
  }) async {
    final api = CharacterApi(ref.read(apiClientProvider));

    // The dialog owns its controllers and disposes them in State.dispose, so
    // disposal happens after the route's reverse transition. Disposing here on
    // pop rebuilds the TextFields against dead controllers ("used after being
    // disposed").
    final result = await showDialog<_ItemDialogResult>(
      context: context,
      builder: (dCtx) => _ItemDialog(item: item),
    );
    if (result == null) return;
    final action = result.action;
    final name = result.name;
    final quantity = result.quantity;
    final weight = result.weight;
    final category = result.category;
    final equipped = result.equipped;

    if (action == 'delete' && item != null) {
      try {
        await api.deleteInventoryItem(item.id);
        ref.invalidate(characterProvider(characterId));
        HapticFeedback.mediumImpact();
        if (context.mounted) {
          ScaffoldMessenger.of(context)
              .showSnackBar(SnackBar(content: Text('${item.name} removed')));
        }
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context)
              .showSnackBar(SnackBar(content: Text('Delete failed: $e')));
        }
      }
      return;
    }
    if (action != 'save') return;
    if (name.isEmpty) {
      if (context.mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('An item needs a name.')));
      }
      return;
    }

    try {
      if (item == null) {
        await api.addInventoryItem(characterId, {
          'name': name,
          'quantity': quantity,
          'weight': weight,
          'category': category,
          'is_equipped': equipped,
        });
      } else {
        await api.updateInventoryItem(
          item.id,
          InventoryItem(
            id: item.id,
            name: name,
            quantity: quantity,
            weight: weight,
            category: category,
            damageDice: item.damageDice,
            damageType: item.damageType,
            weaponProperties: item.weaponProperties,
            acBonus: item.acBonus,
            armorType: item.armorType,
            description: item.description,
            isEquipped: equipped,
            isMagical: item.isMagical,
            attunement: item.attunement,
            isIdentified: item.isIdentified,
            notes: item.notes,
            attackAbility: item.attackAbility,
            attackBonus: item.attackBonus,
          ),
        );
      }
      ref.invalidate(characterProvider(characterId));
      HapticFeedback.lightImpact();
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(item == null ? 'Item added' : 'Item saved')),
        );
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Save failed: $e')));
      }
    }
  }

  Widget _coin(String label, int value, Color color) => Expanded(
    child: Column(
      children: [
        Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: color.withValues(alpha: 0.18),
            border: Border.all(color: color.withValues(alpha: 0.5)),
          ),
          child: Center(
            child: Text(
              label,
              style: GoogleFonts.cinzel(
                fontSize: 10,
                fontWeight: FontWeight.w800,
                color: color,
              ),
            ),
          ),
        ),
        const SizedBox(height: 4),
        Text(
          '$value',
          style: GoogleFonts.lora(
            fontSize: 13,
            fontWeight: FontWeight.w700,
            color: VillumTheme.parchment,
          ),
        ),
      ],
    ),
  );

  /// Refetches the character first so the dialog is seeded with the server's
  /// current coin, then submits the complete five-coin map. Editing one coin
  /// can therefore never zero the others.
  Future<void> _editCurrency(BuildContext context, WidgetRef ref) async {
    Currency? cur;
    try {
      final fresh = await ref.refresh(characterProvider(characterId).future);
      if (!context.mounted) return;
      cur = fresh.currency;
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Could not load coin: $e')));
      }
      return;
    }
    final seeded = cur;
    if (seeded == null) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('The server returned no coin for this character.'),
          ),
        );
      }
      return;
    }
    var saved = false;
    await showDialog<void>(
      context: context,
      builder: (dCtx) => _CurrencyDialog(
        currency: seeded,
        onSave: (updated) async {
          try {
            final ctrl = ref.read(characterProvider(characterId).notifier);
            await ctrl.updateCurrency(updated);
            ref.invalidate(characterProvider(characterId));
            saved = true;
            return true;
          } catch (e) {
            if (context.mounted) {
              ScaffoldMessenger.of(context)
                  .showSnackBar(SnackBar(content: Text('Failed: $e')));
            }
            return false;
          }
        },
      ),
    );
    if (saved && context.mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('Coin updated')));
    }
  }
}

class _FeaturesTab extends StatelessWidget {
  final Character c;
  const _FeaturesTab({required this.c});
  @override
  Widget build(BuildContext context) => SingleChildScrollView(
    padding: const EdgeInsets.only(bottom: 24),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SectionHeader(
          title: 'Features & Traits',
          icon: Icons.auto_awesome_outlined,
        ),
        if (c.features.isEmpty)
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 12),
            child: EmptyState(
              icon: Icons.emoji_events_outlined,
              title: 'No features',
              subtitle: 'Racial traits and class features appear here.',
            ),
          )
        else
          ...c.features.map(
            (f) => Card(
              margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              child: ExpansionTile(
                title: Text(
                  f.name,
                  style: GoogleFonts.lora(fontWeight: FontWeight.w600),
                ),
                subtitle: Text(
                  '${f.source} · Lv ${f.levelGained}',
                  style: GoogleFonts.lora(
                    fontSize: 11,
                    color: VillumTheme.parchmentMuted,
                  ),
                ),
                children: [
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
                    child: Text(
                      f.description,
                      style: GoogleFonts.lora(
                        fontSize: 13,
                        color: VillumTheme.parchmentMuted,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        const SectionHeader(title: 'Notes', icon: Icons.notes_outlined),
        Card(
          margin: const EdgeInsets.symmetric(horizontal: 12),
          child: Padding(
            padding: const EdgeInsets.all(14),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (c.background.isNotEmpty)
                  Text(
                    c.background,
                    style: GoogleFonts.cinzel(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: VillumTheme.gold,
                    ),
                  ),
                if (c.background.isNotEmpty) const SizedBox(height: 6),
                Text(
                  c.alignment.isNotEmpty
                      ? 'Alignment: ${c.alignment}'
                      : 'No notes yet.',
                  style: GoogleFonts.lora(
                    fontSize: 13,
                    color: VillumTheme.parchmentMuted,
                  ),
                ),
                const SizedBox(height: 10),
                OutlinedButton.icon(
                  onPressed: () => ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Edit notes in the Villum web app.'),
                    ),
                  ),
                  icon: const Icon(Icons.edit_outlined, size: 16),
                  label: const Text('Edit'),
                ),
              ],
            ),
          ),
        ),
      ],
    ),
  );
}

/// The values the add/edit-item dialog hands back to [_CharacterSheetScreenState].
class _ItemDialogResult {
  final String action;
  final String name;
  final int quantity;
  final double weight;
  final String category;
  final bool equipped;

  const _ItemDialogResult({
    required this.action,
    required this.name,
    required this.quantity,
    required this.weight,
    required this.category,
    required this.equipped,
  });
}

/// Owns the add/edit-item form state, including its controllers, so they are
/// disposed only once the dialog route (and its reverse transition) is gone.
class _ItemDialog extends StatefulWidget {
  final InventoryItem? item;
  const _ItemDialog({this.item});

  @override
  State<_ItemDialog> createState() => _ItemDialogState();
}

class _ItemDialogState extends State<_ItemDialog> {
  late final TextEditingController _nameCtrl;
  late final TextEditingController _qtyCtrl;
  late final TextEditingController _weightCtrl;
  late final TextEditingController _categoryCtrl;
  late bool _equipped;

  @override
  void initState() {
    super.initState();
    final item = widget.item;
    _nameCtrl = TextEditingController(text: item?.name ?? '');
    _qtyCtrl = TextEditingController(text: '${item?.quantity ?? 1}');
    _weightCtrl = TextEditingController(text: '${item?.weight ?? 0}');
    _categoryCtrl = TextEditingController(text: item?.category ?? '');
    _equipped = item?.isEquipped ?? false;
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    _qtyCtrl.dispose();
    _weightCtrl.dispose();
    _categoryCtrl.dispose();
    super.dispose();
  }

  void _pop(String action) => Navigator.pop(
    context,
    _ItemDialogResult(
      action: action,
      name: _nameCtrl.text.trim(),
      quantity: int.tryParse(_qtyCtrl.text.trim()) ?? 1,
      weight: double.tryParse(_weightCtrl.text.trim()) ?? 0,
      category: _categoryCtrl.text.trim(),
      equipped: _equipped,
    ),
  );

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text(widget.item == null ? 'Add item' : 'Edit item'),
    content: SingleChildScrollView(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: _nameCtrl,
            decoration: const InputDecoration(labelText: 'Name'),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _qtyCtrl,
            keyboardType: TextInputType.number,
            decoration: const InputDecoration(labelText: 'Quantity'),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _weightCtrl,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: const InputDecoration(labelText: 'Weight (lb each)'),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _categoryCtrl,
            decoration: const InputDecoration(labelText: 'Category'),
          ),
          SwitchListTile(
            contentPadding: EdgeInsets.zero,
            title: const Text('Equipped'),
            value: _equipped,
            onChanged: (v) => setState(() => _equipped = v),
          ),
        ],
      ),
    ),
    actions: [
      if (widget.item != null)
        TextButton(
          onPressed: () => _pop('delete'),
          child: const Text('Delete'),
        ),
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(onPressed: () => _pop('save'), child: const Text('Save')),
    ],
  );
}

/// Owns the five coin controllers and saves through [onSave], popping only when
/// the save reports success.
class _CurrencyDialog extends StatefulWidget {
  final Currency currency;
  final Future<bool> Function(Currency) onSave;
  const _CurrencyDialog({required this.currency, required this.onSave});

  @override
  State<_CurrencyDialog> createState() => _CurrencyDialogState();
}

class _CurrencyDialogState extends State<_CurrencyDialog> {
  late final Map<String, TextEditingController> _ctrls;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final c = widget.currency;
    _ctrls = {
      'cp': TextEditingController(text: '${c.cp}'),
      'sp': TextEditingController(text: '${c.sp}'),
      'ep': TextEditingController(text: '${c.ep}'),
      'gp': TextEditingController(text: '${c.gp}'),
      'pp': TextEditingController(text: '${c.pp}'),
    };
  }

  @override
  void dispose() {
    for (final ctrl in _ctrls.values) {
      ctrl.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    final vals = _ctrls.map((k, v) => MapEntry(k, int.tryParse(v.text) ?? 0));
    setState(() => _saving = true);
    final ok = await widget.onSave(
      Currency(
        cp: vals['cp']!,
        sp: vals['sp']!,
        ep: vals['ep']!,
        gp: vals['gp']!,
        pp: vals['pp']!,
      ),
    );
    if (!mounted) return;
    if (!ok) {
      setState(() => _saving = false);
      return;
    }
    Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Edit currency'),
    content: Column(
      mainAxisSize: MainAxisSize.min,
      children: _ctrls.entries
          .map(
            (e) => Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: TextField(
                controller: e.value,
                keyboardType: TextInputType.number,
                decoration: InputDecoration(labelText: e.key.toUpperCase()),
              ),
            ),
          )
          .toList(),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(
        onPressed: _saving ? null : _save,
        child: const Text('Save'),
      ),
    ],
  );
}
