import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';

class LoadingState extends StatelessWidget {
  final String? message;
  const LoadingState({super.key, this.message});
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(
              width: 36,
              height: 36,
              child: CircularProgressIndicator(strokeWidth: 3),
            ),
            if (message != null) ...[
              const SizedBox(height: 16),
              Text(
                message!,
                style: GoogleFonts.lora(
                  fontSize: 14,
                  color: VillumTheme.parchmentMuted,
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class ErrorState extends StatelessWidget {
  final String message;
  final VoidCallback? onRetry;
  const ErrorState({super.key, required this.message, this.onRetry});
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: VillumTheme.hpDying.withValues(alpha: 0.14),
                shape: BoxShape.circle,
                border: Border.all(
                  color: VillumTheme.hpDying.withValues(alpha: 0.3),
                ),
              ),
              child: const Icon(
                Icons.warning_amber_rounded,
                color: VillumTheme.hpDying,
                size: 28,
              ),
            ),
            const SizedBox(height: 14),
            Text(
              'Something went wrong',
              style: GoogleFonts.cinzel(
                fontSize: 15,
                fontWeight: FontWeight.w700,
                color: VillumTheme.parchment,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              message,
              style: GoogleFonts.lora(
                fontSize: 13,
                color: VillumTheme.parchmentMuted,
              ),
              textAlign: TextAlign.center,
            ),
            if (onRetry != null) ...[
              const SizedBox(height: 18),
              FilledButton.icon(
                onPressed: onRetry,
                icon: const Icon(Icons.refresh, size: 18),
                label: const Text('Retry'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class EmptyState extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final Widget? action;
  const EmptyState({
    super.key,
    required this.icon,
    required this.title,
    required this.subtitle,
    this.action,
  });
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 64,
              height: 64,
              decoration: BoxDecoration(
                color: VillumTheme.surfaceCard,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: VillumTheme.inkLight),
              ),
              child: Icon(icon, size: 32, color: VillumTheme.parchmentMuted),
            ),
            const SizedBox(height: 16),
            Text(
              title,
              style: GoogleFonts.cinzel(
                fontSize: 15,
                fontWeight: FontWeight.w700,
                color: VillumTheme.parchment,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 6),
            Text(
              subtitle,
              style: GoogleFonts.lora(
                fontSize: 13,
                color: VillumTheme.parchmentMuted,
              ),
              textAlign: TextAlign.center,
            ),
            if (action != null) ...[const SizedBox(height: 18), action!],
          ],
        ),
      ),
    );
  }
}

class PermissionState extends StatelessWidget {
  final String title;
  final String message;
  final IconData icon;
  const PermissionState({
    super.key,
    this.title = 'DM only',
    this.message = 'Only the Dungeon Master can view this screen.',
    this.icon = Icons.shield_outlined,
  });
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              size: 40,
              color: VillumTheme.gold.withValues(alpha: 0.7),
            ),
            const SizedBox(height: 12),
            Text(
              title,
              style: GoogleFonts.cinzel(
                fontSize: 16,
                fontWeight: FontWeight.w700,
                color: VillumTheme.parchment,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: GoogleFonts.lora(
                fontSize: 13,
                color: VillumTheme.parchmentMuted,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}
