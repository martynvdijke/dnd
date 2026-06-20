/**
 * Dice rolling module — clean 2D dice rendering, rolling animation, history.
 *
 * Replaces the broken CSS-3D polyhedron approach (which couldn't properly
 * render d4/d8/d10/d12/d20 faces) with a polished 2D animation: dice tumble
 * with CSS keyframes while numbers cycle rapidly via JS, then settle with a
 * satisfying pop-in when the server result arrives.
 */
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import { getCurrentView } from './navigation';
import { currentChar } from './lib/state';
import { critCelebration } from './lib/animations';

// ─── Sound Effects (Web Audio API — no asset files needed) ───

let audioCtx: AudioContext | null = null;
let soundEnabled = typeof localStorage !== 'undefined' && localStorage.getItem('diceSoundEnabled') !== 'false';

function getAudioCtx(): AudioContext | null {
  if (!audioCtx) {
    try {
      audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)();
    } catch { return null; }
  }
  return audioCtx;
}

/** Synthesized dice clatter sound. Plays a triumphant chord on crits. */
function playDiceSound(isCrit: boolean = false) {
  if (!soundEnabled) return;
  const ctx = getAudioCtx();
  if (!ctx) return;
  if (ctx.state === 'suspended') ctx.resume();

  // Dice clatter: multiple short noise bursts
  const numClacks = isCrit ? 8 : 5;
  for (let i = 0; i < numClacks; i++) {
    const t = ctx.currentTime + i * 0.06 + Math.random() * 0.02;
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = 'square';
    osc.frequency.value = 180 + Math.random() * 400;
    gain.gain.setValueAtTime(0, t);
    gain.gain.linearRampToValueAtTime(0.12, t + 0.005);
    gain.gain.exponentialRampToValueAtTime(0.001, t + 0.05);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start(t);
    osc.stop(t + 0.05);
  }

  if (isCrit) {
    // Triumphant major chord for crits
    const t = ctx.currentTime + 0.35;
    [523.25, 659.25, 783.99].forEach((freq) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = 'triangle';
      osc.frequency.value = freq;
      gain.gain.setValueAtTime(0, t);
      gain.gain.linearRampToValueAtTime(0.08, t + 0.05);
      gain.gain.exponentialRampToValueAtTime(0.001, t + 0.6);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start(t);
      osc.stop(t + 0.6);
    });
  }
}

export function toggleDiceSound() {
  soundEnabled = !soundEnabled;
  try { localStorage.setItem('diceSoundEnabled', String(soundEnabled)); } catch { /* ignore */ }
  const btn = document.getElementById('diceSoundToggle');
  if (btn) {
    btn.innerHTML = soundEnabled
      ? '<i class="fa-solid fa-volume-high"></i>'
      : '<i class="fa-solid fa-volume-xmark"></i>';
  }
}

// ─── Dice Constants ───

const DICE_PRESETS = ['d4', 'd6', 'd8', 'd10', 'd12', 'd20', 'd100'];
const DICE_NOTATION_PRESETS = [
  { label: 'd20', expr: '1d20' },
  { label: 'Advantage', expr: '2d20kh1', icon: 'fa-solid fa-angles-up' },
  { label: 'Disadvantage', expr: '2d20kl1', icon: 'fa-solid fa-angles-down' },
  { label: 'd6', expr: '1d6' },
  { label: '2d6', expr: '2d6' },
  { label: '3d6', expr: '3d6' },
  { label: '4d6kh3', expr: '4d6kh3', sub: 'stats' },
  { label: 'd8', expr: '1d8' },
  { label: 'd10', expr: '1d10' },
  { label: 'd12', expr: '1d12' },
  { label: 'd100', expr: '1d100' },
  { label: 'd4', expr: '1d4' },
];

// Active rolling timeouts (cleared on settle)
let rollingTimeouts: number[] = [];

// ─── Die Value Helpers ───

export function rollValue(r: any): number {
  if (r == null) return 0;
  return typeof r === 'number' ? r : (r.value ?? 0);
}

export function rollUsed(r: any): boolean {
  return typeof r === 'number' ? true : (r.useInTotal !== false);
}

export function rollFlags(r: any): string {
  return typeof r === 'number' ? '' : (r.modifierFlags || '');
}

export function parseSides(dieLabel: string): number {
  const m = dieLabel.match(/^d(\d+)$/i);
  return m ? parseInt(m[1]) : 0;
}

/**
 * Detect crit success/fail for any die type.
 * Max roll on a d4+ die = success, nat 1 = fail.
 * Returns null for non-dice or unused rolls.
 */
export function isCritRoll(value: number, sides: number, used: boolean): 'success' | 'fail' | null {
  if (!used || sides < 4) return null;
  if (value === sides) return 'success';
  if (value === 1) return 'fail';
  return null;
}

