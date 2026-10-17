import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/providers.dart';
import '../../data/models/character.dart';
import '../../data/models/currency.dart';
import '../shared/hp_bar.dart';
import '../shared/section_header.dart';
import '../shared/stat_block.dart';
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
          subtitle: 'Proficiency +${2}',
          icon: Icons.checklist,
        ),
        Card(
          margin: const EdgeInsets.symmetric(horizontal: 12),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              children: [
                _saveRow('STR save', c.strMod + c.proficiencyBonus, true),
                _saveRow('DEX save', c.dexMod + c.proficiencyBonus, false),
                _saveRow('CON save', c.conMod + c.proficiencyBonus, false),
                _saveRow('INT save', c.intMod, false),
                _saveRow('WIS save', c.wisMod + c.proficiencyBonus, true),
                _saveRow('CHA save', c.chaMod, false),
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
  Widget _saveRow(String label, int mod, bool prof) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 4),
    child: Row(
      children: [
        Container(
          width: 8,
          height: 8,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: prof ? VillumTheme.gold : Colors.transparent,
            border: Border.all(color: VillumTheme.gold),
          ),
        ),
        const SizedBox(width: 10),
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
                const ConditionChip(label: 'Tap to add…', active: false),
              ],
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
  Future<void> _cast(WidgetRef ref, BuildContext context, int level) async {
    try {
      final api = ref.read(apiClientProvider);
      // Find a spell of that level to cast, else just hit cast-spell with slot_level
      final spell = c.spells.where((s) => s.level == level).firstOrNull;
      await api.post(
        '/api/characters/$characterId/cast-spell',
        data: {'spell_id': spell?.id ?? 0, 'slot_level': level},
      );
      ref.invalidate(characterProvider(characterId));
      HapticFeedback.lightImpact();
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Cast failed: $e')));
      }
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sc = c.spellcasting;
    final spells = c.spells;
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
                          child: SlotPips(
                            level: lvl,
                            max: sc.slotsMax[lvl - 1],
                            used: lvl <= sc.slotsUsed.length
                                ? sc.slotsUsed[lvl - 1]
                                : 0,
                            onExpend: (l) => _cast(ref, context, l),
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
              child: Text(
                'No spellcasting — add a spellcasting class in the Villum web app.',
                style: GoogleFonts.lora(
                  fontSize: 12,
                  color: VillumTheme.parchmentMuted,
                ),
              ),
            ),
          const SectionHeader(
            title: 'Known Spells',
            icon: Icons.menu_book_outlined,
          ),
          if (spells.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 12),
              child: EmptyState(
                icon: Icons.auto_stories_outlined,
                title: 'No spells',
                subtitle: 'Add spells in the Villum web app.',
              ),
            )
          else
            ...spells.map(
              (s) => Card(
                margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                child: ListTile(
                  leading: Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: VillumTheme.gold.withValues(alpha: 0.14),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Center(
                      child: Text(
                        '${s.level}',
                        style: GoogleFonts.cinzel(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: VillumTheme.gold,
                        ),
                      ),
                    ),
                  ),
                  title: Text(
                    s.name,
                    style: GoogleFonts.lora(fontWeight: FontWeight.w600),
                  ),
                  subtitle: Text(
                    '${s.school} · ${s.castingTime}',
                    style: GoogleFonts.lora(
                      fontSize: 11,
                      color: VillumTheme.parchmentMuted,
                    ),
                  ),
                  trailing: s.prepared
                      ? Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 7,
                            vertical: 3,
                          ),
                          decoration: BoxDecoration(
                            color: VillumTheme.hpHealthy.withValues(
                              alpha: 0.16,
                            ),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            'Prepared',
                            style: GoogleFonts.lora(
                              fontSize: 10,
                              fontWeight: FontWeight.w700,
                              color: VillumTheme.hpHealthy,
                            ),
                          ),
                        )
                      : null,
                ),
              ),
            ),
        ],
      ),
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
                  Row(
                    children: [
                      _coin('CP', cur?.cp ?? 0, const Color(0xFFB87333)),
                      _coin('SP', cur?.sp ?? 0, const Color(0xFFC0C0C0)),
                      _coin('EP', cur?.ep ?? 0, const Color(0xFF8CA68E)),
                      _coin('GP', cur?.gp ?? 0, VillumTheme.gold),
                      _coin('PP', cur?.pp ?? 0, const Color(0xFFE5E4E2)),
                    ],
                  ),
                  const SizedBox(height: 10),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: () => _editCurrency(context, ref),
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
          if (items.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 12),
              child: EmptyState(
                icon: Icons.inventory_2_outlined,
                title: 'Empty pack',
                subtitle:
                    'Your inventory is empty. Add gear in the Villum web app.',
              ),
            ),
          ...items.map(
            (it) => Card(
              margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              child: ListTile(
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
                    color: it.isEquipped
                        ? VillumTheme.gold
                        : VillumTheme.parchmentMuted,
                  ),
                ),
                title: Text(
                  '${it.name} ×${it.quantity}',
                  style: GoogleFonts.lora(fontWeight: FontWeight.w600),
                ),
                subtitle: Text(
                  [
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
                trailing: Text(
                  '${it.weight} lb',
                  style: GoogleFonts.lora(
                    fontSize: 11,
                    color: VillumTheme.parchmentMuted,
                  ),
                ),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(12),
            child: OutlinedButton.icon(
              onPressed: () => ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text('Add items in the Villum web app.'),
                ),
              ),
              icon: const Icon(Icons.add),
              label: const Text('Add item'),
            ),
          ),
        ],
      ),
    );
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
  void _editCurrency(BuildContext context, WidgetRef ref) {
    final cur = c.currency ?? const Currency(cp: 0, sp: 0, ep: 0, gp: 0, pp: 0);
    final ctrls = {
      'cp': TextEditingController(text: '${cur.cp}'),
      'sp': TextEditingController(text: '${cur.sp}'),
      'ep': TextEditingController(text: '${cur.ep}'),
      'gp': TextEditingController(text: '${cur.gp}'),
      'pp': TextEditingController(text: '${cur.pp}'),
    };
    showDialog(
      context: context,
      builder: (dCtx) => AlertDialog(
        title: const Text('Edit currency'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: ctrls.entries
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
            onPressed: () => Navigator.pop(dCtx),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () async {
              final vals = ctrls.map(
                (k, v) => MapEntry(k, int.tryParse(v.text) ?? 0),
              );
              try {
                final ctrl = ref.read(characterProvider(characterId).notifier);
                await ctrl.updateCurrency(
                  Currency(
                    cp: vals['cp']!,
                    sp: vals['sp']!,
                    ep: vals['ep']!,
                    gp: vals['gp']!,
                    pp: vals['pp']!,
                  ),
                );
                if (dCtx.mounted) Navigator.pop(dCtx);
                if (context.mounted) {
                  ScaffoldMessenger.of(
                    context,
                  ).showSnackBar(const SnackBar(content: Text('Coin updated')));
                }
              } catch (e) {
                if (context.mounted) {
                  ScaffoldMessenger.of(context)
                      .showSnackBar(SnackBar(content: Text('Failed: $e')));
                }
              }
            },
            child: const Text('Save'),
          ),
        ],
      ),
    );
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
