import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

void main() {
  final manifest =
      File('android/app/src/main/AndroidManifest.xml').readAsStringSync();
  final debugManifest =
      File('android/app/src/debug/AndroidManifest.xml').readAsStringSync();
  final gradle = File('android/app/build.gradle.kts').readAsStringSync();

  test('manifest grants INTERNET', () {
    expect(manifest, contains('android.permission.INTERNET'));
  });

  test('manifest accepts shared page images from the Boox notes app', () {
    expect(manifest, contains('android.intent.action.SEND'));
    expect(manifest, contains('android.intent.action.SEND_MULTIPLE'));
    expect(manifest, contains('android:mimeType="image/*"'));
  });

  test('release manifest denies cleartext HTTP', () {
    expect(manifest, contains('android:usesCleartextTraffic="false"'));
    expect(manifest, isNot(contains('android:usesCleartextTraffic="true"')));
  });

  test('debug manifest permits cleartext HTTP for self-hosted backends', () {
    expect(debugManifest, contains('android:usesCleartextTraffic="true"'));
  });

  test('release build wires key.properties signing with a debug fallback', () {
    expect(gradle, contains('key.properties'));
    expect(gradle, contains('releaseSigningConfig'));
    expect(gradle, contains('signingConfigs.getByName("debug")'));
  });
}
