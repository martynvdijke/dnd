/**
 * GSAP-orchestrated animation sequences for high-impact D&D moments.
 *
 * CSS owns element appearance (resting states); GSAP owns motion
 * (timelines, physics-feel easing, multi-element orchestration).
 *
 * All sequences respect `prefers-reduced-motion` via gsap.matchMedia():
 * when reduced motion is requested, animations are skipped (instant state
 * changes) but underlying data updates still happen.
 */
import gsap from 'gsap';

// Shared matchMedia guard — auto-cleans timelines when conditions change.
const mm = gsap.matchMedia();

// Whether the user has requested reduced motion.
// Updated by matchMedia callbacks; checked by animation functions.
let reducedMotion = false;
mm.add('(prefers-reduced-motion: reduce)', () => {
  reducedMotion = true;
  return () => { reducedMotion = false; };
});

export { reducedMotion };

// ─── Crit Celebration ───

/**
 * Play the crit celebration sequence: screen-edge flash, particle burst
 * from the die center, text slam, timed with the existing Web Audio chord.
 *
 * @param critType  'success' (max roll) or 'fail' (nat 1)
 * @param dieLabel  Die label for the text (e.g. 'd20', 'd6')
 * @param originEl  The die element to burst particles from
 */
export function critCelebration(
  critType: 'success' | 'fail',
  dieLabel: string,
  originEl: HTMLElement,
): void {
  if (reducedMotion) return; // sound is handled separately by dice.ts

  const isFail = critType === 'fail';
  const color = isFail ? '#8b0000' : '#ffd700';
  const text = isFail
    ? 'CRITICAL FAIL!'
    : (dieLabel === 'd20' ? 'CRITICAL!' : `Max on ${dieLabel}!`);

  // Get die center coordinates for particle origin
  const rect = originEl.getBoundingClientRect();
  const cx = rect.left + rect.width / 2;
  const cy = rect.top + rect.height / 2;

  // Create overlay (full-screen flash)
  const overlay = document.createElement('div');
  overlay.className = 'crit-overlay';
  overlay.style.background = isFail
    ? 'radial-gradient(circle at center, rgba(139,0,0,0.25), transparent 70%)'
    : 'radial-gradient(circle at center, rgba(255,215,0,0.25), transparent 70%)';
  document.body.appendChild(overlay);

  // Create text slam
  const textEl = document.createElement('div');
  textEl.className = 'crit-text';
  textEl.textContent = text;
  textEl.style.color = color;
  document.body.appendChild(textEl);

  // Create 12 particle sparks
  const sparks: HTMLElement[] = [];
  for (let i = 0; i < 12; i++) {
    const spark = document.createElement('div');
    spark.className = 'crit-spark';
    spark.style.background = color;
    spark.style.left = cx + 'px';
    spark.style.top = cy + 'px';
    document.body.appendChild(spark);
    sparks.push(spark);
  }

  // GSAP timeline: flash → particles → text slam → fadeout → cleanup
  const tl = gsap.timeline({
    onComplete: () => {
      overlay.remove();
      textEl.remove();
      sparks.forEach(s => s.remove());
    },
  });

  // Phase 1: screen flash (0.10s)
  tl.fromTo(overlay, { opacity: 0 }, { opacity: 1, duration: 0.10, ease: 'power2.out' });
  tl.to(overlay, { opacity: 0, duration: 0.40, ease: 'power2.in' });

  // Phase 2: particle burst (0.15s start, staggered)
  sparks.forEach((spark, i) => {
    const angle = (i / 12) * Math.PI * 2;
    const dist = 60 + Math.random() * 40;
    tl.fromTo(spark,
      { x: 0, y: 0, opacity: 1, scale: 1 },
      {
        x: Math.cos(angle) * dist,
        y: Math.sin(angle) * dist,
        opacity: 0,
        scale: 0.3,
        duration: 0.50,
        ease: 'power2.out',
      },
      0.15,
    );
  });

  // Phase 3: text slam (0.20s start, back.ease for impact)
  tl.fromTo(textEl,
    { scale: 0.3, opacity: 0 },
    { scale: 1, opacity: 1, duration: 0.35, ease: 'back.out(1.7)' },
    0.20,
  );
  tl.to(textEl, { opacity: 0, duration: 0.40, ease: 'power1.in' }, 1.10);
}

// ─── HP Damage/Heal Animation ───

