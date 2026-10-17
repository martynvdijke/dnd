import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/auth_repository.dart';
import '../../core/providers.dart';
import '../shared/states.dart';
import '../shared/section_header.dart';

class DmToolsScreen extends ConsumerStatefulWidget {
  final int? campaignId;
  const DmToolsScreen({super.key, this.campaignId});

  @override
  ConsumerState<DmToolsScreen> createState() => _DmToolsScreenState();
}

class _DmToolsScreenState extends ConsumerState<DmToolsScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tab;
  int? _campaignId;

  @override
  void initState() {
    super.initState();
    _tab = TabController(length: 3, vsync: this);
    _campaignId = widget.campaignId;
  }

  @override
  void dispose() {
    _tab.dispose();
    super.dispose();
  }

  bool _isDm(AuthState authState) =>
      authState is AuthAuthenticated && authState.user.isDm;

  @override
  Widget build(BuildContext context) {
    final auth = ref.watch(authControllerProvider);
    final isDm = _isDm(auth);

    if (!isDm) {
      return Scaffold(
        appBar: AppBar(
          title: Text(
            'DM Tools',
            style: GoogleFonts.cinzel(fontWeight: FontWeight.w700),
          ),
        ),
        body: const PermissionState(
          title: 'Dungeon Master only',
          message: 'The DM Screen, generators and AI Draft Assistant are only available to DMs and admins. If you are the DM, log in with a DM account.',
          icon: Icons.shield_outlined,
        ),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(
          'DM Tools',
          style: GoogleFonts.cinzel(fontWeight: FontWeight.w700),
        ),
        bottom: TabBar(
          controller: _tab,
          labelColor: VillumTheme.gold,
          unselectedLabelColor: VillumTheme.parchmentMuted,
          indicatorColor: VillumTheme.gold,
          labelStyle: GoogleFonts.cinzel(
            fontSize: 11,
            fontWeight: FontWeight.w800,
            letterSpacing: 0.6,
          ),
          tabs: const [
            Tab(text: 'SCREEN'),
            Tab(text: 'GENERATORS'),
            Tab(text: 'AI DRAFT'),
          ],
        ),
      ),
      body: Column(
        children: [
          _campaignPicker(),
          Expanded(
            child: TabBarView(
              controller: _tab,
              children: [
                _DmScreenTab(campaignId: _campaignId),
                _GeneratorsTab(campaignId: _campaignId),
                _AiDraftTab(campaignId: _campaignId),
              ],
            ),
          ),
        ],
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
        final items = list.cast<dynamic>();
        if (items.isEmpty) {
          return Padding(
            padding: const EdgeInsets.all(12),
            child: Text(
              'No campaigns.',
              style: GoogleFonts.lora(
                fontSize: 12,
                color: VillumTheme.parchmentMuted,
              ),
            ),
          );
        }
        return Container(
          color: VillumTheme.surfaceDark,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: DropdownButtonFormField<int>(
            key: ValueKey(_campaignId),
            initialValue:
                _campaignId != null &&
                    items.any((e) {
                      final id = (e as dynamic).id ?? (e as Map)['id'];
                      return id == _campaignId;
                    })
                ? _campaignId
                : null,
            hint: const Text('Select campaign'),
            decoration: const InputDecoration(
              labelText: 'Campaign',
              prefixIcon: Icon(Icons.flag),
            ),
            items: items.map((e) {
              final id = (e as dynamic).id ?? (e as Map)['id'] as int;
              final name =
                  (e as dynamic).name ?? (e as Map)['name'] ?? 'Campaign $id';
              return DropdownMenuItem<int>(
                value: id,
                child: Text(name.toString()),
              );
            }).toList(),
            onChanged: (v) => setState(() => _campaignId = v),
          ),
        );
      },
    );
  }
}

