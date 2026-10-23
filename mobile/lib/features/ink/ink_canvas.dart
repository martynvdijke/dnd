import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';

/// A minimal stylus-first ink page. Strokes are kept as point lists and
/// rendered to PNG on save. Touch input is ignored once a stylus has been seen
/// on the page, which is the palm rejection an e-ink tablet needs.
class InkCanvasPage extends StatefulWidget {
  const InkCanvasPage({super.key, this.initialStrokeWidth = 2.6});

  final double initialStrokeWidth;

  @override
  State<InkCanvasPage> createState() => _InkCanvasPageState();
}

class _InkCanvasPageState extends State<InkCanvasPage> {
  final List<List<Offset>> _strokes = [];
  List<Offset>? _current;
  bool _stylusSeen = false;
  double _strokeWidth = 2.6;
  bool _erasing = false;
  Size _canvasSize = Size.zero;
  bool _saving = false;

  void _start(Offset p, PointerDeviceKind kind) {
    if (kind == PointerDeviceKind.stylus ||
        kind == PointerDeviceKind.invertedStylus) {
      _stylusSeen = true;
    } else if (kind == PointerDeviceKind.touch && _stylusSeen) {
      return;
    }
    if (_erasing) {
      _eraseAt(p);
      return;
    }
    setState(() {
      _current = [p];
      _strokes.add(_current!);
    });
  }

  void _extend(Offset p, PointerDeviceKind kind) {
    if (kind == PointerDeviceKind.touch && _stylusSeen) return;
    if (_erasing) return;
    final current = _current;
    if (current == null) return;
    setState(() => current.add(p));
  }

  void _end() {
    _current = null;
  }

  void _eraseAt(Offset p) {
    const radius = 14.0;
    setState(() {
      _strokes.removeWhere((s) => s.any((q) => (q - p).distance <= radius));
    });
  }

  void _undo() {
    if (_strokes.isEmpty) return;
    setState(() => _strokes.removeLast());
  }

  Future<Uint8List?> _renderPng() => renderInkPng(
    strokes: _strokes,
    size: _canvasSize,
    strokeWidth: _strokeWidth,
  );

  Future<void> _save() async {
    if (_strokes.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Write something first.')));
      return;
    }
    setState(() => _saving = true);
    final png = await _renderPng();
    if (!mounted) return;
    setState(() => _saving = false);
    if (png == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Could not export the page.')),
      );
      return;
    }
    Navigator.of(context).pop<Uint8List>(png);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        foregroundColor: Colors.black,
        elevation: 0,
        title: const Text('Ink page'),
        actions: [
          IconButton(
            icon: Icon(_erasing ? Icons.edit : Icons.auto_fix_high),
            color: Colors.black,
            tooltip: _erasing ? 'Draw' : 'Erase',
            onPressed: () => setState(() => _erasing = !_erasing),
          ),
          IconButton(
            icon: const Icon(Icons.undo),
            color: Colors.black,
            tooltip: 'Undo',
            onPressed: _undo,
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline),
            color: Colors.black,
            tooltip: 'Clear',
            onPressed: () => setState(_strokes.clear),
          ),
          IconButton(
            icon: const Icon(Icons.check),
            color: Colors.black,
            tooltip: 'Save',
            onPressed: _saving ? null : _save,
          ),
        ],
      ),
      body: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            _canvasSize = Size(constraints.maxWidth, constraints.maxHeight);
            return Listener(
              behavior: HitTestBehavior.opaque,
              onPointerDown: (e) => _start(e.localPosition, e.kind),
              onPointerMove: (e) => _extend(e.localPosition, e.kind),
              onPointerUp: (_) => _end(),
              onPointerCancel: (_) => _end(),
              child: SizedBox.expand(
                child: CustomPaint(
                  painter: _InkPainter(
                    strokes: _strokes,
                    strokeWidth: _strokeWidth,
                  ),
                ),
              ),
            );
          },
        ),
      ),
      bottomNavigationBar: Container(
        color: Colors.white,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          children: [
            const Text('Stroke'),
            Expanded(
              child: Slider(
                value: _strokeWidth,
                min: 1,
                max: 8,
                divisions: 7,
                activeColor: Colors.black,
                inactiveColor: Colors.black26,
                onChanged: (v) => setState(() => _strokeWidth = v),
              ),
            ),
            FilledButton(
              style: FilledButton.styleFrom(
                backgroundColor: Colors.black,
                foregroundColor: Colors.white,
                minimumSize: const Size(96, 56),
              ),
              onPressed: _saving ? null : _save,
              child: Text(_saving ? 'Saving…' : 'Save page'),
            ),
          ],
        ),
      ),
    );
  }
}

/// Renders ink strokes to a white PNG. Exposed separately from the widget so
/// export can be verified without pumping a full route.
Future<Uint8List?> renderInkPng({
  required List<List<Offset>> strokes,
  required Size size,
  double strokeWidth = 2.6,
}) async {
  if (size == Size.zero) return null;
  final recorder = ui.PictureRecorder();
  final canvas = Canvas(recorder);
  canvas.drawRect(
    Rect.fromLTWH(0, 0, size.width, size.height),
    Paint()..color = const Color(0xFFFFFFFF),
  );
  final paint = Paint()
    ..color = const Color(0xFF000000)
    ..strokeWidth = strokeWidth
    ..strokeCap = StrokeCap.round
    ..strokeJoin = StrokeJoin.round
    ..style = PaintingStyle.stroke;
  for (final stroke in strokes) {
    if (stroke.isEmpty) continue;
    if (stroke.length == 1) {
      canvas.drawCircle(
        stroke.first,
        strokeWidth / 2,
        Paint()..color = const Color(0xFF000000),
      );
      continue;
    }
    final path = Path()..moveTo(stroke.first.dx, stroke.first.dy);
    for (final p in stroke.skip(1)) {
      path.lineTo(p.dx, p.dy);
    }
    canvas.drawPath(path, paint);
  }
  final picture = recorder.endRecording();
  final image = await picture.toImage(size.width.round(), size.height.round());
  final data = await image.toByteData(format: ui.ImageByteFormat.png);
  image.dispose();
  return data?.buffer.asUint8List();
}

class _InkPainter extends CustomPainter {
  _InkPainter({required this.strokes, required this.strokeWidth});
  final List<List<Offset>> strokes;
  final double strokeWidth;

  @override
  void paint(Canvas canvas, Size size) {
    canvas.drawRect(
      Offset.zero & size,
      Paint()..color = const Color(0xFFFFFFFF),
    );
    final paint = Paint()
      ..color = const Color(0xFF000000)
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round
      ..style = PaintingStyle.stroke;
    for (final stroke in strokes) {
      if (stroke.isEmpty) continue;
      if (stroke.length == 1) {
        canvas.drawCircle(
          stroke.first,
          strokeWidth / 2,
          Paint()..color = const Color(0xFF000000),
        );
        continue;
      }
      final path = Path()..moveTo(stroke.first.dx, stroke.first.dy);
      for (final p in stroke.skip(1)) {
        path.lineTo(p.dx, p.dy);
      }
      canvas.drawPath(path, paint);
    }
  }

  @override
  bool shouldRepaint(covariant _InkPainter oldDelegate) => true;
}