// ─── Build 2D Die HTML ───

/**
 * Build a single 2D die element.
 * @param value    The face value to display.
 * @param sides    Number of sides (determines color theme + shape).
 * @param dieLabel Label shown above the die (e.g. "d20").
 * @param extraClass  Extra CSS classes: 'rolling', 'settled', 'die-dropped', 'dice-crit-success', etc.
 */
export function buildDie(value: number, sides: number, dieLabel: string, extraClass: string = ''): string {
  const dieClass = 'd' + (sides >= 100 ? 100 : sides);
  const displayVal = sides >= 100 && value === 100 ? '00' : String(value);
  return `<div class="die-wrapper">
  <span class="die-label">${dieLabel}</span>
  <div class="die ${dieClass} ${extraClass}" data-sides="${sides}" data-value="${value}">
    <span class="die-value">${displayVal}</span>
  </div>
</div>`;
}

// ─── Dice Expression / Quick Roll ───

export function setDiceExpr(expr: string) {
  const input = document.getElementById('diceExpr') as HTMLInputElement;
  input.value = expr;
  doRoll();
}

export async function rollWithAdvantage(isAdv: boolean) {
  const input = document.getElementById('diceExpr') as HTMLInputElement;
  const expr = input.value.trim();
  if (!expr.match(/^\d*d\d+/)) return;

  const resultEl = document.getElementById('diceResult');
  if (resultEl) resultEl.style.display = 'none';

  // Start tumbling animation immediately
  const m = expr.match(/(\d+)d(\d+)/i);
  if (m) {
    const count = parseInt(m[1] || '1');
    const sides = parseInt(m[2]);
    const fakeBreakdown = [{ die: 'd' + sides, rolls: Array(count).fill(1), total: 0 }];
    animateDiceRoll(fakeBreakdown);
  }

  try {
    const result = await api('POST', '/api/roll', {
      expression: expr,
      character_id: currentChar?.id,
      advantage: isAdv ? 'advantage' : 'disadvantage',
    });

    if (result.breakdown) {
      settleDice(result.breakdown);
    }

    if (resultEl) {
      const rolls = result.breakdown?.[0]?.rolls || [];
      const chosen = result.total;
      let badge = '';
      const d20Rolls = result.breakdown?.filter((b: any) => b.die === 'd20' && b.rolls) || [];
      for (const bg of d20Rolls) {
        for (const r of bg.rolls) {
          const v = rollValue(r);
          const used = rollUsed(r);
          const crit = isCritRoll(v, 20, used);
          if (crit === 'success') badge = '<span class="badge bg-success ms-2">Critical Hit!</span>';
          else if (crit === 'fail') badge = '<span class="badge bg-danger ms-2">Critical Fail!</span>';
        }
      }

      setTimeout(() => {
        resultEl.style.display = 'block';
        resultEl.innerHTML = `
          <div class="dice-result-box text-center">
            <div class="roll-expression">${esc(result.expression)} (${isAdv ? 'advantage' : 'disadvantage'})</div>
            <div class="d-flex justify-content-center gap-3 mb-2">
              ${rolls.map((r: any, i: number) => {
                const v = rollValue(r);
                const used = rollUsed(r);
                const style = used ? 'border-color:var(--gold);box-shadow:0 0 0 2px var(--gold)' : 'opacity:0.4';
                return `<span class="die-face${used ? '' : ' die-dropped'}" style="${style}">${v}</span>`;
              }).join('')}
            </div>
            <div class="roll-total-anim">${chosen}</div>
            ${badge}
            <div class="roll-text text-muted">${esc(result.text)}</div>
          </div>`;
      }, 500);
    }
    loadDiceHistory();
  } catch (e: any) {
    toast(e.message, true);
  }
}

// ─── Render Dice Tab ───