// ------------------------------------------------------------------
// DM Screen tab
// ------------------------------------------------------------------
class _DmScreenTab extends ConsumerWidget {
  final int? campaignId;
  const _DmScreenTab({this.campaignId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (campaignId == null) {
      return const EmptyState(
        icon: Icons.dashboard_outlined,
        title: 'Select a campaign',
        subtitle: 'Choose a campaign above to load the DM screen.',
      );
    }
    final async = ref.watch(dmScreenProvider(campaignId!));
    return async.when(
      loading: () => const LoadingState(message: 'Loading DM screen…'),
      error: (e, _) => ErrorState(
        message: e.toString(),
        onRetry: () => ref.invalidate(dmScreenProvider(campaignId!)),
      ),
      data: (dm) {
        return SingleChildScrollView(
          padding: const EdgeInsets.only(bottom: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Combat state
              Card(
                margin: const EdgeInsets.fromLTRB(12, 12, 12, 6),
                child: Padding(
                  padding: const EdgeInsets.all(14),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 3,
                            ),
                            decoration: BoxDecoration(
                              color: dm.combat.active
                                  ? VillumTheme.hpHealthy.withValues(
                                      alpha: 0.18,
                                    )
                                  : VillumTheme.surfaceElevated,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Text(
                              dm.combat.active
                                  ? 'COMBAT · TURN ${dm.combat.currentTurn}'
                                  : 'NO COMBAT',
                              style: GoogleFonts.cinzel(
                                fontSize: 10,
                                fontWeight: FontWeight.w800,
                                color: dm.combat.active
                                    ? VillumTheme.hpHealthy
                                    : VillumTheme.parchmentMuted,
                              ),
                            ),
                          ),
                          const Spacer(),
                          if (dm.ambience.track.isNotEmpty)
                            Row(
                              children: [
                                const Icon(
                                  Icons.music_note,
                                  size: 14,
                                  color: VillumTheme.gold,
                                ),
                                const SizedBox(width: 4),
                                Text(
                                  dm.ambience.track,
                                  style: GoogleFonts.lora(
                                    fontSize: 11,
                                    color: VillumTheme.parchmentMuted,
                                  ),
                                ),
                              ],
                            ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      if (dm.combat.entries.isEmpty)
                        Text(
                          'No active combat.',
                          style: GoogleFonts.lora(
                            fontSize: 12,
                            color: VillumTheme.parchmentMuted,
                          ),
                        )
                      else
                        ...dm.combat.entries.map((e) {
                          final isActive =
                              e['is_active'] == true || e['isActive'] == true;
                          final init =
                              e['initiative_roll'] ??
                              e['initiativeRoll'] ??
                              '-';
                          final name = e['name'] ?? '';
                          final hpC = e['hp_current'] ?? e['hpCurrent'] ?? '?';
                          final hpM = e['hp_max'] ?? e['hpMax'] ?? '?';
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 6),
                            child: Row(
                              children: [
                                Container(
                                  width: 28,
                                  height: 28,
                                  decoration: BoxDecoration(
                                    color: isActive
                                        ? VillumTheme.gold
                                        : VillumTheme.surfaceElevated,
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: Center(
                                    child: Text(
                                      '$init',
                                      style: GoogleFonts.cinzel(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w800,
                                        color: isActive
                                            ? VillumTheme.ink
                                            : VillumTheme.parchment,
                                      ),
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    '$name',
                                    style: GoogleFonts.lora(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                ),
                                Text(
                                  'HP $hpC/$hpM',
                                  style: GoogleFonts.lora(
                                    fontSize: 11,
                                    color: VillumTheme.parchmentMuted,
                                  ),
                                ),
                                if (isActive) ...[
                                  const SizedBox(width: 6),
                                  const Icon(
                                    Icons.play_arrow,
                                    size: 16,
                                    color: VillumTheme.gold,
                                  ),
                                ],
                              ],
                            ),
                          );
                        }),
                    ],
                  ),
                ),
              ),
              const SectionHeader(title: 'Party', icon: Icons.group_outlined),
              if (dm.party.isEmpty)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  child: Text(
                    'No party members in this campaign.',
                    style: GoogleFonts.lora(
                      fontSize: 12,
                      color: VillumTheme.parchmentMuted,
                    ),
                  ),
                )
              else
                ...dm.party.map((p) {
                  final ratio = p.hpMax > 0 ? p.hpCurrent / p.hpMax : 1.0;
                  final c = VillumTheme.hpColor(ratio);
                  return Card(
                    margin: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 4,
                    ),
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Row(
                        children: [
                          Container(
                            width: 36,
                            height: 36,
                            decoration: BoxDecoration(
                              color: c.withValues(alpha: 0.18),
                              shape: BoxShape.circle,
                              border: Border.all(
                                color: c.withValues(alpha: 0.4),
                              ),
                            ),
                            child: Center(
                              child: Text(
                                p.name.isNotEmpty ? p.name[0] : '?',
                                style: GoogleFonts.cinzel(
                                  fontWeight: FontWeight.w700,
                                  color: c,
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
                                  p.name,
                                  style: GoogleFonts.lora(
                                    fontWeight: FontWeight.w700,
                                  ),
                                ),
                                Text(
                                  'HP ${p.hpCurrent}/${p.hpMax} · AC ${p.ac}',
                                  style: GoogleFonts.lora(
                                    fontSize: 11,
                                    color: VillumTheme.parchmentMuted,
                                  ),
                                ),
                                if (p.conditions.isNotEmpty)
                                  Padding(
                                    padding: const EdgeInsets.only(top: 4),
                                    child: Wrap(
                                      spacing: 6,
                                      children: p.conditions
                                          .map(
                                            (cond) => Container(
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    horizontal: 6,
                                                    vertical: 2,
                                                  ),
                                              decoration: BoxDecoration(
                                                color: VillumTheme.hpDying
                                                    .withValues(alpha: 0.14),
                                                borderRadius:
                                                    BorderRadius.circular(8),
                                              ),
                                              child: Text(
                                                cond,
                                                style: GoogleFonts.lora(
                                                  fontSize: 10,
                                                  color: VillumTheme.hpDying,
                                                ),
                                              ),
                                            ),
                                          )
                                          .toList(),
                                    ),
                                  ),
                              ],
                            ),
                          ),
                          Container(
                            width: 44,
                            height: 6,
                            decoration: BoxDecoration(
                              color: VillumTheme.inkMid,
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: FractionallySizedBox(
                              alignment: Alignment.centerLeft,
                              widthFactor: ratio.clamp(0.0, 1.0),
                              child: Container(
                                decoration: BoxDecoration(
                                  color: c,
                                  borderRadius: BorderRadius.circular(6),
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                }),
              const SectionHeader(
                title: 'Ambience',
                icon: Icons.surround_sound_outlined,
              ),
              Card(
                margin: const EdgeInsets.symmetric(horizontal: 12),
                child: Padding(
                  padding: const EdgeInsets.all(14),
                  child: Row(
                    children: [
                      const Icon(Icons.music_note, color: VillumTheme.gold),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              dm.ambience.track.isNotEmpty
                                  ? dm.ambience.track
                                  : 'No track',
                              style: GoogleFonts.lora(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            Text(
                              dm.ambience.action.isNotEmpty
                                  ? dm.ambience.action
                                  : 'Idle',
                              style: GoogleFonts.lora(
                                fontSize: 11,
                                color: VillumTheme.parchmentMuted,
                              ),
                            ),
                          ],
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.play_arrow),
                        tooltip: 'Play',
                        onPressed: () => ScaffoldMessenger.of(context)
                            .showSnackBar(
                              const SnackBar(
                                content: Text(
                                  'Ambience controls live in the Villum web app.',
                                ),
                              ),
                            ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

// ------------------------------------------------------------------
// Generators tab
// ------------------------------------------------------------------
class _GeneratorsTab extends ConsumerStatefulWidget {
  final int? campaignId;
  const _GeneratorsTab({this.campaignId});
  @override
  ConsumerState<_GeneratorsTab> createState() => _GeneratorsTabState();
}

class _GeneratorsTabState extends ConsumerState<_GeneratorsTab> {
  String _selected = 'npc';
  String? _result;
  bool _loading = false;

  static const _kinds = [
    'npc',
    'name',
    'encounter',
    'loot',
    'treasure',
    'tavern',
    'weather',
    'adventure-hook',
    'dungeon-dressing',
    'urban-encounter',
    'road-encounter',
  ];

  Future<void> _roll() async {
    setState(() => _loading = true);
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.get('/api/generate/$_selected');
      final data = resp.data;
      String text;
      if (data is String) {
        text = data;
      } else if (data is Map && data['result'] != null) {
        text = data['result'].toString();
      } else if (data is Map && data['text'] != null) {
        text = data['text'].toString();
      } else {
        text = data.toString();
      }
      if (!mounted) return;
      setState(() => _result = text);
      HapticFeedback.lightImpact();
    } catch (e) {
      if (!mounted) return;
      setState(() => _result = 'Failed: $e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SectionHeader(
            title: 'Random Generators',
            subtitle: 'One-tap roll & show',
            icon: Icons.casino_outlined,
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Wrap(
              spacing: 8,
              runSpacing: 8,
              children: _kinds.map((k) {
                final sel = _selected == k;
                return ChoiceChip(
                  label: Text(k),
                  selected: sel,
                  onSelected: (_) => setState(() => _selected = k),
                  selectedColor: VillumTheme.gold.withValues(alpha: 0.22),
                );
              }).toList(),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(12),
            child: SizedBox(
              height: 48,
              child: FilledButton.icon(
                onPressed: _loading ? null : _roll,
                icon: _loading
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: VillumTheme.ink,
                        ),
                      )
                    : const Icon(Icons.casino),
                label: Text(_loading ? 'Rolling…' : 'Roll $_selected'),
              ),
            ),
          ),
          if (_result != null)
            Card(
              margin: const EdgeInsets.symmetric(horizontal: 12),
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(
                          Icons.auto_awesome,
                          size: 16,
                          color: VillumTheme.gold,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          _selected.toUpperCase(),
                          style: GoogleFonts.cinzel(
                            fontSize: 11,
                            fontWeight: FontWeight.w800,
                            color: VillumTheme.gold,
                          ),
                        ),
                        const Spacer(),
                        IconButton(
                          icon: const Icon(Icons.copy, size: 18),
                          tooltip: 'Copy',
                          onPressed: () {
                            Clipboard.setData(ClipboardData(text: _result!));
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Copied')),
                            );
                          },
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    SelectableText(
                      _result!,
                      style: GoogleFonts.lora(
                        fontSize: 14,
                        height: 1.45,
                        color: VillumTheme.parchment,
                      ),
                    ),
                  ],
                ),
              ),
            )
          else
            Padding(
              padding: const EdgeInsets.all(24),
              child: Text(
                'Pick a generator and roll.',
                style: GoogleFonts.lora(
                  fontSize: 12,
                  color: VillumTheme.parchmentMuted,
                ),
                textAlign: TextAlign.center,
              ),
            ),
        ],
      ),
    );
  }
}

// ------------------------------------------------------------------
// AI Draft Assistant tab
// ------------------------------------------------------------------
class _AiDraftTab extends ConsumerStatefulWidget {
  final int? campaignId;
  const _AiDraftTab({this.campaignId});
  @override
  ConsumerState<_AiDraftTab> createState() => _AiDraftTabState();
}

class _AiDraftTabState extends ConsumerState<_AiDraftTab> {
  DraftSession? _session;
  final _input = TextEditingController();
  bool _busy = false;
  String _status = 'ready'; // drafting | chatting | ready

  Future<void> _startDraft() async {
    if (_input.text.trim().isEmpty) return;
    setState(() => _busy = true);
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.post(
        '/api/ai/draft',
        data: {
          'prompt': _input.text.trim(),
          if (widget.campaignId != null) 'campaign_id': widget.campaignId,
        },
      );
      final data = resp.data;
      DraftSession s;
      if (data is Map<String, dynamic>) {
        s = DraftSession.fromJson(data);
      } else if (data is Map) {
        s = DraftSession.fromJson(Map<String, dynamic>.from(data));
      } else {
        s = DraftSession(
          id: 'new',
          entityType: '',
          status: 'chatting',
          message: data.toString(),
          messages: const [],
        );
      }
      if (!mounted) return;
      setState(() {
        _session = s;
        _status = s.status;
        _input.clear();
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Draft failed: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _turn(String msg) async {
    if (_session == null || msg.trim().isEmpty) return;
    setState(() => _busy = true);
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.post(
        '/api/ai/draft/${_session!.id}/turn',
        data: {'message': msg},
      );
      final data = resp.data;
      DraftSession s;
      if (data is Map<String, dynamic>) {
        s = DraftSession.fromJson(data);
      } else if (data is Map) {
        s = DraftSession.fromJson(Map<String, dynamic>.from(data));
      } else {
        s = _session!;
      }
      if (!mounted) return;
      setState(() {
        _session = s;
        _status = s.status;
        _input.clear();
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Turn failed: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _commit() async {
    if (_session == null) return;
    setState(() => _busy = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.post('/api/ai/draft/${_session!.id}/commit');
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('Draft committed')));
      setState(() {
        _session = null;
        _status = 'ready';
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Commit failed: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  void dispose() {
    _input.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(12, 12, 12, 8),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: _status == 'drafting'
                      ? VillumTheme.gold.withValues(alpha: 0.18)
                      : _status == 'chatting'
                      ? VillumTheme.hpTemp.withValues(alpha: 0.18)
                      : VillumTheme.surfaceElevated,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: VillumTheme.inkLight),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: _status == 'drafting'
                            ? VillumTheme.gold
                            : _status == 'chatting'
                            ? VillumTheme.hpTemp
                            : VillumTheme.hpHealthy,
                      ),
                    ),
                    const SizedBox(width: 6),
                    Text(
                      _status.toUpperCase(),
                      style: GoogleFonts.cinzel(
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                        color: VillumTheme.parchment,
                      ),
                    ),
                  ],
                ),
              ),
              const Spacer(),
              if (_session != null)
                OutlinedButton.icon(
                  onPressed: _busy ? null : _commit,
                  icon: const Icon(Icons.check, size: 16),
                  label: const Text('Commit'),
                ),
            ],
          ),
        ),
        Expanded(
          child: _session == null
              ? Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(
                        Icons.smart_toy_outlined,
                        size: 40,
                        color: VillumTheme.gold,
                      ),
                      const SizedBox(height: 12),
                      Text(
                        'AI Draft Assistant',
                        style: GoogleFonts.cinzel(
                          fontSize: 15,
                          fontWeight: FontWeight.w700,
                          color: VillumTheme.parchment,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        'Describe what you want to draft — an NPC, encounter, location, or lore note. The assistant will iterate with you, then commit it to your campaign.',
                        style: GoogleFonts.lora(
                          fontSize: 12,
                          color: VillumTheme.parchmentMuted,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 8,
                  ),
                  itemCount:
                      _session!.messages.length +
                      (_session!.draft != null ? 1 : 0),
                  itemBuilder: (ctx, i) {
                    if (_session!.draft != null &&
                        i == _session!.messages.length) {
                      return Card(
                        color: VillumTheme.gold.withValues(alpha: 0.08),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                          side: BorderSide(
                            color: VillumTheme.gold.withValues(alpha: 0.3),
                          ),
                        ),
                        child: Padding(
                          padding: const EdgeInsets.all(14),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'DRAFT PREVIEW',
                                style: GoogleFonts.cinzel(
                                  fontSize: 10,
                                  fontWeight: FontWeight.w800,
                                  color: VillumTheme.gold,
                                ),
                              ),
                              const SizedBox(height: 6),
                              SelectableText(
                                _session!.draft.toString(),
                                style: GoogleFonts.lora(
                                  fontSize: 13,
                                  color: VillumTheme.parchment,
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    }
                    final m = _session!.messages[i];
                    final isUser = m.role == 'user';
                    return Align(
                      alignment: isUser
                          ? Alignment.centerRight
                          : Alignment.centerLeft,
                      child: Container(
                        margin: const EdgeInsets.symmetric(vertical: 4),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 10,
                        ),
                        constraints: BoxConstraints(
                          maxWidth: MediaQuery.of(context).size.width * 0.78,
                        ),
                        decoration: BoxDecoration(
                          color: isUser
                              ? VillumTheme.gold.withValues(alpha: 0.18)
                              : VillumTheme.surfaceCard,
                          borderRadius: BorderRadius.circular(14),
                          border: Border.all(
                            color: isUser
                                ? VillumTheme.gold.withValues(alpha: 0.3)
                                : VillumTheme.inkLight,
                          ),
                        ),
                        child: SelectableText(
                          m.content,
                          style: GoogleFonts.lora(
                            fontSize: 13,
                            color: VillumTheme.parchment,
                          ),
                        ),
                      ),
                    );
                  },
                ),
        ),
        const Divider(height: 1),
        Padding(
          padding: const EdgeInsets.fromLTRB(12, 8, 12, 12),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _input,
                  minLines: 1,
                  maxLines: 3,
                  decoration: InputDecoration(
                    hintText: _session == null
                        ? 'Describe what to draft…'
                        : 'Reply to the assistant…',
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 12,
                    ),
                  ),
                  onSubmitted: (v) =>
                      _session == null ? _startDraft() : _turn(v),
                ),
              ),
              const SizedBox(width: 8),
              SizedBox(
                height: 48,
                width: 48,
                child: FilledButton(
                  onPressed: _busy
                      ? null
                      : () {
                          if (_session == null) {
                            _startDraft();
                          } else {
                            _turn(_input.text);
                          }
                        },
                  style: FilledButton.styleFrom(
                    padding: EdgeInsets.zero,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: _busy
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: VillumTheme.ink,
                          ),
                        )
                      : const Icon(Icons.send, size: 18),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
