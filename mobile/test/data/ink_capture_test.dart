import 'package:dnd/data/models/ink_capture.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('parses pages, joins text and reports promotion', () {
    final capture = InkCapture.fromJson({
      'id': 4,
      'campaign_id': 2,
      'character_id': null,
      'owner_user_id': 9,
      'title': 'Session 12',
      'source': 'share',
      'status': 'transcribed',
      'pages': [
        {'id': 10, 'page_index': 0, 'text': 'First page', 'ocr_source': 'boox'},
        {'id': 11, 'page_index': 1, 'text': '  ', 'ocr_source': 'none'},
        {
          'id': 12,
          'page_index': 2,
          'text': 'Third page',
          'ocr_source': 'vision',
        },
      ],
    });

    expect(capture.id, 4);
    expect(capture.pages, hasLength(3));
    expect(capture.text, 'First page\n\nThird page');
    expect(capture.isPromoted, isFalse);

    final promoted = InkCapture.fromJson({
      'id': 5,
      'campaign_id': 2,
      'recap_id': 8,
      'pages': const [],
    });
    expect(promoted.isPromoted, isTrue);
    expect(promoted.text, isEmpty);
  });
}
