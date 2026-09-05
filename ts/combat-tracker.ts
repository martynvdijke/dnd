import { esc, showModal, hideModal, toast } from './lib/dom';
import { api } from './lib/api';
import { showView } from './navigation';
import { animateHpChange, animateTurnChange } from './lib/animations';
import { expose } from './lib/expose';

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
      el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-swords fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Combatants</p><p class="small text-muted">Add combat entries from a character sheet or create them here.</p><button class="btn btn-gold btn-sm mt-2" onclick="showAddCombatEntry()"><i class="fa-solid fa-plus me-1"></i>Add Combatant</button></div>';
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
            <button class="btn btn-sm btn-outline-danger py-0 px-1" style="font-size:0.65rem" onclick="deleteCombatEntry(${entry.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </td>
      </tr>`;
    }
    html += '</tbody></table></div>';
    el.innerHTML = html;
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(msg)}</p></div>`;
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