export function renderDiceTab() {
  const targetId = getCurrentView() === 'dice' ? 'diceViewSection' : 'diceSection';
  const el = document.getElementById(targetId);
  if (!el) return;
  el.innerHTML = `
    <div class="text-center dice-roller">
      <h5>Dice Roller</h5>
      <div class="row justify-content-center mb-2">
        <div class="col-md-8">
          <label class="form-label">Expression</label>
          <input class="form-control text-center" id="diceExpr" value="1d20" placeholder="e.g. 2d6+3, 4d6kh3, 1d20!" style="font-size:1.3rem;font-weight:700">
        </div>
      </div>
      <div class="dice-quick-btns mb-3">
        ${DICE_NOTATION_PRESETS.map(p =>
          `<button class="btn btn-sm dice-btn" onclick="setDiceExpr('${esc(p.expr)}')" title="${p.sub ? p.sub : p.expr}">
            ${p.icon ? `<i class="${p.icon} me-1"></i>` : ''}${esc(p.label)}
          </button>`
        ).join('')}
      </div>
      <div id="dice3dContainer" class="dice-container"></div>
      <div id="diceResult" class="mb-3" style="display:none"></div>
      <button class="btn btn-gold" onclick="doRoll()"><i class="fa-solid fa-dice me-2"></i>Roll the Bones</button>
      <button class="btn btn-sm btn-outline-secondary dice-sound-toggle ms-2" id="diceSoundToggle" onclick="toggleDiceSound()" title="Toggle sound effects">
        <i class="fa-solid fa-${soundEnabled ? 'volume-high' : 'volume-xmark'}"></i>
      </button>
      <div class="ornament my-3">✧</div>
      <h5>Recent Rolls</h5>
      <div id="diceHistory"></div>
    </div>`;
  const input = document.getElementById('diceExpr') as HTMLInputElement;
  input.addEventListener('keydown', (e) => { if (e.key === 'Enter') doRoll(); });
  loadDiceHistory();
}

// ─── Rolling Animation ───

/**
 * Show tumbling dice with rapidly cycling numbers.
 * Each die gets a CSS tumble animation + a JS recursive timeout that
 * cycles the displayed number with deceleration to simulate a real die
 * settling down.
 */
function animateDiceRoll(breakdown: any[]) {
  // Clear any existing timeouts
  rollingTimeouts.forEach(id => clearTimeout(id));
  rollingTimeouts = [];

  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  container.innerHTML = breakdown.map((b: any) => {
    if (!b.rolls || b.rolls.length === 0) return '';
    const sides = parseSides(b.die);
    if (sides === 0) return '';
    return b.rolls.map(() => buildDie(1, sides, b.die, 'rolling')).join('');
  }).join('');

  // Start number cycling with deceleration for each rolling die
  const dieElements = container.querySelectorAll('.die.rolling');
  dieElements.forEach((dieEl) => {
    const dieSides = parseInt(dieEl.getAttribute('data-sides') || '6');
    const valueEl = dieEl.querySelector('.die-value');
    if (valueEl) {
      const cycle = (delay: number) => {
        const randVal = Math.floor(Math.random() * dieSides) + 1;
        valueEl.textContent = dieSides >= 100 && randVal === 100 ? '00' : String(randVal);
        const nextDelay = Math.min(delay + 6, 180); // decelerate up to 180ms
        const id = window.setTimeout(() => cycle(nextDelay), nextDelay);
        rollingTimeouts.push(id);
      };
      cycle(50); // start fast at 50ms
    }
  });

  playDiceSound(false);
}

// ─── Settle Dice (Final Result) ───

/**
 * Stop the rolling animation and show final values with a pop-in effect.
 * Called after the server responds with the real roll results.
 * Detects crits (max roll / nat 1) on any die type d4+.
 */
function settleDice(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  setTimeout(() => {
    // Clear cycling timeouts
    rollingTimeouts.forEach(id => clearTimeout(id));
    rollingTimeouts = [];

    let hasCrit = false;
    let critType: 'success' | 'fail' | null = null;
    let critDieLabel = '';
    let critDieEl: HTMLElement | null = null;
    container.innerHTML = breakdown.map((b: any) => {
      if (!b.rolls || b.rolls.length === 0) return '';
      const sides = parseSides(b.die);
      if (sides === 0) return '';

      return b.rolls.map((r: any) => {
        const v = rollValue(r);
        const used = rollUsed(r);
        let extraClass = 'settled';
        if (!used) extraClass += ' die-dropped';
        const crit = isCritRoll(v, sides, used);
        if (crit === 'success') {
          extraClass += ' dice-crit-success';
          hasCrit = true;
          if (!critType) { critType = 'success'; critDieLabel = b.die; }
        } else if (crit === 'fail') {
          extraClass += ' dice-crit-fail';
          hasCrit = true;
          if (!critType) { critType = 'fail'; critDieLabel = b.die; }
        }
        const dieHtml = buildDie(v, sides, b.die, extraClass);
        if (crit && !critDieEl) {
          // Extract the die element from the HTML to get a reference
          // We'll find it after innerHTML is set
        }
        return dieHtml;
      }).join('');
    }).join('');

    if (hasCrit) {
      playDiceSound(true);
      // Find the crit die element for particle origin
      critDieEl = container.querySelector('.dice-crit-success, .dice-crit-fail') as HTMLElement;
      if (critDieEl && critType) {
        critCelebration(critType, critDieLabel, critDieEl);
      }
    }
  }, 900);
}

// ─── Main Roll Handler ───

