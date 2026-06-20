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

// Active rolling intervals (cleared on settle)
let rollingIntervals: number[] = [];

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
          if (v === 20) badge = '<span class="badge bg-success ms-2">Critical Hit!</span>';
          else if (v === 1) badge = '<span class="badge bg-danger ms-2">Critical Fail!</span>';
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
 * Each die gets a CSS tumble animation + a JS interval that cycles
 * the displayed number to simulate the die rolling.
 */
function animateDiceRoll(breakdown: any[]) {
  // Clear any existing intervals
  rollingIntervals.forEach(id => clearInterval(id));
  rollingIntervals = [];

  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  container.innerHTML = breakdown.map((b: any) => {
    if (!b.rolls || b.rolls.length === 0) return '';
    const sides = parseSides(b.die);
    if (sides === 0) return '';
    return b.rolls.map(() => buildDie(1, sides, b.die, 'rolling')).join('');
  }).join('');

  // Start number cycling for each rolling die
  const dieElements = container.querySelectorAll('.die.rolling');
  dieElements.forEach((dieEl) => {
    const dieSides = parseInt(dieEl.getAttribute('data-sides') || '6');
    const valueEl = dieEl.querySelector('.die-value');
    if (valueEl) {
      const intervalId = window.setInterval(() => {
        const randVal = Math.floor(Math.random() * dieSides) + 1;
        valueEl.textContent = dieSides >= 100 && randVal === 100 ? '00' : String(randVal);
      }, 80);
      rollingIntervals.push(intervalId);
    }
  });
}

// ─── Settle Dice (Final Result) ───

/**
 * Stop the rolling animation and show final values with a pop-in effect.
 * Called after the server responds with the real roll results.
 */
function settleDice(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  setTimeout(() => {
    // Clear cycling intervals
    rollingIntervals.forEach(id => clearInterval(id));
    rollingIntervals = [];

    container.innerHTML = breakdown.map((b: any) => {
      if (!b.rolls || b.rolls.length === 0) return '';
      const sides = parseSides(b.die);
      if (sides === 0) return '';

      return b.rolls.map((r: any) => {
        const v = rollValue(r);
        const used = rollUsed(r);
        let extraClass = 'settled';
        if (!used) extraClass += ' die-dropped';
        if (sides === 20) {
          if (v === 20) extraClass += ' dice-crit-success';
          else if (v === 1) extraClass += ' dice-crit-fail';
        }
        return buildDie(v, sides, b.die, extraClass);
      }).join('');
    }).join('');
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
        if (!b.rolls || b.rolls.length === 0) return '';
        const dieLabel = b.die;
        const sides = parseSides(dieLabel);
        if (sides === 0) return '';
        const rolls = b.rolls.map((r: any) => {
          const v = rollValue(r);
          const used = rollUsed(r);
          const flags = rollFlags(r);
          const itemClass = used ? 'die-face die-kept' : 'die-face die-dropped';
          const flagText = flags ? ` data-flags="${esc(flags)}"` : '';
          return `<span class="${itemClass}"${flagText}>${v}${flags === 'dropped' ? '✕' : ''}</span>`;
        }).join('');
        return `<div class="die-group"><span class="die-label">${dieLabel}:</span> <span class="die-faces">${rolls}</span></div>`;
      }).filter((h: string) => h).join('');
    }

    let critBadge = '';
    if (result.breakdown) {
      for (const b of result.breakdown) {
        if (b.die === 'd20' && b.rolls) {
          for (const r of b.rolls) {
            const v = rollValue(r);
            if (v === 20) critBadge = '<span class="badge bg-success ms-2"><i class="fa-solid fa-bolt me-1"></i>Critical Hit!</span>';
            else if (v === 1) critBadge = '<span class="badge bg-danger ms-2"><i class="fa-solid fa-skull me-1"></i>Critical Fail!</span>';
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
