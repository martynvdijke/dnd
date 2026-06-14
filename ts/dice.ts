/**
 * Dice rolling module — 3D dice rendering, rolling logic, history.
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

// Pip positions for d6 faces (1-6)
const D6_PIPS: Record<number, Array<{ top: string; left: string }>> = {
  1: [{ top: '50%', left: '50%' }],
  2: [{ top: '25%', left: '25%' }, { top: '75%', left: '75%' }],
  3: [{ top: '25%', left: '25%' }, { top: '50%', left: '50%' }, { top: '75%', left: '75%' }],
  4: [{ top: '25%', left: '25%' }, { top: '25%', left: '75%' }, { top: '75%', left: '25%' }, { top: '75%', left: '75%' }],
  5: [{ top: '25%', left: '25%' }, { top: '25%', left: '75%' }, { top: '50%', left: '50%' }, { top: '75%', left: '25%' }, { top: '75%', left: '75%' }],
  6: [{ top: '25%', left: '25%' }, { top: '25%', left: '75%' }, { top: '50%', left: '25%' }, { top: '50%', left: '75%' }, { top: '75%', left: '25%' }, { top: '75%', left: '75%' }],
};

// Cache pre-computed face transforms
const FACE_TRANSFORMS: Record<number, Array<{ rx: number; ry: number; rz: number }>> = {};
[4, 6, 8, 10, 12, 20].forEach(s => { FACE_TRANSFORMS[s] = getFaceTransforms(s); });

// ─── 3D Geometry Helpers ───

/** Generate N evenly-distributed points on a sphere (Fibonacci sphere algorithm). */
function fibonacciSphere(n: number, radius: number): Array<{ x: number; y: number; z: number }> {
  if (n <= 1) return [{ x: 0, y: 0, z: radius }];
  const points: Array<{ x: number; y: number; z: number }> = [];
  const phi = Math.PI * (3 - Math.sqrt(5));
  for (let i = 0; i < n; i++) {
    const y = 1 - (i / (n - 1)) * 2;
    const r = Math.sqrt(1 - y * y);
    const theta = phi * i;
    points.push({ x: r * Math.cos(theta) * radius, y: y * radius, z: r * Math.sin(theta) * radius });
  }
  return points;
}

/** Compute CSS rotation angles to make (nx, ny, nz) face the viewer (+Z). */
function normalToRotation(nx: number, ny: number, nz: number): { rx: number; ry: number } {
  const ry = Math.atan2(nx, nz) * (180 / Math.PI);
  const len = Math.sqrt(nx * nx + ny * ny + nz * nz);
  const rx = Math.asin(-ny / len) * (180 / Math.PI);
  return { rx, ry };
}

/** Get pre-calculated face transforms for a given die type. */
function getFaceTransforms(sides: number): Array<{ rx: number; ry: number; rz: number }> {
  const radius = 36;
  if (sides === 6) {
    return [
      { rx: 0, ry: 0, rz: 0 },
      { rx: 0, ry: 180, rz: 0 },
      { rx: 0, ry: 90, rz: 0 },
      { rx: 0, ry: -90, rz: 0 },
      { rx: -90, ry: 0, rz: 0 },
      { rx: 90, ry: 0, rz: 0 },
    ];
  }
  const points = fibonacciSphere(sides, radius);
  if (sides === 4) {
    const t = Math.sqrt(1 / 3);
    const tetraVerts = [
      { x: t * radius, y: t * radius, z: t * radius },
      { x: -t * radius, y: -t * radius, z: t * radius },
      { x: -t * radius, y: t * radius, z: -t * radius },
      { x: t * radius, y: -t * radius, z: -t * radius },
    ];
    return tetraVerts.map(p => {
      const norm = normalToRotation(p.x, p.y, p.z);
      return { rx: norm.rx, ry: norm.ry, rz: 0 };
    });
  }
  return points.map(p => {
    const norm = normalToRotation(p.x, p.y, p.z);
    return { rx: norm.rx, ry: norm.ry, rz: 0 };
  });
}

/** Compute die rotation string to show a specific face value. */
function rotateDieToShow(sides: number, value: number): string {
  const faceIdx = Math.max(0, Math.min(sides - 1, value - 1));
  const transforms = FACE_TRANSFORMS[sides as keyof typeof FACE_TRANSFORMS]
    || getFaceTransforms(sides);
  if (faceIdx < transforms.length) {
    const t = transforms[faceIdx];
    return `rotateX(${-t.rx}deg) rotateY(${-t.ry}deg)`;
  }
  return '';
}

/** Compute the rolling animation class for a die type. */
function rollingClass(sides: number): string {
  if (sides === 4) return 'rolling-d4';
  if (sides === 6) return 'rolling';
  if (sides === 8) return 'rolling-d8';
  if (sides === 10) return 'rolling-d10';
  if (sides === 12) return 'rolling-d12';
  if (sides === 20) return 'rolling-d20';
  return 'rolling';
}

