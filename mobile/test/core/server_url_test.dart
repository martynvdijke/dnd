import 'package:flutter_test/flutter_test.dart';

import 'package:dnd/core/server_url.dart';

void main() {
  group('normalizeServerUrl', () {
    test('trims surrounding whitespace', () {
      expect(normalizeServerUrl('  http://host:6270  '), 'http://host:6270');
    });

    test('strips trailing slashes', () {
      expect(normalizeServerUrl('https://vin.dev/'), 'https://vin.dev');
      expect(normalizeServerUrl('https://vin.dev///'), 'https://vin.dev');
    });

    test('keeps the scheme separator of a bare scheme', () {
      // Must not degrade "https://" into "https:/" while stripping.
      expect(normalizeServerUrl('https://'), 'https://');
    });

    test('keeps paths and ports untouched', () {
      expect(
        normalizeServerUrl('http://192.168.1.10:6270/dnd/'),
        'http://192.168.1.10:6270/dnd',
      );
    });
  });

  group('serverUrlError', () {
    test('rejects an empty address', () {
      expect(serverUrlError('   '), 'Enter your Villum server address.');
    });

    test('rejects non-http schemes and missing hosts', () {
      const message = 'Enter a valid URL starting with http:// or https://';
      expect(serverUrlError('ftp://example.com'), message);
      expect(serverUrlError('example.com:6270'), message);
      expect(serverUrlError('http://'), message);
      expect(serverUrlError('https://'), message);
    });

    test('accepts http and https addresses', () {
      expect(serverUrlError('http://10.0.2.2:6270'), isNull);
      expect(serverUrlError('https://villum.example.com'), isNull);
      expect(serverUrlError(' https://villum.example.com/ '), isNull);
    });
  });
}
