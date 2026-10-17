import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../app/theme.dart';

class StatTile extends StatelessWidget {
  final String label;
  final int score;
  final int modifier;
  final VoidCallback? onTap;
  const StatTile({
    super.key,
    required this.label,
    required this.score,
    required this.modifier,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: VillumTheme.surfaceCard,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 6),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: VillumTheme.inkLight),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                label.toUpperCase(),
                style: GoogleFonts.cinzel(
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 0.8,
                  color: VillumTheme.gold,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                '$score',
                style: GoogleFonts.cinzel(
                  fontSize: 18,
                  fontWeight: FontWeight.w700,
                  color: VillumTheme.parchment,
                ),
              ),
              const SizedBox(height: 2),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: VillumTheme.gold.withValues(alpha: 0.14),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  modifier >= 0 ? '+$modifier' : '$modifier',
                  style: GoogleFonts.lora(
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                    color: VillumTheme.goldLight,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class ConditionChip extends StatelessWidget {
  final String label;
  final bool active;
  final VoidCallback? onTap;
  final VoidCallback? onRemove;
  const ConditionChip({
    super.key,
    required this.label,
    this.active = true,
    this.onTap,
    this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: onTap != null,
      label: 'Condition $label',
      child: Material(
        color: active
            ? VillumTheme.gold.withValues(alpha: 0.16)
            : VillumTheme.surfaceCard,
        borderRadius: BorderRadius.circular(20),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(20),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(20),
              border: Border.all(
                color: active
                    ? VillumTheme.gold.withValues(alpha: 0.4)
                    : VillumTheme.inkLight,
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 6,
                  height: 6,
                  decoration: BoxDecoration(
                    color: active
                        ? VillumTheme.gold
                        : VillumTheme.parchmentMuted,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 6),
                Text(
                  label,
                  style: GoogleFonts.lora(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: VillumTheme.parchment,
                  ),
                ),
                if (onRemove != null) ...[
                  const SizedBox(width: 6),
                  InkWell(
                    onTap: onRemove,
                    child: const Icon(
                      Icons.close,
                      size: 14,
                      color: VillumTheme.parchmentMuted,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class SlotPips extends StatelessWidget {
  final int level;
  final int max;
  final int used;
  final ValueChanged<int>? onExpend;
  final ValueChanged<int>? onRestore;
  const SlotPips({
    super.key,
    required this.level,
    required this.max,
    required this.used,
    this.onExpend,
    this.onRestore,
  });

  @override
  Widget build(BuildContext context) {
    final remaining = (max - used).clamp(0, max);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: VillumTheme.gold.withValues(alpha: 0.14),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: VillumTheme.gold.withValues(alpha: 0.25),
                ),
              ),
              child: Text(
                'LVL $level',
                style: GoogleFonts.cinzel(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  color: VillumTheme.gold,
                ),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              '$remaining / $max',
              style: GoogleFonts.lora(
                fontSize: 12,
                color: VillumTheme.parchmentMuted,
              ),
            ),
            const Spacer(),
            if (remaining > 0 && onExpend != null)
              _smallBtn('Cast', onExpend!, level)
            else if (remaining == 0 && onRestore != null)
              _smallBtn('Restore', onRestore!, level),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: List.generate(max, (i) {
            final isUsed = i < used;
            return Semantics(
              label:
                  'Spell slot level $level, ${isUsed ? 'expended' : 'available'}',
              child: Container(
                width: 18,
                height: 18,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isUsed ? Colors.transparent : VillumTheme.gold,
                  border: Border.all(
                    color: isUsed ? VillumTheme.inkLight : VillumTheme.gold,
                    width: 1.4,
                  ),
                  boxShadow: isUsed
                      ? null
                      : [
                          BoxShadow(
                            color: VillumTheme.gold.withValues(alpha: 0.35),
                            blurRadius: 6,
                            spreadRadius: 1,
                          ),
                        ],
                ),
                child: isUsed
                    ? const Icon(
                        Icons.close,
                        size: 10,
                        color: VillumTheme.parchmentMuted,
                      )
                    : null,
              ),
            );
          }),
        ),
      ],
    );
  }

  Widget _smallBtn(String label, ValueChanged<int> cb, int lvl) => SizedBox(
    height: 32,
    child: OutlinedButton(
      onPressed: () => cb(lvl),
      style: OutlinedButton.styleFrom(
        padding: const EdgeInsets.symmetric(horizontal: 10),
        minimumSize: const Size(48, 32),
        textStyle: GoogleFonts.lora(fontSize: 12, fontWeight: FontWeight.w600),
      ),
      child: Text(label),
    ),
  );
}

class DiceReveal extends StatelessWidget {
  final String title;
  final String value;
  final String? subtitle;
  final Color? accent;
  const DiceReveal({
    super.key,
    required this.title,
    required this.value,
    this.subtitle,
    this.accent,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: VillumTheme.surfaceCard,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: (accent ?? VillumTheme.gold).withValues(alpha: 0.3),
        ),
      ),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: (accent ?? VillumTheme.gold).withValues(alpha: 0.14),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(
              Icons.casino_rounded,
              color: accent ?? VillumTheme.gold,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: GoogleFonts.cinzel(
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 0.6,
                    color: VillumTheme.gold,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  value,
                  style: GoogleFonts.cinzel(
                    fontSize: 18,
                    fontWeight: FontWeight.w800,
                    color: VillumTheme.parchment,
                  ),
                ),
                if (subtitle != null)
                  Text(
                    subtitle!,
                    style: GoogleFonts.lora(
                      fontSize: 12,
                      color: VillumTheme.parchmentMuted,
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
