import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dnd/app/theme.dart';

void main() {
  testWidgets('dark scheme is dark and gold accent present', (tester) async {
    final theme = VillumTheme.dark;
    expect(theme.colorScheme.brightness, Brightness.dark);
    expect(theme.colorScheme.primary, VillumTheme.gold);
    // Not material default blue
    expect(theme.colorScheme.primary, isNot(const Color(0xFF2196F3)));
    // Verify scaffold background is ink
    expect(theme.scaffoldBackgroundColor, VillumTheme.ink);
    // Smoke: Theme.of returns dark
    await tester.pumpWidget(
      MaterialApp(
        theme: theme,
        home: Builder(
          builder: (c) {
            final cs = Theme.of(c).colorScheme;
            expect(cs.brightness, Brightness.dark);
            return const SizedBox();
          },
        ),
      ),
    );
  });
}