/**
 * Animate an HP change: floating damage/heal number, HP value flash,
 * elastic-eased HP bar transition, low-HP pulse trigger.
 *
 * @param hpEl    The element displaying the HP number
 * @param barEl   The HP bar element (width animated)
 * @param oldHp   HP before the change
 * @param newHp   HP after the change
 * @param maxHp   Max HP (for bar width + low-HP threshold)
 */
export function animateHpChange(
  hpEl: HTMLElement,
  barEl: HTMLElement,
  oldHp: number,
  newHp: number,
  maxHp: number,
): void {
  const delta = newHp - oldHp;
  if (delta === 0) return;

  const isDamage = delta < 0;
  const absDelta = Math.abs(delta);

  // Update bar width immediately (GSAP animates the transition)
  const newPct = Math.max(0, Math.min(100, (newHp / maxHp) * 100));

  if (reducedMotion) {
    // Instant update — no animation, but state is correct
    barEl.style.width = newPct + '%';
    applyLowHpPulse(barEl, newHp, maxHp);
    return;
  }

  // Floating number
  const floatEl = document.createElement('div');
  floatEl.className = 'damage-float';
  floatEl.textContent = (isDamage ? '-' : '+') + absDelta;
  floatEl.style.color = isDamage ? '#dc3545' : '#28a745';
  const rect = hpEl.getBoundingClientRect();
  floatEl.style.left = rect.left + rect.width / 2 + 'px';
  floatEl.style.top = rect.top + 'px';
  document.body.appendChild(floatEl);

  // HP value flash
  const flashClass = isDamage ? 'hp-flash-damage' : 'hp-flash-heal';

  const tl = gsap.timeline({
    onComplete: () => {
      floatEl.remove();
      hpEl.classList.remove(flashClass);
    },
  });

  // Floating number flies up + fades
  tl.fromTo(floatEl,
    { y: 0, opacity: 1 },
    { y: -40, opacity: 0, duration: 0.40, ease: 'power2.out' },
  );

  // HP value flash (CSS class triggers color animation)
  hpEl.classList.add(flashClass);

  // HP bar elastic ease
  tl.to(barEl, {
    width: newPct + '%',
    duration: 0.50,
    ease: 'elastic.out(1, 0.6)',
  }, 0);

  // Low-HP pulse trigger
  applyLowHpPulse(barEl, newHp, maxHp);
}

function applyLowHpPulse(barEl: HTMLElement, currentHp: number, maxHp: number): void {
  if (maxHp > 0 && currentHp / maxHp < 0.25) {
    barEl.classList.add('hp-bar-pulse');
  } else {
    barEl.classList.remove('hp-bar-pulse');
  }
}

// ─── Combat Turn Transition ───

/**
 * Animate a combat turn change: dim previous row, slide new active row
 * in from the right with gold glow, name scale pulse, optional monster
 * red tint sweep.
 *
 * @param prevRow   The previously active combatant row
 * @param nextRow   The newly active combatant row
 * @param isMonster Whether the new active combatant is a monster
 */
export function animateTurnChange(
  prevRow: HTMLElement | null,
  nextRow: HTMLElement,
  isMonster: boolean,
): void {
  // Class swaps happen regardless of motion preference (accessibility)
  // Note: prevRow may be null if the DOM was re-rendered
  if (prevRow) {
    prevRow.classList.remove('combatant-row-active');
    prevRow.classList.add('combatant-row-dimmed');
  }
  nextRow.classList.remove('combatant-row-dimmed');
  nextRow.classList.add('combatant-row-active');
  if (isMonster) {
    nextRow.classList.add('turn-tint-monster');
  } else {
    nextRow.classList.remove('turn-tint-monster');
  }

  if (reducedMotion) return; // classes set, no motion

  // Use the second td (name cell) for the name pulse, or the row itself
  const nameEl = (nextRow.querySelector('td:nth-child(2)') as HTMLElement | null) || nextRow;

  const tl = gsap.timeline();

  // Slide new row in from right
  tl.fromTo(nextRow,
    { x: 20, opacity: 0.7 },
    { x: 0, opacity: 1, duration: 0.30, ease: 'power2.out' },
  );

  // Name scale pulse
  tl.fromTo(nameEl,
    { scale: 1 },
    { scale: 1.05, duration: 0.15, ease: 'power2.out', yoyo: true, repeat: 1 },
    0.15,
  );
}
