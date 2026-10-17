import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';

class HpBar extends StatelessWidget {
  final int current;
  final int max;
  final int tempHp;
  final double height;
  const HpBar({
    super.key,
    required this.current,
    required this.max,
    this.tempHp = 0,
    this.height = 14,
  });

  @override
  Widget build(BuildContext context) {
    final ratio = max <= 0 ? 0.0 : (current / max).clamp(0.0, 1.0);
    final color = VillumTheme.hpColor(ratio);
    final tempRatio = max <= 0 ? 0.0 : (tempHp / max).clamp(0.0, 1.0);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: SizedBox(
            height: height,
            child: Stack(
              children: [
                Container(color: VillumTheme.inkMid),
                FractionallySizedBox(
                  widthFactor: ratio,
                  child: Container(
                    decoration: BoxDecoration(
                      color: color,
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),
                if (tempHp > 0)
                  FractionallySizedBox(
                    widthFactor: (ratio + tempRatio).clamp(0.0, 1.0),
                    child: Container(
                      decoration: BoxDecoration(
                        color: VillumTheme.hpTemp.withValues(alpha: 0.45),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: VillumTheme.hpTemp.withValues(alpha: 0.6),
                          width: 1,
                        ),
                      ),
                    ),
                  ),
                // subtle highlight
                Positioned.fill(
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: Colors.white.withValues(alpha: 0.06),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 6),
        Row(
          children: [
            Icon(
              ratio <= 0
                  ? Icons.heart_broken
                  : ratio < 0.5
                  ? Icons.favorite_border
                  : Icons.favorite,
              size: 14,
              color: color,
              semanticLabel: ratio <= 0
                  ? 'Dying'
                  : ratio < 0.5
                  ? 'Bloodied'
                  : 'Healthy',
            ),
            const SizedBox(width: 6),
            Text(
              '$current / $max',
              style: GoogleFonts.lora(
                fontSize: 13,
                fontWeight: FontWeight.w700,
                color: VillumTheme.parchment,
              ),
            ),
            if (tempHp > 0) ...[
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: VillumTheme.hpTemp.withValues(alpha: 0.18),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: VillumTheme.hpTemp.withValues(alpha: 0.35),
                  ),
                ),
                child: Text(
                  '+$tempHp temp',
                  style: GoogleFonts.lora(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: VillumTheme.hpTemp,
                  ),
                ),
              ),
            ],
            const Spacer(),
            Text(
              ratio <= 0
                  ? 'DYING'
                  : ratio < 0.5
                  ? 'BLOODIED'
                  : 'HEALTHY',
              style: GoogleFonts.cinzel(
                fontSize: 10,
                fontWeight: FontWeight.w800,
                letterSpacing: 0.8,
                color: color,
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class HpStepper extends StatelessWidget {
  final Future<void> Function(int delta, {bool temp}) onAdjust;
  const HpStepper({super.key, required this.onAdjust});

  @override
  Widget build(BuildContext context) {
    Widget btn(int delta, {bool temp = false}) => SizedBox(
      height: 48,
      child: temp
          ? OutlinedButton(
              onPressed: () {
                HapticFeedback.lightImpact();
                onAdjust(delta, temp: true);
              },
              style: OutlinedButton.styleFrom(
                foregroundColor: VillumTheme.hpTemp,
                side: BorderSide(
                  color: VillumTheme.hpTemp.withValues(alpha: 0.5),
                ),
              ),
              child: Text(
                '+${delta}T',
                semanticsLabel: 'Add $delta temporary hit points',
              ),
            )
          : FilledButton(
              onPressed: () {
                HapticFeedback.mediumImpact();
                onAdjust(delta);
              },
              style: FilledButton.styleFrom(
                backgroundColor: delta < 0
                    ? VillumTheme.hpDying
                    : VillumTheme.hpHealthy,
                foregroundColor: Colors.white,
                minimumSize: const Size(48, 48),
              ),
              child: Text(
                delta > 0 ? '+$delta' : '$delta',
                style: const TextStyle(fontWeight: FontWeight.w800),
              ),
            ),
    );

    return Column(
      children: [
        Row(
          children: [
            Expanded(child: btn(-10)),
            const SizedBox(width: 8),
            Expanded(child: btn(-5)),
            const SizedBox(width: 8),
            Expanded(child: btn(-1)),
            const SizedBox(width: 8),
            Expanded(child: btn(1)),
            const SizedBox(width: 8),
            Expanded(child: btn(5)),
          ],
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(child: btn(10)),
            const SizedBox(width: 8),
            Expanded(child: btn(1, temp: true)),
            const SizedBox(width: 8),
            Expanded(child: btn(5, temp: true)),
            const SizedBox(width: 8),
            const Expanded(child: SizedBox()),
            const SizedBox(width: 8),
            const Expanded(child: SizedBox()),
          ],
        ),
        const SizedBox(height: 4),
        Text(
          'Tap to damage or heal. +T adds temporary HP.',
          style: GoogleFonts.lora(
            fontSize: 11,
            color: VillumTheme.parchmentMuted,
          ),
        ),
      ],
    );
  }
}

class DeathSaveTracker extends StatelessWidget {
  final int successes;
  final int failures;
  final VoidCallback? onSuccess;
  final VoidCallback? onFailure;
  const DeathSaveTracker({
    super.key,
    required this.successes,
    required this.failures,
    this.onSuccess,
    this.onFailure,
  });

  @override
  Widget build(BuildContext context) {
    Widget pip(bool filled, Color color) => Container(
      width: 22,
      height: 22,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: filled ? color : Colors.transparent,
        border: Border.all(color: color, width: 1.4),
      ),
      child: filled
          ? Icon(Icons.check, size: 12, color: VillumTheme.ink)
          : null,
    );

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: VillumTheme.hpDying.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: VillumTheme.hpDying.withValues(alpha: 0.35)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              const Icon(
                Icons.warning_rounded,
                size: 18,
                color: VillumTheme.hpDying,
              ),
              const SizedBox(width: 8),
              Text(
                'DEATH SAVES',
                style: GoogleFonts.cinzel(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1,
                  color: VillumTheme.hpDying,
                ),
              ),
              const Spacer(),
              Text(
                '$successes successes · $failures failures',
                style: GoogleFonts.lora(
                  fontSize: 11,
                  color: VillumTheme.parchmentMuted,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Row(
            children: [
              Text(
                'Success',
                style: GoogleFonts.lora(
                  fontSize: 12,
                  color: VillumTheme.parchmentMuted,
                ),
              ),
              const SizedBox(width: 8),
              for (int i = 0; i < 3; i++) ...[
                pip(i < successes, VillumTheme.hpHealthy),
                const SizedBox(width: 6),
              ],
              const Spacer(),
              Text(
                'Failure',
                style: GoogleFonts.lora(
                  fontSize: 12,
                  color: VillumTheme.parchmentMuted,
                ),
              ),
              const SizedBox(width: 8),
              for (int i = 0; i < 3; i++) ...[
                pip(i < failures, VillumTheme.hpDying),
                if (i < 2) const SizedBox(width: 6),
              ],
            ],
          ),
          if (onSuccess != null || onFailure != null) ...[
            const SizedBox(height: 10),
            Row(
              children: [
                if (onSuccess != null)
                  Expanded(
                    child: OutlinedButton(
                      onPressed: onSuccess,
                      style: OutlinedButton.styleFrom(
                        foregroundColor: VillumTheme.hpHealthy,
                      ),
                      child: const Text('Add success'),
                    ),
                  ),
                if (onSuccess != null && onFailure != null)
                  const SizedBox(width: 8),
                if (onFailure != null)
                  Expanded(
                    child: OutlinedButton(
                      onPressed: onFailure,
                      style: OutlinedButton.styleFrom(
                        foregroundColor: VillumTheme.hpDying,
                      ),
                      child: const Text('Add failure'),
                    ),
                  ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}
