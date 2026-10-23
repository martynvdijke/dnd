import 'package:dnd/features/ink/ink_canvas.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('renders strokes to a PNG', () async {
    final png = await renderInkPng(
      strokes: [
        [const Offset(10, 10), const Offset(90, 60)],
        [const Offset(40, 30)],
      ],
      size: const Size(120, 80),
    );
    expect(png, isNotNull);
    expect(png!.take(4).toList(), [0x89, 0x50, 0x4E, 0x47]);
  });

  test('refuses to render a zero-size page', () async {
    expect(await renderInkPng(strokes: const [], size: Size.zero), isNull);
  });

  testWidgets('an empty page asks for a stroke before saving', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: InkCanvasPage()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Save page'));
    await tester.pump();

    expect(find.text('Write something first.'), findsOneWidget);
  });

  testWidgets('a stylus stroke is accepted and clears the empty warning', (
    tester,
  ) async {
    await tester.pumpWidget(const MaterialApp(home: InkCanvasPage()));
    await tester.pumpAndSettle();

    final gesture = await tester.startGesture(
      const Offset(120, 220),
      kind: PointerDeviceKind.stylus,
      pointer: 1,
    );
    await gesture.moveTo(const Offset(240, 320));
    await gesture.up();
    await tester.pump();

    await tester.tap(find.byIcon(Icons.check));
    await tester.pump();

    expect(find.text('Write something first.'), findsNothing);
  });
}
