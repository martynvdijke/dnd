import 'dart:io';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/api_client.dart';
import '../../core/providers.dart';
import '../../data/api/ink_api.dart';
import '../../data/models/ink_capture.dart';
import 'ink_canvas.dart';
import 'ink_queue.dart';

/// Review and capture handwriting. Deliberately monochrome with 56dp targets so
/// it stays readable on an e-ink tablet: no elevation, gradients, or animation.
class InkScreen extends ConsumerStatefulWidget {
  const InkScreen({super.key, this.initialCampaignId});

  final int? initialCampaignId;

  @override
  ConsumerState<InkScreen> createState() => _InkScreenState();
}

class _PendingPage {
  const _PendingPage({
    required this.path,
    required this.bytes,
    required this.label,
  });
  final String path;
  final Uint8List bytes;
  final String label;
}

class _InkScreenState extends ConsumerState<InkScreen> {
  static const _ink = Color(0xFF000000);
  static const _paper = Color(0xFFFFFFFF);

  int? _campaignId;
  int? _characterId;
  List<InkCapture>? _captures;
  String? _error;
  List<_PendingPage> _pending = const [];
  bool _busy = false;
  final _title = TextEditingController();

  InkQueue get _queue => InkQueue(ChannelInkQueueStore());

  @override
  void initState() {
    super.initState();
    _campaignId = widget.initialCampaignId;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _loadPending();
      if (_campaignId != null) _loadCaptures();
    });
  }

  @override
  void dispose() {
    _title.dispose();
    super.dispose();
  }

  Future<void> _loadPending() async {
    try {
      final shared = await const InkChannel().takeSharedPaths();
      final paths = await _queue.addAll(shared);
      final pending = <_PendingPage>[];
      for (final path in paths) {
        final f = File(path);
        if (!await f.exists()) continue;
        final bytes = await f.readAsBytes();
        final name = path.split(Platform.pathSeparator).last;
        pending.add(_PendingPage(path: path, bytes: bytes, label: name));
      }
      if (!mounted) return;
      setState(() => _pending = pending);
    } catch (_) {
      // Platform channel is absent on widget tests and desktop; the queue just
      // stays empty.
    }
  }

  Future<void> _loadCaptures() async {
    final cid = _campaignId;
    if (cid == null) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final api = InkApi(ref.read(apiClientProvider));
      final captures = await api.list(campaignId: cid);
      if (!mounted) return;
      setState(() {
        _captures = captures;
        _busy = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _busy = false;
      });
    }
  }

  Future<void> _drawPage() async {
    final png = await Navigator.of(
      context,
    ).push<Uint8List>(MaterialPageRoute(builder: (_) => const InkCanvasPage()));
    if (png == null) return;
    final path =
        '${Directory.systemTemp.path}/ink_${DateTime.now().millisecondsSinceEpoch}.png';
    await File(path).writeAsBytes(png);
    await _queue.addAll([path]);
    await _loadPending();
  }

  Future<void> _uploadPending() async {
    final cid = _campaignId;
    if (cid == null) {
      _toast('Choose a campaign first.');
      return;
    }
    if (_pending.isEmpty) return;
    setState(() => _busy = true);
    try {
      final api = InkApi(ref.read(apiClientProvider));
      await api.create(
        campaignId: cid,
        characterId: _characterId,
        title: _title.text.trim(),
        source: 'share',
        pages: _pending.map((p) => p.bytes).toList(),
      );
      await _queue.removeAll(_pending.map((p) => p.path));
      _title.clear();
      await _loadPending();
      await _loadCaptures();
      _toast('Pages synced.');
    } catch (e) {
      // Keep the pages queued: the next attempt (or the next app launch) retries.
      _toast('Still queued: $e');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _toast(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        backgroundColor: _ink,
        content: Text(message, style: const TextStyle(color: _paper)),
      ),
    );
  }

  Future<void> _editText(InkCapture capture, InkPage page) async {
    final controller = TextEditingController(text: page.text);
    final text = await showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: _paper,
        title: Text('Page ${page.pageIndex + 1}'),
        content: TextField(
          controller: controller,
          maxLines: 10,
          autofocus: true,
          style: const TextStyle(color: _ink),
          decoration: const InputDecoration(
            hintText: 'Type what the page says',
            border: OutlineInputBorder(),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: _ink,
              foregroundColor: _paper,
            ),
            onPressed: () => Navigator.pop(ctx, controller.text),
            child: const Text('Save'),
          ),
        ],
      ),
    );
    controller.dispose();
    if (text == null) return;
    try {
      final api = InkApi(ref.read(apiClientProvider));
      await api.updateText(capture.id, {page.id: text});
      await _loadCaptures();
    } catch (e) {
      _toast('$e');
    }
  }

  Future<void> _ocr(InkCapture capture) async {
    try {
      final api = InkApi(ref.read(apiClientProvider));
      await api.ocr(capture.id);
      await _loadCaptures();
      _toast('Transcription ready.');
    } on ApiException catch (e) {
      _toast('${e.message} Type the text manually.');
    } catch (e) {
      _toast('$e');
    }
  }

  Future<void> _promote(InkCapture capture, String target) async {
    try {
      final api = InkApi(ref.read(apiClientProvider));
      await api.promote(capture.id, target);
      await _loadCaptures();
      _toast(
        target == 'recap'
            ? 'Saved as an unpublished recap.'
            : 'Saved to the journal.',
      );
    } catch (e) {
      _toast('$e');
    }
  }

  Future<void> _delete(InkCapture capture) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: _paper,
        title: const Text('Delete capture?'),
        content: const Text('The page images stay in the media library.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: _ink,
              foregroundColor: _paper,
            ),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (ok != true) return;
    try {
      await InkApi(ref.read(apiClientProvider)).delete(capture.id);
      await _loadCaptures();
    } catch (e) {
      _toast('$e');
    }
  }

  @override
  Widget build(BuildContext context) {
    final campaigns = ref.watch(campaignsListProvider);
    final characters = ref.watch(charactersListProvider);
    return Scaffold(
      backgroundColor: _paper,
      appBar: AppBar(
        backgroundColor: _paper,
        foregroundColor: _ink,
        elevation: 0,
        title: const Text('Ink'),
        actions: [
          IconButton(
            style: IconButton.styleFrom(minimumSize: const Size(56, 56)),
            icon: const Icon(Icons.refresh),
            tooltip: 'Reload',
            onPressed: () {
              _loadPending();
              _loadCaptures();
            },
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        backgroundColor: _ink,
        foregroundColor: _paper,
        onPressed: _drawPage,
        icon: const Icon(Icons.gesture),
        label: const Text('Draw page'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 96),
        children: [
          campaigns.when(
            loading: () => const LinearProgressIndicator(color: _ink),
            error: (e, _) => Text('Could not load campaigns: $e'),
            data: (list) {
              if (list.isEmpty) {
                return const Text('Create a campaign on the web app first.');
              }
              _campaignId ??= list.first.id;
              return DropdownButtonFormField<int>(
                initialValue: _campaignId,
                decoration: const InputDecoration(
                  labelText: 'Campaign',
                  border: OutlineInputBorder(),
                ),
                style: const TextStyle(color: _ink),
                items: list
                    .map(
                      (c) => DropdownMenuItem(value: c.id, child: Text(c.name)),
                    )
                    .toList(),
                onChanged: (id) {
                  setState(() => _campaignId = id);
                  _loadCaptures();
                },
              );
            },
          ),
          const SizedBox(height: 12),
          characters.maybeWhen(
            data: (list) {
              if (list.isEmpty) return const SizedBox.shrink();
              return DropdownButtonFormField<int?>(
                initialValue: _characterId,
                decoration: const InputDecoration(
                  labelText: 'Personal notes for (optional)',
                  helperText: 'Leave empty for campaign ink (DM). Pick a character to file it in their journal.',
                  border: OutlineInputBorder(),
                ),
                style: const TextStyle(color: _ink),
                items: [
                  const DropdownMenuItem<int?>(
                    value: null,
                    child: Text('Campaign ink'),
                  ),
                  ...list
                      .where((c) => c.canEdit)
                      .map(
                        (c) => DropdownMenuItem<int?>(
                          value: c.id,
                          child: Text(c.name),
                        ),
                      ),
                ],
                onChanged: (id) {
                  setState(() => _characterId = id);
                  _loadCaptures();
                },
              );
            },
            orElse: () => const SizedBox.shrink(),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _title,
            style: const TextStyle(color: _ink),
            decoration: const InputDecoration(
              labelText: 'Title (optional)',
              hintText: 'Session 12 notes',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 16),
          _pendingSection(),
          const SizedBox(height: 16),
          Row(
            children: [
              const Expanded(
                child: Text(
                  'Captures',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
              TextButton(
                style: TextButton.styleFrom(
                  foregroundColor: _ink,
                  minimumSize: const Size(56, 56),
                ),
                onPressed: _busy ? null : _loadCaptures,
                child: const Text('Reload'),
              ),
            ],
          ),
          if (_error != null)
            Text(_error!, style: const TextStyle(color: _ink)),
          if (_captures == null && _busy)
            const LinearProgressIndicator(color: _ink),
          if (_captures != null && _captures!.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 12),
              child: Text(
                'No captures yet. Share pages from the Boox notes app or draw one here.',
              ),
            ),
          ...?_captures?.map(_captureCard),
        ],
      ),
    );
  }

  Widget _pendingSection() {
    if (_pending.isEmpty) {
      return const Text(
        'No waiting pages. Share images from the Boox notes app to Villum, or tap Draw page.',
      );
    }
    return Container(
      decoration: BoxDecoration(border: Border.all(color: _ink)),
      padding: const EdgeInsets.all(8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Waiting to sync (${_pending.length})',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _pending
                .map(
                  (p) => Column(
                    children: [
                      Image.memory(
                        p.bytes,
                        width: 88,
                        height: 88,
                        fit: BoxFit.contain,
                      ),
                      SizedBox(
                        width: 88,
                        child: Text(
                          p.label,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                )
                .toList(),
          ),
          const SizedBox(height: 8),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: _ink,
              foregroundColor: _paper,
              minimumSize: const Size(120, 56),
            ),
            onPressed: _busy ? null : _uploadPending,
            child: Text(_busy ? 'Syncing…' : 'Sync pages'),
          ),
        ],
      ),
    );
  }

  Widget _captureCard(InkCapture capture) {
    final base = ref.watch(serverUrlProvider);
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(border: Border.all(color: _ink)),
      padding: const EdgeInsets.all(8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  capture.title.isEmpty ? 'Untitled capture' : capture.title,
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
              Text(capture.status.toUpperCase()),
              PopupMenuButton<String>(
                iconColor: _ink,
                tooltip: 'Actions',
                onSelected: (value) {
                  switch (value) {
                    case 'ocr':
                      _ocr(capture);
                    case 'recap':
                      _promote(capture, 'recap');
                    case 'journal':
                      _promote(capture, 'journal');
                    case 'delete':
                      _delete(capture);
                  }
                },
                itemBuilder: (_) => [
                  const PopupMenuItem(value: 'ocr', child: Text('Transcribe')),
                  if (!capture.isPromoted) ...[
                    const PopupMenuItem(
                      value: 'recap',
                      child: Text('Save as recap'),
                    ),
                    if (capture.characterId != null)
                      const PopupMenuItem(
                        value: 'journal',
                        child: Text('Save to journal'),
                      ),
                  ],
                  const PopupMenuItem(value: 'delete', child: Text('Delete')),
                ],
              ),
            ],
          ),
          if (capture.error.isNotEmpty) Text(capture.error),
          ...capture.pages.map(
            (p) => Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (p.thumbnailUrl.isNotEmpty)
                    Image.network(
                      '$base${p.thumbnailUrl}',
                      width: 88,
                      height: 120,
                      fit: BoxFit.contain,
                    )
                  else
                    const Icon(Icons.description_outlined, size: 56),
                  const SizedBox(width: 8),
                  Expanded(
                    child: p.text.isEmpty
                        ? const Text('No text yet — transcribe or type it.')
                        : Text(
                            p.text,
                            maxLines: 4,
                            overflow: TextOverflow.ellipsis,
                          ),
                  ),
                  IconButton(
                    style: IconButton.styleFrom(
                      minimumSize: const Size(56, 56),
                    ),
                    icon: const Icon(Icons.edit),
                    tooltip: 'Edit text',
                    onPressed: () => _editText(capture, p),
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
