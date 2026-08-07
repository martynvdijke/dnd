// @ts-nocheck
/**
 * Character combat rendering — HP/AC, damage/heal, rest, death saves, conditions.
 * Extracted from app.ts (address-tech-debt-and-ux). Replaces the earlier
 * partial port; app.ts versions are authoritative.
 */
import { expose } from '../lib/expose';
import { currentChar, setCurrentChar } from '../lib/state';
import { esc, toast } from '../lib/dom';
import { api } from '../lib/api';
import { animateHpChange } from '../lib/animations';

// ─── Roll / Combat Actions ───

export async function rollCheck(type: string, name: string, adv: string) {
  if (!currentChar) return;
  try {
    const result = await api('POST', '/api/roll/check', {
      character_id: currentChar.id, type, name, advantage: adv,
    });
    toast(result.text);
  } catch (e: any) {
    toast(e.message, true);
  }
}

export async function applyHeal() {
  if (!currentChar) return;
  const heal = parseInt((document.getElementById('healInput') as HTMLInputElement)?.value || '0');
  if (!heal) return;
  const oldHp = currentChar.hp_current;
  const newHp = Math.min(currentChar.hp_max, currentChar.hp_current + heal);
  await (window as any).updateField('hp_current', newHp);
  await (window as any).saveCharacter?.();
  (window as any).renderSheet?.();
  // Animate HP change after re-render
  const bar = document.getElementById('charHpBarFill');
  const hpText = document.getElementById('charHpText');
  if (bar && hpText) {
    bar.style.width = Math.max(0, Math.min(100, (oldHp / currentChar.hp_max) * 100)) + '%';
    animateHpChange(hpText, bar, oldHp, currentChar.hp_current, currentChar.hp_max);
  }
}

export async function doRest(type: string) {
  if (!currentChar) return;
  try {
    const oldHp = currentChar.hp_current;
    const result = await api('POST', `/api/characters/${currentChar.id}/rest`, { rest_type: type, hit_dice_count: type === 'short' ? 1 : 0 });
    toast(`${type} rest: healed ${result.hp_healed} HP`);
    setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
    (window as any).renderSheet?.();
    // Animate HP change after re-render
    const bar = document.getElementById('charHpBarFill');
    const hpText = document.getElementById('charHpText');
    if (bar && hpText && result.hp_healed > 0) {
      bar.style.width = Math.max(0, Math.min(100, (oldHp / currentChar.hp_max) * 100)) + '%';
      animateHpChange(hpText, bar, oldHp, currentChar.hp_current, currentChar.hp_max);
    }
  } catch (e: any) {
    toast(e.message, true);
  }
}

export async function doLevelUp() {
  if (!currentChar) return;
  try {
    const result = await api('POST', `/api/characters/${currentChar.id}/levelup`);
    toast(`Level Up! Now level ${result.new_level} (+${result.hp_gain} HP)`);
    setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
    (window as any).renderSheet?.();
  } catch (e: any) {
    toast(e.message, true);
  }
}

// ─── Combat Section Update for conditions and concentration ───

