import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

/// Dark tabletop / parchment theme.
/// Gold highlight is #b8963e — the backend battlemap token default.
class VillumTheme {
  // Palette — warm ink on leather.
  static const Color gold = Color(0xFFB8963E);
  static const Color goldLight = Color(0xFFD4B86A);
  static const Color goldDark = Color(0xFF8A6E2A);
  static const Color ink = Color(0xFF1A1410);
  static const Color inkMid = Color(0xFF2A2520);
  static const Color inkLight = Color(0xFF3D352D);
  static const Color parchment = Color(0xFFF5ECD7);
  static const Color parchmentMuted = Color(0xFFE8DDC8);
  static const Color parchmentDark = Color(0xFFD5C9A8);
  static const Color surfaceDark = Color(0xFF211C17);
  static const Color surfaceCard = Color(0xFF2E281F);
  static const Color surfaceElevated = Color(0xFF3A3328);

  // Semantic HP
  static const Color hpHealthy = Color(0xFF5CB85C);
  static const Color hpBloodied = Color(0xFFE8A838);
  static const Color hpDying = Color(0xFFE53935);
  static const Color hpTemp = Color(0xFF64B5F6);

  // Conditions
  static const Color conditionBg = Color(0xFF3A3328);
  static const Color conditionFg = Color(0xFFF5ECD7);

  static Color hpColor(double ratio) {
    if (ratio <= 0) return hpDying;
    if (ratio < 0.5) return hpBloodied;
    return hpHealthy;
  }

