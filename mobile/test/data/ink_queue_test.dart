import 'dart:io';

import 'package:dnd/features/ink/ink_queue.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late Directory temp;

  setUp(() async {
    temp = await Directory.systemTemp.createTemp('ink_queue_test');
  });

  tearDown(() async {
    if (await temp.exists()) await temp.delete(recursive: true);
  });

  Future<String> page(String name) async {
    final f = File('${temp.path}/$name');
    await f.writeAsBytes([1, 2, 3]);
    return f.path;
  }

  test('starts empty and tolerates corrupt storage', () async {
    final queue = InkQueue(MemoryInkQueueStore('not json'));
    expect(await queue.load(), isEmpty);
  });

  test('adds, deduplicates and removes pages', () async {
    final a = await page('a.png');
    final b = await page('b.png');
    final queue = InkQueue(MemoryInkQueueStore());

    expect(await queue.addAll([a, b, a]), [a, b]);
    expect(await queue.load(), [a, b]);

    expect(await queue.removeAll([a]), [b]);
    expect(await queue.load(), [b]);
  });

  test('drops paths that no longer exist', () async {
    final a = await page('a.png');
    final b = await page('b.png');
    final queue = InkQueue(MemoryInkQueueStore());
    await queue.addAll([a, b]);
    await File(b).delete();

    expect(await queue.load(), [a]);
  });

  test('persists across queue instances', () async {
    final a = await page('a.png');
    final store = MemoryInkQueueStore();
    await InkQueue(store).addAll([a]);

    expect(await InkQueue(store).load(), [a]);
  });
}