export function renderCombat() {
  const c = currentChar;
  const el = document.getElementById('combatSection')!;
  const pct = c.hp_max > 0 ? Math.round((c.hp_current / c.hp_max) * 100) : 0;
  el.innerHTML = `
    <div class="combat-stack row g-3">
      <div class="combat-stat-row col-4"><div class="combat-stat" title="Armor Class"><div class="stat-label">AC</div><div class="stat-value">${(window as any).renderStepper('ac', c.ac, 1, 0, undefined, 'AC')}</div></div></div>
      <div class="combat-stat-row col-4"><div class="combat-stat" title="Initiative modifier"><div class="stat-label">Initiative</div><div class="stat-value">${(window as any).renderStepper('initiative', c.initiative, 1, undefined, undefined, 'Initiative')}</div></div></div>
      <div class="combat-stat-row col-4"><div class="combat-stat" title="Movement speed"><div class="stat-label">Speed</div><div class="stat-value">${(window as any).renderStepper('speed', c.speed, 5, 0, undefined, 'Speed')}</div></div></div>
    </div>
    <h5 class="mt-3">Hit Points</h5>
    <div class="hp-bar position-relative mb-2" title="${c.hp_current} / ${c.hp_max} HP${c.temp_hp > 0 ? ' (+' + c.temp_hp + ' temporary)' : ''}">
      <div class="hp-bar-fill" id="charHpBarFill" style="width:${pct}%"></div>
      <div class="position-absolute top-0 start-0 end-0 bottom-0 d-flex align-items-center justify-content-center text-white small fw-bold" id="charHpText" style="font-size:0.8rem">${c.hp_current} / ${c.hp_max}${c.temp_hp > 0 ? ' (+' + c.temp_hp + ' temp)' : ''}</div>
    </div>
    <div class="row g-2">
      <div class="col-4"><label class="form-label small">HP Max</label>${(window as any).renderStepper('hp_max', c.hp_max, 1, 1, undefined, 'HP Max', 'lg')}</div>
      <div class="col-4"><label class="form-label small">Current</label>${(window as any).renderStepper('hp_current', c.hp_current, 1, 0, undefined, 'HP Current', 'lg')}</div>
      <div class="col-4"><label class="form-label small">Temp HP</label>${(window as any).renderStepper('temp_hp', c.temp_hp, 1, 0, undefined, 'Temp HP', 'lg')}</div>
    </div>
    <div class="row g-2 mt-2">
      <div class="col-6">
        <label class="form-label small">Damage</label>
        <div class="input-group input-group-sm"><input type="number" class="form-control" id="dmgInput" value="0"><button class="btn btn-danger" onclick="applyDamage()">Apply</button></div>
      </div>
      <div class="col-6">
        <label class="form-label small">Heal</label>
        <div class="input-group input-group-sm"><input type="number" class="form-control" id="healInput" value="0"><button class="btn btn-success" onclick="applyHeal()">Apply</button></div>
      </div>
    </div>
    <div class="d-flex gap-2 mt-3 flex-wrap">
      <button class="btn btn-sm btn-outline-primary" onclick="doRest('short')"><i class="fa-solid fa-campground me-1"></i>Short Rest</button>
      <button class="btn btn-sm btn-outline-primary" onclick="doRest('long')"><i class="fa-solid fa-moon me-1"></i>Long Rest</button>
      <button class="btn btn-sm btn-gold" onclick="doLevelUp()"><i class="fa-solid fa-arrow-up me-1"></i>Level Up</button>
    </div>
    <div id="conditionsArea" class="mt-3">
      <div class="d-flex justify-content-between align-items-center">
        <h5 class="mt-0 mb-2">Conditions</h5>
        <div class="d-flex gap-1">
          <button class="btn btn-sm btn-outline-primary" onclick="showAddCondition()"><i class="fa-solid fa-plus"></i></button>
          <button class="btn btn-sm btn-outline-secondary" onclick="tickConditions()" title="Advance 1 round"><i class="fa-solid fa-forward"></i></button>
        </div>
      </div>
      <div id="conditionBadges"></div>
    </div>
    <h5 class="mt-3">Saving Throws <small class="text-muted fw-normal">(click to roll)</small></h5>
    <div class="d-flex flex-wrap gap-1 mb-3">
      ${['str','dex','con','int','wis','cha'].map(a => {
        const mod = (c as any)[`${a}_mod`];
        const total = c.proficiency_bonus + mod;
        const sign = total >= 0 ? '+' : '';
        return `<span class="badge badge-gold" style="cursor:pointer;font-size:0.85rem;padding:0.4rem 0.6rem" onclick="rollCheck('save','${a}','normal')">${a.toUpperCase()} ${sign}${total}</span>`;
      }).join('')}
    </div>
    <h5 class="mt-3">Death Saves</h5>
    <div class="row g-2">
      <div class="col-6"><label class="form-label small">Successes</label>${(window as any).renderStepper('death_saves_successes', c.death_saves_successes, 1, 0, 3, 'Death Save Successes')}</div>
      <div class="col-6"><label class="form-label small">Failures</label>${(window as any).renderStepper('death_saves_failures', c.death_saves_failures, 1, 0, 3, 'Death Save Failures')}</div>
    </div>
    <h5 class="mt-3">Concentration</h5>
    <div class="form-check"><input type="checkbox" class="form-check-input" id="concentrationCb" ${c.concentrating ? 'checked' : ''} onchange="autoSaveField('concentrating',this)"><label class="form-check-label" for="concentrationCb">Concentrating on a spell</label></div>
    <div class="mt-2">
      <label class="form-label small">Concentrating On</label>
      <input class="form-control form-control-sm" value="${esc(c.concentrating_on)}" oninput="autoSaveField('concentrating_on',this)" placeholder="e.g. Hunter's Mark" style="min-height:44px;font-size:1rem">
    </div>
    <h5 class="mt-3">Hit Dice</h5>
    <div class="row g-2">
      <div class="col-6"><label class="form-label small">Total</label>${(window as any).renderStepper('hit_dice_total', c.hit_dice_total, 1, 0, undefined, 'Hit Dice Total')}</div>
      <div class="col-6"><label class="form-label small">Used</label>${(window as any).renderStepper('hit_dice_used', c.hit_dice_used, 1, 0, undefined, 'Hit Dice Used')}</div>
    </div>`;
  // Load condition badges async
  loadConditionBadges();
}