  static ThemeData get dark {
    final base = ThemeData.dark();

    final colorScheme = const ColorScheme.dark(
      primary: gold,
      onPrimary: ink,
      primaryContainer: goldDark,
      onPrimaryContainer: parchment,
      secondary: parchmentMuted,
      onSecondary: ink,
      secondaryContainer: inkLight,
      onSecondaryContainer: parchment,
      tertiary: goldLight,
      onTertiary: ink,
      surface: surfaceDark,
      onSurface: parchment,
      surfaceContainerHighest: surfaceElevated,
      onSurfaceVariant: parchmentMuted,
      outline: Color(0xFF6B5E4F),
      outlineVariant: Color(0xFF4A4237),
      error: hpDying,
      onError: parchment,
      scrim: Color(0x99000000),
    );

    final textTheme = base.textTheme.copyWith(
      displayLarge: GoogleFonts.cinzel(
        fontSize: 32,
        fontWeight: FontWeight.w800,
        letterSpacing: -0.5,
        color: parchment,
      ),
      displayMedium: GoogleFonts.cinzel(
        fontSize: 26,
        fontWeight: FontWeight.w700,
        letterSpacing: -0.3,
        color: parchment,
      ),
      displaySmall: GoogleFonts.cinzel(
        fontSize: 20,
        fontWeight: FontWeight.w600,
        letterSpacing: 0.2,
        color: parchment,
      ),
      headlineLarge: GoogleFonts.cinzel(
        fontSize: 22,
        fontWeight: FontWeight.w700,
        color: parchment,
      ),
      headlineMedium: GoogleFonts.cinzel(
        fontSize: 18,
        fontWeight: FontWeight.w600,
        color: parchment,
      ),
      headlineSmall: GoogleFonts.cinzel(
        fontSize: 16,
        fontWeight: FontWeight.w600,
        color: parchment,
      ),
      titleLarge: GoogleFonts.lora(
        fontSize: 18,
        fontWeight: FontWeight.w600,
        color: parchment,
      ),
      titleMedium: GoogleFonts.lora(
        fontSize: 16,
        fontWeight: FontWeight.w600,
        color: parchmentMuted,
      ),
      titleSmall: GoogleFonts.lora(
        fontSize: 14,
        fontWeight: FontWeight.w500,
        color: parchmentMuted,
      ),
      bodyLarge: GoogleFonts.lora(fontSize: 15, height: 1.45, color: parchment),
      bodyMedium: GoogleFonts.lora(
        fontSize: 14,
        height: 1.4,
        color: parchmentMuted,
      ),
      bodySmall: GoogleFonts.lora(
        fontSize: 12,
        height: 1.35,
        color: parchmentMuted,
      ),
      labelLarge: GoogleFonts.lora(
        fontSize: 14,
        fontWeight: FontWeight.w600,
        letterSpacing: 0.4,
        color: parchment,
      ),
      labelMedium: GoogleFonts.lora(
        fontSize: 12,
        letterSpacing: 0.3,
        color: parchmentMuted,
      ),
      labelSmall: GoogleFonts.lora(
        fontSize: 11,
        letterSpacing: 0.4,
        color: parchmentMuted,
      ),
    );

    // Attach font families to base so widgets without explicit style inherit.
    return base.copyWith(
      colorScheme: colorScheme,
      scaffoldBackgroundColor: ink,
      textTheme: textTheme,
      primaryTextTheme: textTheme,
      appBarTheme: AppBarTheme(
        backgroundColor: surfaceDark,
        foregroundColor: parchment,
        elevation: 0,
        scrolledUnderElevation: 2,
        centerTitle: false,
        titleTextStyle: GoogleFonts.cinzel(
          fontSize: 20,
          fontWeight: FontWeight.w700,
          color: parchment,
          letterSpacing: 0.3,
        ),
        iconTheme: const IconThemeData(color: parchment, size: 24),
      ),
      cardTheme: CardThemeData(
        color: surfaceCard,
        elevation: 2,
        shadowColor: Colors.black54,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(14),
          side: const BorderSide(color: Color(0xFF3D352D), width: 0.8),
        ),
        margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: gold,
          foregroundColor: ink,
          minimumSize: const Size(48, 48),
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
          textStyle: GoogleFonts.lora(
            fontSize: 15,
            fontWeight: FontWeight.w700,
          ),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: gold,
          side: const BorderSide(color: gold, width: 1.2),
          minimumSize: const Size(48, 48),
          padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
          textStyle: GoogleFonts.lora(
            fontSize: 14,
            fontWeight: FontWeight.w600,
          ),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: goldLight,
          minimumSize: const Size(48, 40),
          textStyle: GoogleFonts.lora(
            fontSize: 14,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: surfaceDark,
        indicatorColor: gold.withValues(alpha: 0.22),
        indicatorShape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(14),
        ),
        labelTextStyle: WidgetStateProperty.resolveWith((s) {
          if (s.contains(WidgetState.selected)) {
            return GoogleFonts.lora(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              color: gold,
            );
          }
          return GoogleFonts.lora(
            fontSize: 11,
            fontWeight: FontWeight.w500,
            color: parchmentMuted,
          );
        }),
        iconTheme: WidgetStateProperty.resolveWith((s) {
          if (s.contains(WidgetState.selected)) {
            return const IconThemeData(color: gold, size: 24);
          }
          return const IconThemeData(color: parchmentMuted, size: 22);
        }),
        height: 72,
        elevation: 6,
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: inkMid,
        hintStyle: GoogleFonts.lora(
          color: parchmentMuted.withValues(alpha: 0.6),
          fontSize: 14,
        ),
        labelStyle: GoogleFonts.lora(color: parchmentMuted, fontSize: 13),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 14,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF4A4237)),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF4A4237)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: gold, width: 1.6),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: hpDying),
        ),
        prefixIconColor: parchmentMuted,
        suffixIconColor: parchmentMuted,
      ),
      chipTheme: ChipThemeData(
        backgroundColor: conditionBg,
        selectedColor: gold.withValues(alpha: 0.2),
        disabledColor: inkMid,
        labelStyle: GoogleFonts.lora(fontSize: 12, color: parchment),
        secondaryLabelStyle: GoogleFonts.lora(fontSize: 12, color: parchment),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
          side: const BorderSide(color: Color(0xFF4A4237)),
        ),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      ),
      sliderTheme: SliderThemeData(
        activeTrackColor: gold,
        inactiveTrackColor: inkLight,
        thumbColor: gold,
        overlayColor: gold.withValues(alpha: 0.18),
        valueIndicatorColor: gold,
        valueIndicatorTextStyle: GoogleFonts.lora(
          color: ink,
          fontSize: 12,
          fontWeight: FontWeight.w700,
        ),
      ),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: surfaceElevated,
        contentTextStyle: GoogleFonts.lora(color: parchment, fontSize: 14),
        actionTextColor: gold,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        behavior: SnackBarBehavior.floating,
        elevation: 6,
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: surfaceCard,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        titleTextStyle: GoogleFonts.cinzel(
          fontSize: 18,
          fontWeight: FontWeight.w700,
          color: parchment,
        ),
        contentTextStyle: GoogleFonts.lora(fontSize: 14, color: parchmentMuted),
      ),
      listTileTheme: ListTileThemeData(
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
        titleTextStyle: GoogleFonts.lora(
          fontSize: 15,
          fontWeight: FontWeight.w600,
          color: parchment,
        ),
        subtitleTextStyle: GoogleFonts.lora(
          fontSize: 12,
          color: parchmentMuted,
        ),
        iconColor: parchmentMuted,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      ),
      dividerTheme: const DividerThemeData(
        color: Color(0xFF3D352D),
        thickness: 0.8,
        space: 1,
      ),
      progressIndicatorTheme: const ProgressIndicatorThemeData(color: gold),
      iconTheme: const IconThemeData(color: parchmentMuted),
    );
  }

  /// Bonus light parchment theme (off-white paper, ink text).
  static ThemeData get light {
    final base = ThemeData.light();
    const inkOnPaper = Color(0xFF2B2118);
    final colorScheme = const ColorScheme.light(
      primary: goldDark,
      onPrimary: parchment,
      primaryContainer: Color(0xFFF0D9A0),
      onPrimaryContainer: inkOnPaper,
      secondary: Color(0xFF6B5E4F),
      onSecondary: parchment,
      surface: parchment,
      onSurface: inkOnPaper,
      surfaceContainerHighest: Color(0xFFEDE3CC),
      onSurfaceVariant: Color(0xFF5C4F3D),
      outline: Color(0xFF9A8B76),
      outlineVariant: Color(0xFFD5C9A8),
      error: hpDying,
      onError: parchment,
    );
    return base.copyWith(
      colorScheme: colorScheme,
      scaffoldBackgroundColor: parchment,
      textTheme: base.textTheme.copyWith(
        displayLarge: GoogleFonts.cinzel(
          fontSize: 32,
          fontWeight: FontWeight.w800,
          color: inkOnPaper,
        ),
        displayMedium: GoogleFonts.cinzel(
          fontSize: 26,
          fontWeight: FontWeight.w700,
          color: inkOnPaper,
        ),
        headlineMedium: GoogleFonts.cinzel(
          fontSize: 18,
          fontWeight: FontWeight.w600,
          color: inkOnPaper,
        ),
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: parchment,
        foregroundColor: inkOnPaper,
        titleTextStyle: GoogleFonts.cinzel(
          fontSize: 20,
          fontWeight: FontWeight.w700,
          color: inkOnPaper,
        ),
      ),
      cardTheme: CardThemeData(
        color: Colors.white.withValues(alpha: 0.9),
        elevation: 1,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(14),
          side: const BorderSide(color: Color(0xFFE0D5BE)),
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: goldDark,
          foregroundColor: parchment,
          minimumSize: const Size(48, 48),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
    );
  }
}

/// Extension for HP semantic helpers on BuildContext
extension VillumHpContext on BuildContext {
  Color hpColorFor(int current, int max) {
    if (max <= 0) return VillumTheme.hpDying;
    final ratio = current / max;
    return VillumTheme.hpColor(ratio);
  }
}
