import 'dart:convert';
import 'dart:io';

import 'package:flutter/services.dart';

/// Native bridge for the ink feature.
///
/// `takeSharedImages` copies Android share-sheet content URIs into app storage
/// and returns their paths, so Dart never has to deal with `content://` URIs.
/// `readPending`/`writePending` persist the offline upload queue as JSON.
class InkChannel {
  static const MethodChannel channel = MethodChannel('xyz.vandijke.dnd/ink');

  const InkChannel();

  Future<List<String>> takeSharedPaths() async {
    final paths = await channel.invokeListMethod<String>('takeSharedImages');
    return paths ?? const [];
  }

  Future<List<Uint8List>> takeSharedImages() async {
    final paths = await takeSharedPaths();
    final out = <Uint8List>[];
    for (final path in paths) {
      final f = File(path);
      if (await f.exists()) out.add(await f.readAsBytes());
    }
    return out;
  }

  Future<String?> readPending() => channel.invokeMethod<String>('readPending');

  Future<void> writePending(String json) =>
      channel.invokeMethod<void>('writePending', {'json': json});
}

/// Durable store for the offline queue. Abstracted so widget tests can use an
/// in-memory implementation instead of the platform channel.
abstract class InkQueueStore {
  Future<String?> read();
  Future<void> write(String json);
}

class ChannelInkQueueStore implements InkQueueStore {
  final InkChannel channel;
  ChannelInkQueueStore([this.channel = const InkChannel()]);
  @override
  Future<String?> read() => channel.readPending();
  @override
  Future<void> write(String json) => channel.writePending(json);
}

class MemoryInkQueueStore implements InkQueueStore {
  String? value;
  MemoryInkQueueStore([this.value]);
  @override
  Future<String?> read() async => value;
  @override
  Future<void> write(String json) async => value = json;
}

/// A queue of page image paths waiting to be uploaded. Pages survive process
/// death because the paths live in app storage.
class InkQueue {
  final InkQueueStore store;
  InkQueue(this.store);

  Future<List<String>> load() async {
    final raw = await store.read();
    if (raw == null || raw.trim().isEmpty) return [];
    try {
      final decoded = jsonDecode(raw);
      if (decoded is! List) return [];
      return decoded
          .whereType<String>()
          .where((p) => File(p).existsSync())
          .toList();
    } catch (_) {
      return [];
    }
  }

  Future<void> save(List<String> paths) => store.write(jsonEncode(paths));

  Future<List<String>> addAll(Iterable<String> paths) async {
    final current = await load();
    for (final p in paths) {
      if (!current.contains(p)) current.add(p);
    }
    await save(current);
    return current;
  }

  Future<List<String>> removeAll(Iterable<String> paths) async {
    final remove = paths.toSet();
    final current = (await load()).where((p) => !remove.contains(p)).toList();
    await save(current);
    return current;
  }
}
