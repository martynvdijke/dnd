import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/providers.dart';
import '../shared/states.dart';

class CompendiumScreen extends ConsumerStatefulWidget {
  const CompendiumScreen({super.key});

  @override
  ConsumerState<CompendiumScreen> createState() => _CompendiumScreenState();
}

class _CompendiumScreenState extends ConsumerState<CompendiumScreen> {
  final _searchCtrl = TextEditingController();
  String? _selectedType;
  Timer? _debounce;
  String _query = '';
  List<CompendiumSearchResult> _results = [];
  bool _loading = false;
  String? _error;

  static const _types = [
    'spell',
    'equipment',
    'monster',
    'race',
    'feat',
    'background',
    'class',
  ];

  @override
  void dispose() {
    _debounce?.cancel();
    _searchCtrl.dispose();
    super.dispose();
  }

  void _onChanged(String v) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 250), () {
      setState(() => _query = v.trim());
      _search();
    });
  }

  Future<void> _search() async {
    if (_query.isEmpty) {
      setState(() {
        _results = [];
        _loading = false;
        _error = null;
      });
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.get(
        '/api/compendium/search',
        queryParameters: {'q': _query, 'type': ?_selectedType},
      );
      final data = resp.data;
      List<CompendiumSearchResult> out = [];
      if (data is List) {
        out = data
            .map(
              (e) => CompendiumSearchResult.fromJson(
                Map<String, dynamic>.from(e as Map),
              ),
            )
            .toList();
      } else if (data is Map && data['results'] is List) {
        out = (data['results'] as List)
            .map(
              (e) => CompendiumSearchResult.fromJson(
                Map<String, dynamic>.from(e as Map),
              ),
            )
            .toList();
      }
      if (!mounted) return;
      setState(() {
        _results = out;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Compendium',
          style: GoogleFonts.cinzel(fontWeight: FontWeight.w700),
        ),
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 12, 12, 8),
            child: TextField(
              controller: _searchCtrl,
              onChanged: _onChanged,
              textInputAction: TextInputAction.search,
              decoration: InputDecoration(
                hintText: 'Search spells, monsters, items…',
                prefixIcon: const Icon(Icons.search),
                suffixIcon: _query.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchCtrl.clear();
                          _onChanged('');
                        },
                        tooltip: 'Clear',
                      )
                    : null,
              ),
            ),
          ),
          SizedBox(
            height: 44,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              itemCount: _types.length + 1,
              separatorBuilder: (_, _) => const SizedBox(width: 8),
              itemBuilder: (ctx, i) {
                if (i == 0) {
                  final sel = _selectedType == null;
                  return FilterChip(
                    label: const Text('All'),
                    selected: sel,
                    onSelected: (_) {
                      setState(() => _selectedType = null);
                      _search();
                    },
                    selectedColor: VillumTheme.gold.withValues(alpha: 0.22),
                    checkmarkColor: VillumTheme.gold,
                  );
                }
                final t = _types[i - 1];
                final sel = _selectedType == t;
                return FilterChip(
                  label: Text(t[0].toUpperCase() + t.substring(1)),
                  selected: sel,
                  onSelected: (_) {
                    setState(() => _selectedType = sel ? null : t);
                    _search();
                  },
                  selectedColor: VillumTheme.gold.withValues(alpha: 0.22),
                  checkmarkColor: VillumTheme.gold,
                );
              },
            ),
          ),
          const Divider(height: 1),
          Expanded(child: _body()),
        ],
      ),
    );
  }

  Widget _body() {
    if (_error != null) {
      return ErrorState(message: _error!, onRetry: _search);
    }
    if (_loading) {
      return const LoadingState(message: 'Searching…');
    }
    if (_query.isEmpty) {
      return EmptyState(
        icon: Icons.menu_book_outlined,
        title: 'Search the compendium',
        subtitle: 'Type a name — e.g. "fireball", "goblin", or "longsword". Results appear instantly.',
        action: Wrap(
          spacing: 8,
          children: _types
              .take(4)
              .map(
                (t) => ActionChip(
                  label: Text(t),
                  onPressed: () {
                    setState(() => _selectedType = t);
                    _searchCtrl.text = t == 'spell'
                        ? 'fire'
                        : t == 'monster'
                        ? 'gob'
                        : 'sword';
                    _onChanged(_searchCtrl.text);
                  },
                ),
              )
              .toList(),
        ),
      );
    }
    if (_results.isEmpty) {
      return EmptyState(
        icon: Icons.search_off,
        title: 'No results',
        subtitle:
            'No matches for "$_query"${_selectedType != null ? ' in ${_selectedType}s' : ''}. Try a broader term.',
      );
    }
    return ListView.separated(
      padding: const EdgeInsets.symmetric(vertical: 8),
      itemCount: _results.length,
      separatorBuilder: (_, _) =>
          const Divider(height: 1, indent: 16, endIndent: 16),
      itemBuilder: (ctx, i) {
        final r = _results[i];
        return ListTile(
          leading: _typeIcon(r.type),
          title: Text(
            r.name,
            style: GoogleFonts.lora(fontWeight: FontWeight.w600),
          ),
          subtitle: Text(
            _subtitle(r),
            style: GoogleFonts.lora(
              fontSize: 12,
              color: VillumTheme.parchmentMuted,
            ),
          ),
          trailing: Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
            decoration: BoxDecoration(
              color: VillumTheme.surfaceElevated,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: VillumTheme.inkLight),
            ),
            child: Text(
              r.type.toUpperCase(),
              style: GoogleFonts.cinzel(
                fontSize: 9,
                fontWeight: FontWeight.w800,
                color: VillumTheme.gold,
              ),
            ),
          ),
          onTap: () => _openDetail(r),
        );
      },
    );
  }

  Widget _typeIcon(String type) {
    IconData icon;
    Color color;
    switch (type) {
      case 'spell':
        icon = Icons.auto_fix_high;
        color = const Color(0xFF7E57C2);
        break;
      case 'monster':
        icon = Icons.pets;
        color = VillumTheme.hpDying;
        break;
      case 'equipment':
        icon = Icons.shield_outlined;
        color = VillumTheme.gold;
        break;
      case 'race':
        icon = Icons.face;
        color = const Color(0xFF4FC3F7);
        break;
      case 'feat':
        icon = Icons.emoji_events_outlined;
        color = const Color(0xFFFFB74D);
        break;
      default:
        icon = Icons.article_outlined;
        color = VillumTheme.parchmentMuted;
    }
    return Container(
      width: 40,
      height: 40,
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.14),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: color.withValues(alpha: 0.28)),
      ),
      child: Icon(icon, size: 20, color: color),
    );
  }

  String _subtitle(CompendiumSearchResult r) {
    if (r.type == 'spell' && r.level != null) {
      final lvl = r.level == 0 ? 'Cantrip' : 'Level ${r.level}';
      return '$lvl · ${r.subtype ?? '—'}';
    }
    if (r.type == 'monster' && r.cr != null) {
      return 'CR ${r.cr} · ${r.subtype ?? ''}';
    }
    if (r.type == 'equipment') return r.subtype ?? 'Equipment';
    if (r.subtype != null && r.subtype!.isNotEmpty) return r.subtype!;
    return r.type;
  }

  void _openDetail(CompendiumSearchResult r) {
    showModalBottomSheet(
      context: context,
      backgroundColor: VillumTheme.surfaceCard,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.fromLTRB(20, 20, 20, 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                _typeIcon(r.type),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        r.name,
                        style: GoogleFonts.cinzel(
                          fontSize: 16,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      Text(
                        _subtitle(r),
                        style: GoogleFonts.lora(
                          fontSize: 12,
                          color: VillumTheme.parchmentMuted,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              'Full details live in the Villum web app. This preview shows the indexed compendium hit — tap Open to view it on the server.',
              style: GoogleFonts.lora(
                fontSize: 12,
                color: VillumTheme.parchmentMuted,
              ),
            ),
            const SizedBox(height: 14),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: () => Navigator.pop(ctx),
                icon: const Icon(Icons.check, size: 18),
                label: const Text('Close'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