/** Face shape class for the polyhedron type. */
function faceShapeClass(sides: number): string {
  if (sides === 4 || sides === 8 || sides === 20) return 'tri-face';
  if (sides === 10) return 'kite-face';
  if (sides === 12) return 'pent-face';
  return '';
}

// ─── Build 3D Die HTML ───

function build3DDie(value: number, sides: number, dieLabel: string): string {
  const maxSides = sides >= 100 ? 100 : sides;
  const dieClass = 'd' + (maxSides >= 100 ? 100 : maxSides);
  const transforms = FACE_TRANSFORMS[maxSides as keyof typeof FACE_TRANSFORMS]
    || getFaceTransforms(maxSides);
  const shapeCls = faceShapeClass(maxSides);
  const dieRot = rotateDieToShow(maxSides, value);

  const facesHtml = transforms.map((t, i) => {
    const faceValue = i + 1;
    const displayVal = maxSides === 100
      ? (faceValue === 10 ? '00' : String(faceValue * 10))
      : String(faceValue);
    if (maxSides === 6 && faceValue >= 1 && faceValue <= 6) {
      const pips = D6_PIPS[faceValue] || [];
      const pipHtml = pips.map(p =>
        `<span class="pip" style="top:${p.top};left:${p.left};transform:translate(-50%,-50%)"></span>`
      ).join('');
      return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${pipHtml}</div>`;
    }
    return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${displayVal}</div>`;
  }).join('');

  return `<div class="dice-3d-die ${dieClass}" data-sides="${maxSides}" data-value="${value}" style="transform:${dieRot}">${facesHtml}</div>`;
}

// ─── Die Value Helpers ───

function rollValue(r: any): number {
  return typeof r === 'number' ? r : (r.value ?? 0);
}

function rollUsed(r: any): boolean {
  return typeof r === 'number' ? true : (r.useInTotal !== false);
}

function rollFlags(r: any): string {
  return typeof r === 'number' ? '' : (r.modifierFlags || '');
}

function parseSides(dieLabel: string): number {
  const m = dieLabel.match(/^d(\d+)$/i);
  return m ? parseInt(m[1]) : 0;
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
  try {
    const result = await api('POST', '/api/roll', {
      expression: expr,
      character_id: currentChar?.id,
      advantage: isAdv ? 'advantage' : 'disadvantage',
    });
    const container = document.getElementById('dice3dContainer');
    const resultDiv = document.getElementById('diceResult');
    if (container) container.innerHTML = '';
    if (resultDiv) {
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
      resultDiv.style.display = 'block';
      resultDiv.innerHTML = `
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
      animateDiceRoll(result.breakdown);
    }
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
      <div id="dice3dContainer" class="dice-3d-container"></div>
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

function animateDiceRoll(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  container.innerHTML = breakdown.map((b: any) => {
    if (!b.rolls || b.rolls.length === 0) return '';
    const sides = parseSides(b.die);
    if (sides === 0) return '';
    const dieLabel = b.die;
    const rollCls = rollingClass(sides);
    const transforms = FACE_TRANSFORMS[sides as keyof typeof FACE_TRANSFORMS]
      || getFaceTransforms(sides);
    const shapeCls = faceShapeClass(sides);

    return b.rolls.map((r: any) => {
      const facesHtml = transforms.map((t, i) => {
        const displayVal = sides >= 100
          ? (i + 1 === 10 ? '00' : String((i + 1) * 10))
          : String(i + 1);
        if (sides === 6) {
          const pips = D6_PIPS[i + 1] || [];
          const pipHtml = pips.map(p =>
            `<span class="pip" style="top:${p.top};left:${p.left};transform:translate(-50%,-50%)"></span>`
          ).join('');
          return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${pipHtml}</div>`;
        }
        return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${displayVal}</div>`;
      }).join('');
      const dieHtml = `<div class="dice-3d-die ${b.die} ${rollCls}" data-sides="${sides}" data-value="0">${facesHtml}</div>`;
      return `<div class="dice-3d-wrapper"><span class="dice-3d-label">${dieLabel}</span>${dieHtml}</div>`;
    }).join('');
  }).join('');
}

// ─── Settle Dice (Final Result) ───

function settleDice(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  setTimeout(() => {
    container.innerHTML = breakdown.map((b: any) => {
      if (!b.rolls || b.rolls.length === 0) return '';
      const sides = parseSides(b.die);
      if (sides === 0) return '';
      const dieLabel = b.die;

      return b.rolls.map((r: any) => {
        const v = rollValue(r);
        const dieHtml = build3DDie(v, sides, dieLabel);

        let extraClass = '';
        if (sides === 20) {
          if (v === 20) extraClass = ' dice-crit-success';
          else if (v === 1) extraClass = ' dice-crit-fail';
        }

        const wrapper = document.createElement('div');
        wrapper.className = 'dice-3d-wrapper' + extraClass;
        wrapper.innerHTML = `<span class="dice-3d-label">${dieLabel}</span>${dieHtml}`;
        return wrapper.outerHTML;
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