export async function doRoll() {
  const expr = (document.getElementById('diceExpr') as HTMLInputElement).value;
  if (!expr) return;

  const resultEl = document.getElementById('diceResult')!;
  resultEl.style.display = 'none';

  // Start tumbling animation immediately with a fake breakdown
  const container = document.getElementById('dice3dContainer');
  if (container) {
    const m = expr.match(/(\d+)d(\d+)/gi);
    if (m) {
      const parts = m.map(s => {
        const [, count, sides] = s.match(/(\d+)d(\d+)/i)!;
        return { die: 'd' + sides, count: parseInt(count || '1'), sides: parseInt(sides) };
      });
      const breakdown = parts.flatMap(p =>
        Array.from({ length: p.count }, () => ({ die: p.die, rolls: [1], total: 0 }))
      );
      animateDiceRoll(breakdown);
    } else {
      const m2 = expr.match(/d(\d+)/i);
      if (m2) {
        const sides = parseInt(m2[1]);
        animateDiceRoll([{ die: 'd' + sides, rolls: [1], total: 0 }]);
      }
    }
  }

  try {
    const result = await api('POST', '/api/roll', { expression: expr, character_id: currentChar?.id });
    resultEl.style.display = 'block';

    if (result.breakdown) {
      settleDice(result.breakdown);
    }

    let facesHtml = '';
    if (result.breakdown) {
      facesHtml = result.breakdown.map((b: any) => {
        if (!b.rolls || b.rolls.length === 0) {
          // Modifier entry (no rolls, just a total)
          return `<span class="die-term">${esc(b.die)} = ${b.total}</span>`;
        }
        const dieLabel = b.die;
        const sides = parseSides(dieLabel);
        if (sides === 0) {
          return `<span class="die-term">${esc(b.die)} = ${b.total}</span>`;
        }
        const rolls = b.rolls.map((r: any) => {
          const v = rollValue(r);
          const used = rollUsed(r);
          const flags = rollFlags(r);
          const itemClass = used ? 'die-face die-kept' : 'die-face die-dropped';
          return `<span class="${itemClass}">${v}${flags === 'dropped' ? '✕' : ''}</span>`;
        }).join(' + ');
        return `<div class="die-group"><span class="die-label">${dieLabel}:</span> <span class="die-faces">${rolls}</span> <span class="die-subtotal">= ${b.total}</span></div>`;
      }).filter((h: string) => h).join('');
    }

    let critBadge = '';
    if (result.breakdown) {
      for (const b of result.breakdown) {
        if (!b.rolls) continue;
        const sides = parseSides(b.die);
        for (const r of b.rolls) {
          const v = rollValue(r);
          const used = rollUsed(r);
          const crit = isCritRoll(v, sides, used);
          if (crit === 'success') {
            if (sides === 20) {
              critBadge = '<span class="badge bg-success ms-2"><i class="fa-solid fa-bolt me-1"></i>Critical Hit!</span>';
            } else if (!critBadge) {
              critBadge = `<span class="badge bg-warning ms-2"><i class="fa-solid fa-star me-1"></i>Max on ${b.die}!</span>`;
            }
          } else if (crit === 'fail') {
            if (sides === 20) {
              critBadge = '<span class="badge bg-danger ms-2"><i class="fa-solid fa-skull me-1"></i>Critical Fail!</span>';
            } else if (!critBadge) {
              critBadge = `<span class="badge bg-secondary ms-2"><i class="fa-solid fa-circle-down me-1"></i>Nat 1 on ${b.die}</span>`;
            }
          }
        }
      }
    }

    setTimeout(() => {
      resultEl.innerHTML = `
        <div class="dice-result-box">
          <div class="roll-total-anim">${result.total} ${critBadge}</div>
          <div class="roll-expression">${esc(result.expression)}</div>
          <div class="roll-breakdown">${facesHtml}</div>
          <div class="roll-text text-muted small">${esc(result.text)}</div>
        </div>`;
    }, 500);
    loadDiceHistory();
  } catch (e: any) {
    toast(e.message, true);
  }
}

export async function loadDiceHistory() {
  const el = document.getElementById('diceHistory');
  if (!el) return;
  try {
    const rolls = await api('GET', '/api/dice-rolls' + (currentChar ? `?character_id=${currentChar.id}` : ''));
    el.innerHTML = rolls.slice(0, 20).map((r: any) =>
      `<div class="d-flex justify-content-between py-1 border-bottom dice-history-item">
        <span class="small">${esc(r.expression)}</span>
        <span><strong>${r.total}</strong> <span class="text-muted small">${esc(r.result)}</span></span>
      </div>`
    ).join('') || '<div class="text-center text-muted py-3">No rolls yet</div>';
  } catch { }
}
