import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';
import '../../core/auth_repository.dart';
import '../../core/providers.dart';
import '../shared/states.dart';

class BattlemapScreen extends ConsumerStatefulWidget {
  final int? campaignId;
  const BattlemapScreen({super.key, this.campaignId});

  @override
  ConsumerState<BattlemapScreen> createState() => _BattlemapScreenState();
}

class _BattlemapScreenState extends ConsumerState<BattlemapScreen> {
  int? _campaignId;
  final TransformationController _transform = TransformationController();
  Size _mapSize = const Size(
    800,
    600,
  ); // logical map size for normalised conversion

  @override
  void initState() {
    super.initState();
    _campaignId = widget.campaignId;
  }

  @override
  void dispose() {
    _transform.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Listen for battlemap_update WS
    ref.listen(wsEventsProvider, (prev, next) {
      next.whenData((ev) {
        if (ev.type == 'battlemap_update' && _campaignId != null) {
          final cid = ev.payload['campaign_id'];
          if (cid == _campaignId) {
            ref.invalidate(battlemapProvider(_campaignId!));
          }
        }
      });
    });

    final auth = ref.watch(authControllerProvider);
    final isDm = _isDm(auth);

    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Battlemap',
          style: GoogleFonts.cinzel(fontWeight: FontWeight.w700),
        ),
        actions: [
          if (_campaignId != null)
            IconButton(
              icon: const Icon(Icons.my_location),
              tooltip: 'Reset view',
              onPressed: () => _transform.value = Matrix4.identity(),
            ),
          if (_campaignId != null && isDm)
            IconButton(
              icon: const Icon(Icons.sync),
              tooltip: 'Sync map',
              onPressed: () => _syncMap(),
            ),
        ],
      ),
      body: Column(
        children: [
          _campaignPicker(isDm),
          Expanded(
            child: _campaignId == null ? _emptyCampaign() : _mapBody(isDm),
          ),
        ],
      ),
      floatingActionButton: _campaignId == null || !isDm
          ? null
          : FloatingActionButton.extended(
              backgroundColor: VillumTheme.gold,
              foregroundColor: VillumTheme.ink,
              onPressed: () => _addTokenDialog(),
              icon: const Icon(Icons.add_location_alt),
              label: const Text('Add token'),
            ),
    );
  }

  bool _isDm(AuthState authState) =>
      authState is AuthAuthenticated && authState.user.isDm;

  Widget _campaignPicker(bool isDm) {
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
            padding: const EdgeInsets.all(12),
            child: Text(
              'No campaigns yet.',
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
          child: Row(
            children: [
              Expanded(
                child: DropdownButtonFormField<int>(
                  key: ValueKey(_campaignId),
                  initialValue:
                      _campaignId != null &&
                          list.any((e) => e.id == _campaignId)
                      ? _campaignId
                      : null,
                  hint: const Text('Select campaign'),
                  decoration: const InputDecoration(
                    labelText: 'Campaign',
                    prefixIcon: Icon(Icons.flag_outlined),
                  ),
                  items: list
                      .map(
                        (e) => DropdownMenuItem<int>(
                          value: e.id,
                          child: Text(e.name),
                        ),
                      )
                      .toList(),
                  onChanged: (v) => setState(() => _campaignId = v),
                ),
              ),
              if (!isDm) ...[
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: VillumTheme.surfaceElevated,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: VillumTheme.inkLight),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(
                        Icons.visibility,
                        size: 14,
                        color: VillumTheme.parchmentMuted,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        'Player view',
                        style: GoogleFonts.lora(
                          fontSize: 11,
                          color: VillumTheme.parchmentMuted,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ],
          ),
        );
      },
    );
  }

  Widget _emptyCampaign() => const EmptyState(
    icon: Icons.map_outlined,
    title: 'Choose a campaign',
    subtitle: 'Pick a campaign to load its battlemap. Tokens are shared live with the table.',
  );

  Widget _mapBody(bool isDm) {
    final async = ref.watch(battlemapProvider(_campaignId!));
    return async.when(
      loading: () => const LoadingState(message: 'Loading battlemap…'),
      error: (e, _) => ErrorState(
        message: e.toString(),
        onRetry: () => ref.invalidate(battlemapProvider(_campaignId!)),
      ),
      data: (state) {
        final tokens = state.tokens;
        final meta = state.map;
        if (meta != null) {
          _mapSize = Size(meta.width.toDouble(), meta.height.toDouble());
        }
        if (meta == null && tokens.isEmpty) {
          return EmptyState(
            icon: Icons.map,
            title: 'No battlemap',
            subtitle: isDm
                ? 'Create a battlemap in the Villum web app, then add tokens here.'
                : 'No battlemap for this campaign yet.',
            action: isDm
                ? FilledButton.icon(
                    onPressed: _addTokenDialog,
                    icon: const Icon(Icons.add),
                    label: const Text('Add first token'),
                  )
                : null,
          );
        }
        return LayoutBuilder(
          builder: (ctx, constraints) {
            return Stack(
              children: [
                // Map canvas
                InteractiveViewer(
                  transformationController: _transform,
                  boundaryMargin: const EdgeInsets.all(120),
                  minScale: 0.3,
                  maxScale: 4.0,
                  constrained: false,
                  child: Container(
                    width: _mapSize.width,
                    height: _mapSize.height,
                    decoration: BoxDecoration(
                      color: const Color(0xFF2A2520),
                      border: Border.all(
                        color: VillumTheme.gold.withValues(alpha: 0.35),
                        width: 2,
                      ),
                      borderRadius: BorderRadius.circular(8),
                      image: meta != null && meta.imageUrl.isNotEmpty
                          ? DecorationImage(
                              image: NetworkImage(meta.imageUrl),
                              fit: BoxFit.cover,
                            )
                          : null,
                    ),
                    child: meta == null || meta.imageUrl.isEmpty
                        ? CustomPaint(
                            size: _mapSize,
                            painter: _GridPainter(
                              gridSize: (meta?.gridSize ?? 50).toDouble(),
                            ),
                          )
                        : null,
                  ),
                ),
                // Tokens overlay — positioned against map size then transformed by InteractiveViewer? Simplified: stack tokens as Positioned within InteractiveViewer child
                // For correctness we embed tokens inside the InteractiveViewer child via a second InteractiveViewer layer trick: we use a Stack inside a SizedBox.
                // Simpler for now: render tokens as an overlay that tracks transform — we render them inside the InteractiveViewer by using its child stacking.
                // To keep code simple, re-render tokens inside the viewer via a custom child.
                Positioned.fill(
                  child: InteractiveViewer(
                    transformationController: _transform,
                    boundaryMargin: const EdgeInsets.all(120),
                    minScale: 0.3,
                    maxScale: 4.0,
                    constrained: false,
                    panEnabled: false,
                    scaleEnabled: false,
                    child: SizedBox(
                      width: _mapSize.width,
                      height: _mapSize.height,
                      child: Stack(
                        children: tokens
                            .map((t) => _tokenWidget(t, isDm))
                            .toList(),
                      ),
                    ),
                  ),
                ),
                // Legend
                Positioned(
                  left: 12,
                  bottom: 12,
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      color: VillumTheme.surfaceCard.withValues(alpha: 0.92),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: VillumTheme.inkLight),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 10,
                          height: 10,
                          decoration: const BoxDecoration(
                            color: VillumTheme.gold,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          '${tokens.length} tokens',
                          style: GoogleFonts.lora(
                            fontSize: 12,
                            color: VillumTheme.parchment,
                          ),
                        ),
                        const SizedBox(width: 10),
                        const Icon(
                          Icons.open_with,
                          size: 14,
                          color: VillumTheme.parchmentMuted,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          'Drag to move',
                          style: GoogleFonts.lora(
                            fontSize: 11,
                            color: VillumTheme.parchmentMuted,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                if (!isDm)
                  Positioned(
                    right: 12,
                    top: 12,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: VillumTheme.ink.withValues(alpha: 0.85),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: VillumTheme.gold.withValues(alpha: 0.3),
                        ),
                      ),
                      child: Text(
                        'Tap & drag your token — any member may move',
                        style: GoogleFonts.lora(
                          fontSize: 11,
                          color: VillumTheme.parchmentMuted,
                        ),
                      ),
                    ),
                  ),
              ],
            );
          },
        );
      },
    );
  }

  Widget _tokenWidget(BattlemapToken token, bool isDm) {
    final left = token.x * _mapSize.width - 22;
    final top = token.y * _mapSize.height - 22;
    final max = token.hpMax ?? 10;
    final cur = token.hpCurrent ?? max;
    final hpRatio = max > 0 ? (cur / max).clamp(0.0, 1.0) : 1.0;
    final hpColor = VillumTheme.hpColor(hpRatio);
    final tokenColor = _parseColor(token.color);

    return Positioned(
      left: left,
      top: top,
      child: GestureDetector(
        onPanUpdate: (d) => _onDrag(token, d.delta),
        onPanEnd: (_) => _persist(token),
        child: Semantics(
          label:
              'Token ${token.name} at ${(token.x * 100).round()}% ${(token.y * 100).round()}%',
          button: true,
          child: SizedBox(
            width: 44,
            height: 52,
            child: Stack(
              alignment: Alignment.topCenter,
              children: [
                // Aura / vision radii (visual only, scaled to grid)
                if (token.auraRadius > 0)
                  Positioned(
                    left: 22 - token.auraRadius * 8,
                    top: 22 - token.auraRadius * 8,
                    child: Container(
                      width: token.auraRadius * 16,
                      height: token.auraRadius * 16,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: tokenColor.withValues(alpha: 0.10),
                        border: Border.all(
                          color: tokenColor.withValues(alpha: 0.28),
                          width: 1,
                        ),
                      ),
                    ),
                  ),
                // Token circle with HP ring
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: tokenColor,
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.9),
                      width: 2,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black45,
                        blurRadius: 6,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Stack(
                    children: [
                      Center(
                        child: Text(
                          token.name.isNotEmpty
                              ? token.name[0].toUpperCase()
                              : '?',
                          style: GoogleFonts.cinzel(
                            fontSize: 16,
                            fontWeight: FontWeight.w900,
                            color: Colors.white,
                          ),
                        ),
                      ),
                      // HP ring
                      Positioned.fill(
                        child: CircularProgressIndicator(
                          value: hpRatio,
                          strokeWidth: 3,
                          backgroundColor: Colors.black26,
                          valueColor: AlwaysStoppedAnimation(hpColor),
                        ),
                      ),
                    ],
                  ),
                ),
                // Name label
                Positioned(
                  bottom: 0,
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: VillumTheme.ink.withValues(alpha: 0.9),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(color: VillumTheme.inkLight),
                    ),
                    child: Text(
                      token.name,
                      style: GoogleFonts.lora(
                        fontSize: 9,
                        fontWeight: FontWeight.w700,
                        color: VillumTheme.parchment,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ),
                // Condition dot
                if (token.conditions != null && token.conditions!.isNotEmpty)
                  Positioned(
                    right: 0,
                    top: 0,
                    child: Container(
                      width: 14,
                      height: 14,
                      decoration: BoxDecoration(
                        color: VillumTheme.hpDying,
                        shape: BoxShape.circle,
                        border: Border.all(color: Colors.white, width: 1.2),
                      ),
                      child: Center(
                        child: Text(
                          '${token.conditions!.length}',
                          style: GoogleFonts.lora(
                            fontSize: 8,
                            fontWeight: FontWeight.w800,
                            color: Colors.white,
                          ),
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  final Map<int, BattlemapToken> _dirty = {};

  void _onDrag(BattlemapToken token, Offset delta) {
    final scale = _transform.value.getMaxScaleOnAxis();
    final nx = (token.x + delta.dx / _mapSize.width / scale).clamp(0.0, 1.0);
    final ny = (token.y + delta.dy / _mapSize.height / scale).clamp(0.0, 1.0);
    final updated = BattlemapToken(
      id: token.id,
      campaignId: token.campaignId,
      combatEntryId: token.combatEntryId,
      name: token.name,
      x: nx,
      y: ny,
      color: token.color,
      size: token.size,
      hpCurrent: token.hpCurrent,
      hpMax: token.hpMax,
      ac: token.ac,
      conditions: token.conditions,
      auraRadius: token.auraRadius,
      visionRadius: token.visionRadius,
    );
    setState(() {
      _dirty[token.id] = updated;
    });
    try {
      final notifier = ref.read(battlemapProvider(_campaignId!).notifier);
      notifier.moveToken(token.id, nx, ny);
    } catch (_) {}
  }

  Future<void> _persist(BattlemapToken token) async {
    final pending = _dirty[token.id];
    if (pending == null) return;
    _dirty.remove(token.id);
    try {
      final api = ref.read(apiClientProvider);
      await api.put(
        '/api/battlemap-tokens/${token.id}',
        data: {'x': pending.x, 'y': pending.y},
      );
      HapticFeedback.lightImpact();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Move failed: $e')));
      }
      // revert via refetch
      if (_campaignId != null) ref.invalidate(battlemapProvider(_campaignId!));
    }
  }

  Future<void> _syncMap() async {
    if (_campaignId == null) return;
    try {
      final api = ref.read(apiClientProvider);
      await api.post('/api/campaigns/$_campaignId/battlemap/sync');
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('Battlemap synced')));
      }
      ref.invalidate(battlemapProvider(_campaignId!));
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Sync failed: $e')));
      }
    }
  }

  void _addTokenDialog() {
    final nameCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (dCtx) => AlertDialog(
        title: const Text('Add token'),
        content: TextField(
          controller: nameCtrl,
          decoration: const InputDecoration(
            labelText: 'Token name',
            prefixIcon: Icon(Icons.person_pin_circle_outlined),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dCtx),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () async {
              final name = nameCtrl.text.trim();
              if (name.isEmpty) return;
              try {
                final api = ref.read(apiClientProvider);
                await api.post(
                  '/api/campaigns/$_campaignId/battlemap/tokens',
                  data: {
                    'name': name,
                    'x': 0.5,
                    'y': 0.5,
                    'color': '#b8963e',
                    'size': 1,
                  },
                );
                if (dCtx.mounted) Navigator.pop(dCtx);
                if (_campaignId != null) {
                  ref.invalidate(battlemapProvider(_campaignId!));
                }
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Token "$name" added')),
                  );
                }
              } catch (e) {
                if (mounted) {
                  ScaffoldMessenger.of(context)
                      .showSnackBar(SnackBar(content: Text('Add failed: $e')));
                }
              }
            },
            child: const Text('Add'),
          ),
        ],
      ),
    );
  }

  Color _parseColor(String hex) {
    try {
      var h = hex.replaceFirst('#', '');
      if (h.length == 6) h = 'FF$h';
      return Color(int.parse(h, radix: 16));
    } catch (_) {
      return VillumTheme.gold;
    }
  }
}

class _GridPainter extends CustomPainter {
  final double gridSize;
  const _GridPainter({required this.gridSize});
  @override
  void paint(Canvas canvas, Size size) {
    final p = Paint()
      ..color = const Color(0xFF3D352D).withValues(alpha: 0.5)
      ..strokeWidth = 0.5;
    for (double x = 0; x <= size.width; x += gridSize) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), p);
    }
    for (double y = 0; y <= size.height; y += gridSize) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), p);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