export async function loadConditionBadges() {
  if (!currentChar) return;
  try {
    const conds = await api('GET', `/api/conditions/summary?character_id=${currentChar.id}`);
    const el = document.getElementById('conditionBadges');
    if (!el) return;
    if (!conds.length) {
      el.innerHTML = '<div class="text-muted small fst-italic">No active conditions</div>';
      return;
    }
    const iconMap: Record<string, string> = {
      blinded: 'fa-eye-slash', charmed: 'fa-heart', deafened: 'fa-ear-deaf',
      exhaustion: 'fa-battery-quarter', frightened: 'fa-ghost', grappled: 'fa-handcuffs',
      incapacitated: 'fa-bed', invisible: 'fa-ghost', paralyzed: 'fa-snowflake',
      petrified: 'fa-monument', poisoned: 'fa-skull', prone: 'fa-person-falling',
      restrained: 'fa-lock', stunned: 'fa-star', unconscious: 'fa-circle',
      concentration: 'fa-brain',
    };
    const colorMap: Record<string, string> = {
      blinded: '#8b0000', charmed: '#dda0dd', deafened: '#666',
      exhaustion: '#ff8c00', frightened: '#4b0082', grappled: '#8b4513',
      incapacitated: '#555', invisible: '#87ceeb', paralyzed: '#00bfff',
      petrified: '#808080', poisoned: '#32cd32', prone: '#d2b48c',
      restrained: '#ffd700', stunned: '#ff4500', unconscious: '#2f4f4f',
      concentration: '#4169e1',
    };
    el.innerHTML = '<div class="d-flex flex-wrap gap-1 mb-2">' + conds.map((cond: any) => {
      const icon = iconMap[cond.type] || 'fa-circle';
      const color = colorMap[cond.type] || '#b8963e';
      const durStr = cond.duration_type === 'permanent' ? 'perm' : cond.duration + cond.duration_type.substring(0, 1);
      return `<span class="badge" style="background:${color};color:#fff;font-size:0.75rem;padding:0.3rem 0.5rem;border-radius:4px" title="${esc(cond.name)} (${durStr})">
        <i class="fa-solid ${icon} me-1"></i>${esc(cond.name)} ${durStr}
        <a href="#" onclick="deleteCondition(${cond.id});return false" class="text-white text-decoration-none ms-1">×</a>
      </span>`;
    }).join('') + '</div>';
  } catch {}
}

// Window registrations (centralized)
expose('rollCheck', rollCheck);
expose('applyHeal', applyHeal);
expose('doRest', doRest);
expose('doLevelUp', doLevelUp);
expose('renderCombat', renderCombat);
