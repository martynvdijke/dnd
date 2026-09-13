import { esc, showModal, hideModal, toast } from './lib/dom';
import { api } from './lib/api';
import { showView } from './navigation';
import { animateHpChange, animateTurnChange } from './lib/animations';
import { expose } from './lib/expose';
import { currentCampaign } from './lib/state';

interface CombatEntry {
  id: number;
  name: string;
  type: string;
  ac: number;
  hp_current: number;
  hp_max: number;
  initiative_roll: number;
  initiative_mod: number;
  turn_order: number;
  is_active: boolean;
  [key: string]: unknown;
}

// ─── Combat Tracker ───

expose('showCombatTracker', async function (): Promise<void> {
  showView('combatTracker');
  const el = document.getElementById('combatTrackerContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading combat tracker... ✧</div>';
  try {
    const [entries, _campaigns] = await Promise.all([
      api<CombatEntry[]>('GET', '/api/combat'),
      api<unknown[]>('GET', '/api/campaigns'),
    ]);
    if (!entries.length) {
      el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-swords fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Combatants</p><p class="small text-muted">Add combat entries from a character sheet or create them here.</p><button class="btn btn-gold btn-sm mt-2" onclick="showAddCombatEntry()"><i class="fa-solid fa-plus me-1"></i>Add Combatant</button></div><div data-testid="combat-log" id="combatLogPanel"></div>';
      await refreshCombatLog();
      return;
    }
    const sorted = [...entries].sort((a, b) => b.initiative_roll - a.initiative_roll || b.turn_order - a.turn_order);

    let html = `<div class="d-flex justify-content-between align-items-center mb-3 flex-wrap gap-2">
      <div class="d-flex gap-2">
        <button class="btn btn-gold btn-sm" onclick="showAddCombatEntry()"><i class="fa-solid fa-plus me-1"></i>Add</button>
        <button class="btn btn-outline-primary btn-sm" onclick="rollAllInitiative()"><i class="fa-solid fa-dice me-1"></i>Roll Init</button>
        <button class="btn btn-outline-secondary btn-sm" onclick="advanceCombatTurn()"><i class="fa-solid fa-forward me-1"></i>Next Turn</button>
      </div>
    </div>
    <div class="table-responsive">
      <table class="table table-hover align-middle mb-0" id="combatTrackerTable">
        <thead><tr>
          <th style="width:30px"></th>
          <th>Name</th>
          <th style="width:60px">Init</th>
          <th style="width:80px">AC</th>
          <th style="width:120px">HP</th>
          <th style="width:60px">Status</th>
          <th style="width:140px">Actions</th>
        </tr></thead>
        <tbody id="combatTrackerBody">`;

    for (const entry of sorted) {
      const active = entry.is_active;
      const hpPct = entry.hp_max > 0 ? Math.round((entry.hp_current / entry.hp_max) * 100) : 0;
      const hpColor = hpPct > 50 ? 'var(--bs-success)' : hpPct > 25 ? 'var(--gold)' : 'var(--bs-danger)';
      const rowClass = active ? 'table-active fw-bold' : '';
      const icon = entry.type === 'character' ? 'fa-user' : entry.type === 'monster' ? 'fa-dragon' : 'fa-user-group';
      html += `<tr class="${rowClass}" draggable="true" id="ce-${entry.id}"
        ondragstart="dragCombatEntry(event, ${entry.id})"
        ondrop="dropCombatEntry(event, ${entry.id})"
        ondragover="event.preventDefault()">
        <td class="text-muted" style="cursor:grab"><i class="fa-solid fa-grip-vertical"></i></td>
        <td><i class="fa-solid ${icon} me-1 text-muted"></i>${esc(entry.name)}
          ${entry.type === 'character' ? '<span class="badge badge-blood ms-1" style="font-size:0.6rem">PC</span>' : ''}
        </td>
        <td class="text-center fw-bold">${entry.initiative_roll > 0 ? entry.initiative_roll : '-'}</td>
        <td class="text-center">${entry.ac}</td>
        <td>
          <div class="d-flex align-items-center gap-1">
            <div class="hp-bar flex-grow-1" style="height:6px;min-width:50px">
              <div class="hp-bar-fill" style="width:${hpPct}%;height:100%;background:${hpColor}"></div>
            </div>
            <span class="small" style="font-size:0.7rem;white-space:nowrap">${entry.hp_current}/${entry.hp_max}</span>
          </div>
          <div class="d-flex gap-1 mt-1">
            <input type="number" class="form-control form-control-sm" id="qdamage-${entry.id}" placeholder="dmg" style="width:55px;font-size:0.7rem;height:24px">
            <button class="btn btn-sm btn-danger py-0 px-1" style="font-size:0.65rem;height:24px" onclick="combatTrackerDamage(${entry.id})"><i class="fa-solid fa-minus"></i></button>
            <button class="btn btn-sm btn-success py-0 px-1" style="font-size:0.65rem;height:24px" onclick="combatTrackerHeal(${entry.id})"><i class="fa-solid fa-plus"></i></button>
          </div>
        </td>
        <td class="text-center">
          <button class="btn btn-sm ${active ? 'btn-gold' : 'btn-outline-secondary'} py-0 px-1" style="font-size:0.65rem" onclick="toggleCombatActive(${entry.id})">
            ${active ? '<i class="fa-solid fa-check"></i>' : '<i class="fa-solid fa-pause"></i>'}
          </button>
        </td>
        <td>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-outline-primary py-0 px-1" data-testid="combat-attack-btn" style="font-size:0.65rem" onclick="showAttackModal(${entry.id})"><i class="fa-solid fa-crosshairs me-1"></i>Attack</button>
            <button class="btn btn-sm btn-outline-danger py-0 px-1" style="font-size:0.65rem" onclick="deleteCombatEntry(${entry.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </td>
      </tr>`;
    }
    html += '</tbody></table></div><div data-testid="combat-log" id="combatLogPanel" class="mt-4"></div>';
    el.innerHTML = html;
    await refreshCombatLog();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(msg)}</p></div><div data-testid="combat-log" id="combatLogPanel"></div>`;
  }
});

async function findCombatEntry(id: number): Promise<CombatEntry | undefined> {
  const entries = await api<CombatEntry[]>('GET', '/api/combat');
  return entries.find((entry) => entry.id === id);
}

expose('combatTrackerDamage', async function (id: number): Promise<void> {
  const input = document.getElementById('qdamage-' + id) as HTMLInputElement | null;
  const dmg = parseInt(input?.value || '0', 10);
  if (!dmg) return;
  try {
    const entry = await findCombatEntry(id);
    if (!entry) { toast('Entry not found', true); return; }
    const oldHp = entry.hp_current;
    entry.hp_current = Math.max(0, entry.hp_current - dmg);
    await api('PUT', '/api/combat/' + id, entry);
    await window.showCombatTracker();
    const row = document.getElementById('ce-' + id);
    if (row) {
      const bar = row.querySelector('.hp-bar-fill') as HTMLElement | null;
      const hpText = row.querySelector('span.small') as HTMLElement | null;
      if (bar && hpText) {
        bar.style.width = Math.max(0, Math.min(100, (oldHp / entry.hp_max) * 100)) + '%';
        animateHpChange(hpText, bar, oldHp, entry.hp_current, entry.hp_max);
      }
    }
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

expose('combatTrackerHeal', async function (id: number): Promise<void> {
  const input = document.getElementById('qdamage-' + id) as HTMLInputElement | null;
  const heal = parseInt(input?.value || '0', 10);
  if (!heal) return;
  try {
    const entry = await findCombatEntry(id);
    if (!entry) { toast('Entry not found', true); return; }
    const oldHp = entry.hp_current;
    entry.hp_current = Math.min(entry.hp_max, entry.hp_current + heal);
    await api('PUT', '/api/combat/' + id, entry);
    await window.showCombatTracker();
    const row = document.getElementById('ce-' + id);
    if (row) {
      const bar = row.querySelector('.hp-bar-fill') as HTMLElement | null;
      const hpText = row.querySelector('span.small') as HTMLElement | null;
      if (bar && hpText) {
        bar.style.width = Math.max(0, Math.min(100, (oldHp / entry.hp_max) * 100)) + '%';
        animateHpChange(hpText, bar, oldHp, entry.hp_current, entry.hp_max);
      }
    }
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

expose('toggleCombatActive', async function (id: number): Promise<void> {
  try {
    const entry = await findCombatEntry(id);
    if (!entry) { toast('Entry not found', true); return; }
    entry.is_active = !entry.is_active;
    await api('PUT', '/api/combat/' + id, entry);
    window.showCombatTracker();
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

expose('deleteCombatEntry', async function (id: number): Promise<void> {
  if (!confirm('Remove this combatant?')) return;
  try {
    await api('DELETE', '/api/combat/' + id);
    window.showCombatTracker();
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

expose('rollAllInitiative', async function (): Promise<void> {
  try {
    const entries = await api<CombatEntry[]>('GET', '/api/combat');
    for (const e of entries) {
      const result = await api<{ total: number }>('POST', '/api/roll', { expression: '1d20' });
      const roll = (result.total || 0) + (e.initiative_mod || 0);
      e.initiative_roll = roll;
      try { await api('PUT', '/api/combat/' + e.id, e); } catch { /* ignore per-entry */ }
    }
    window.showCombatTracker();
    toast('Initiative rolled for all combatants');
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

expose('advanceCombatTurn', async function (): Promise<void> {
  try {
    const prevActiveRow = document.querySelector('tr.table-active') as HTMLElement | null;
    const result = await api<{ current_entry?: { name: string } }>('POST', '/api/combat/next-turn');
    await window.showCombatTracker();
    const nextActiveRow = document.querySelector('tr.table-active') as HTMLElement | null;
    if (nextActiveRow) {
      const isMonster = nextActiveRow.querySelector('.fa-dragon') !== null;
      animateTurnChange(prevActiveRow, nextActiveRow, isMonster);
    }
    toast(result.current_entry ? `Turn: ${result.current_entry.name}` : 'Turn advanced');
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

expose('showAddCombatEntry', function (): void {
  showModal('Add Combatant', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="ceName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Type</label>
        <select class="form-select" id="ceType"><option value="character">Character</option><option value="monster">Monster</option><option value="npc">NPC</option></select></div>
      <div class="col-6"><label class="form-label">AC</label><input class="form-control" id="ceAC" type="number" value="10"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">HP Max</label><input class="form-control" id="ceHPMax" type="number" value="10"></div>
      <div class="col-6"><label class="form-label">Init Mod</label><input class="form-control" id="ceInitMod" type="number" value="0"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveNewCombatEntry()"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
});

expose('saveNewCombatEntry', async function (): Promise<void> {
  await api('POST', '/api/combat', {
    name: (document.getElementById('ceName') as HTMLInputElement).value,
    type: (document.getElementById('ceType') as HTMLSelectElement).value,
    ac: +(document.getElementById('ceAC') as HTMLInputElement).value || 10,
    hp_max: +(document.getElementById('ceHPMax') as HTMLInputElement).value || 10,
    hp_current: +(document.getElementById('ceHPMax') as HTMLInputElement).value || 10,
    initiative_mod: +(document.getElementById('ceInitMod') as HTMLInputElement).value || 0,
  });
  hideModal();
  window.showCombatTracker();
  toast('Combatant added');
});

let draggedCombatId: number | null = null;

expose('dragCombatEntry', function (ev: DragEvent, id: number): void {
  draggedCombatId = id;
  if (ev.dataTransfer) ev.dataTransfer.effectAllowed = 'move';
});

expose('dropCombatEntry', async function (ev: DragEvent, targetId: number): Promise<void> {
  ev.preventDefault();
  if (draggedCombatId === null || draggedCombatId === targetId) return;
  try {
    const entries = await api<CombatEntry[]>('GET', '/api/combat');
    const dragged = entries.find((e) => e.id === draggedCombatId);
    const target = entries.find((e) => e.id === targetId);
    if (!dragged || !target) return;
    const tempOrder = dragged.turn_order;
    dragged.turn_order = target.turn_order;
    target.turn_order = tempOrder;
    await api('PUT', '/api/combat/' + dragged.id, dragged);
    await api('PUT', '/api/combat/' + target.id, target);
    draggedCombatId = null;
    window.showCombatTracker();
    toast('Reordered');
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

// ─── Attack modal ───

let pendingAttackTargetId: number | null = null;

expose('showAttackModal', async function (entryId: number): Promise<void> {
  pendingAttackTargetId = entryId;
  let chars: any[] = [];
  try {
    const res: any = await api('GET', '/api/characters');
    chars = Array.isArray(res) ? res : (res.items || res.characters || []);
  } catch { chars = []; }
  const cid = (currentCampaign as any)?.id;
  if (cid) {
    const filtered = chars.filter((c: any) => !c.campaign_id || c.campaign_id === cid);
    if (filtered.length) chars = filtered;
  }
  showModal('Attack', `
    <div class="mb-3"><label class="form-label">Attacker</label><select class="form-select" id="atkAttacker" data-testid="combat-attack-attacker">${chars.map((c: any) => `<option value="${c.id}">${esc(c.name)}</option>`).join('') || '<option value="">No characters</option>'}</select></div>
    <div class="mb-3"><label class="form-label">Weapon</label><select class="form-select" id="atkWeapon" data-testid="combat-attack-weapon"><option value="">Loading...</option></select></div>
    <div class="row g-2 mb-3">
      <div class="col-4"><label class="form-label small">Attack Bonus</label><input class="form-control" id="atkBonus" type="number" value="5" data-testid="combat-attack-bonus"></div>
      <div class="col-4"><label class="form-label small">Damage Dice</label><input class="form-control" id="atkDice" value="1d8" data-testid="combat-attack-dice"></div>
      <div class="col-4"><label class="form-label small">Damage Type</label><input class="form-control" id="atkDmgType" value="" placeholder="slashing" data-testid="combat-attack-dmgtype"></div>
    </div>
    <div class="row g-2 mb-3">
      <div class="col-6"><label class="form-label small">Advantage</label><select class="form-select" id="atkAdv" data-testid="combat-attack-advantage"><option value="">Normal</option><option value="advantage">Advantage</option><option value="disadvantage">Disadvantage</option></select></div>
      <div class="col-6"><label class="form-label small">Condition (optional)</label><input class="form-control" id="atkCondition" placeholder="prone" data-testid="combat-attack-condition"></div>
    </div>
    <div class="d-flex gap-2 mb-3">
      <button class="btn btn-outline-primary flex-fill" data-testid="combat-attack-preview" onclick="previewAttack()">Preview</button>
      <button class="btn btn-primary flex-fill" data-testid="combat-attack-apply" onclick="applyAttack()">Apply</button>
    </div>
    <div id="atkResult" data-testid="combat-attack-result" class="small border rounded p-2" style="min-height:40px"></div>
  `);
  // populate weapons for first attacker
  const sel = document.getElementById('atkAttacker') as HTMLSelectElement | null;
  const loadWeapons = async () => {
    const aid = (document.getElementById('atkAttacker') as HTMLSelectElement)?.value;
    const wsel = document.getElementById('atkWeapon') as HTMLSelectElement | null;
    if (!wsel || !aid) return;
    try {
      const char = await api<any>('GET', `/api/characters/${aid}`);
      const inv: any[] = char.inventory || [];
      const weapons = inv.filter((i: any) => i.damage_dice);
      wsel.innerHTML = '<option value="">No weapon / manual</option>' + weapons.map((w: any) => `<option value="${w.id}" data-dice="${esc(w.damage_dice)}" data-dtype="${esc(w.damage_type || '')}">${esc(w.name)} (${esc(w.damage_dice)}${w.damage_type ? ' ' + esc(w.damage_type) : ''})</option>`).join('');
      // auto-fill first weapon
      if (weapons.length) {
        // leave manual fields as-is; user can select
      }
    } catch {
      wsel.innerHTML = '<option value="">No weapon / manual</option>';
    }
  };
  sel?.addEventListener('change', loadWeapons);
  const wsel = document.getElementById('atkWeapon') as HTMLSelectElement | null;
  wsel?.addEventListener('change', () => {
    const opt = wsel.selectedOptions[0] as HTMLOptionElement | undefined;
    if (opt && opt.value) {
      (document.getElementById('atkDice') as HTMLInputElement).value = opt.dataset.dice || '';
      (document.getElementById('atkDmgType') as HTMLInputElement).value = opt.dataset.dtype || '';
    }
  });
  await loadWeapons();
});

function buildAttackBody(apply: boolean): Record<string, unknown> {
  const attackerId = +(document.getElementById('atkAttacker') as HTMLSelectElement)?.value || 0;
  const weaponId = (document.getElementById('atkWeapon') as HTMLSelectElement)?.value || '';
  const adv = (document.getElementById('atkAdv') as HTMLSelectElement)?.value || '';
  const bonusVal = (document.getElementById('atkBonus') as HTMLInputElement)?.value;
  const diceVal = (document.getElementById('atkDice') as HTMLInputElement)?.value;
  const dmgType = (document.getElementById('atkDmgType') as HTMLInputElement)?.value;
  const cond = (document.getElementById('atkCondition') as HTMLInputElement)?.value || undefined;
  const body: Record<string, unknown> = {
    attacker_type: 'character',
    attacker_id: attackerId,
    target_type: 'combat',
    target_id: pendingAttackTargetId,
    apply,
  };
  if (weaponId) body.item_id = +weaponId;
  else {
    if (bonusVal !== '' && bonusVal !== undefined) body.attack_bonus = +bonusVal;
    if (diceVal) body.damage_dice = diceVal;
    if (dmgType) body.damage_type = dmgType;
  }
  if (diceVal && weaponId) {
    // also send manual dice override if typed
    body.damage_dice = diceVal;
    if (dmgType) body.damage_type = dmgType;
  }
  if (adv === 'advantage') body.advantage = 'advantage';
  else if (adv === 'disadvantage') body.advantage = 'disadvantage';
  if (cond) body.condition = cond;
  const cid = (currentCampaign as any)?.id;
  if (cid) body.campaign_id = cid;
  return body;
}

function renderAttackResult(r: any): void {
  const el = document.getElementById('atkResult');
  if (!el) return;
  const hitStr = r.critical ? 'Critical Hit!' : r.fumble ? 'Fumble!' : r.hit ? 'Hit' : 'Miss';
  el.innerHTML = `<div>Roll: ${r.attack_roll ?? ''} + ${r.attack_bonus ?? ''} = ${r.attack_total ?? ''} vs AC ${r.target_ac ?? ''} — <strong>${hitStr}</strong></div>${r.damage !== undefined ? `<div>Damage: ${r.damage} ${esc(r.damage_type || '')} ${r.damage_breakdown ? '(' + esc(r.damage_breakdown) + ')' : ''}</div>` : ''}${r.condition_applied ? `<div>Condition: ${esc(r.condition_applied)}</div>` : ''}`;
}

expose('previewAttack', async function (): Promise<void> {
  try {
    const body = buildAttackBody(false);
    const res = await api<any>('POST', '/api/combat/attack', body);
    renderAttackResult(res);
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

expose('applyAttack', async function (): Promise<void> {
  try {
    const body = buildAttackBody(true);
    const res = await api<any>('POST', '/api/combat/attack', body);
    renderAttackResult(res);
    hideModal();
    await window.showCombatTracker();
    await refreshCombatLog();
    if (res.target_hp !== undefined) toast(`Applied: ${res.damage ?? 0} damage`);
  } catch (e: unknown) { toast(e instanceof Error ? e.message : String(e), true); }
});

// ─── Combat log ───

export async function refreshCombatLog(): Promise<void> {
  const panel = document.getElementById('combatLogPanel');
  if (!panel) return;
  try {
    const cid = (currentCampaign as any)?.id;
    const q = cid ? `?campaign_id=${cid}&limit=50` : '?limit=50';
    const entries = await api<any[]>('GET', `/api/combat-log${q}`);
    if (!entries.length) {
      panel.innerHTML = '<div class="text-muted small fst-italic mt-3">No combat log entries.</div>';
      return;
    }
    panel.innerHTML = '<h6 class="mt-3">Combat Log</h6>' + entries.map((e: any) => `
      <div class="small border rounded p-2 mb-1">
        <span class="fw-bold">${esc(e.actor_name)}</span> ${esc(e.action)} <span class="fw-bold">${esc(e.target_name || '')}</span>
        ${e.damage ? `<span class="badge bg-danger ms-1">${e.damage} ${esc(e.damage_type || '')}</span>` : ''}
        ${e.healing ? `<span class="badge bg-success ms-1">+${e.healing} HP</span>` : ''}
        ${e.roll_expression ? `<span class="text-muted ms-1">${esc(e.roll_expression)}=${e.roll_total}</span>` : ''}
        ${e.is_critical ? '<span class="badge bg-warning ms-1">crit</span>' : ''}
        ${e.condition_applied ? `<span class="badge bg-info ms-1">${esc(e.condition_applied)}</span>` : ''}
        <div class="text-muted">${esc(e.description || '')}</div>
      </div>
    `).join('');
  } catch {
    panel.innerHTML = '<div class="text-muted small">Could not load combat log.</div>';
  }
}

expose('refreshCombatLog', refreshCombatLog);
