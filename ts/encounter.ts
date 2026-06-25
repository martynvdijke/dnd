// @ts-nocheck — extracted from app.ts, window-level self-registration
import { showView } from './navigation';
import { esc, showModal, hideModal, toast } from './lib/dom';
import { api } from './lib/api';

// ─── Encounter Builder ───

(window as any).showEncounterBuilder = async function () {
  showView('encounter');
  const el = document.getElementById('encounterContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading encounters... ✧</div>';
  try {
    const encounters = await api('GET', '/api/encounters');
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center mb-3">
        <div>
          <button class="btn btn-gold btn-sm me-2" onclick="showCreateEncounter()"><i class="fa-solid fa-plus me-1"></i>New Encounter</button>
          <button class="btn btn-outline-primary btn-sm" onclick="showEncounterXPCalc()"><i class="fa-solid fa-calculator me-1"></i>XP Calculator</button>
        </div>
      </div>
      <div class="row g-3" id="encounterList">
        ${encounters.length ? encounters.map((e: any) => `
          <div class="col-md-6 col-lg-4">
            <div class="character-card" onclick="showEncounterDetail(${e.id})">
              <div class="char-name">${esc(e.name)}</div>
              <div class="char-detail">${esc(e.environment)} · ${esc(e.difficulty)} · ${e.total_xp} XP</div>
              <div class="char-hp mt-1">${esc(e.description).substring(0, 100)}</div>
            </div>
          </div>`).join('')
          : '<div class="empty-state"><i class="fa-solid fa-crosshairs fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Encounters Yet</p><p class="small text-muted">Build balanced encounters with XP budgeting.</p></div>'}
      </div>`;
  } catch (e: any) { el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(e.message)}</p></div>`; }
};

(window as any).showCreateEncounter = function () {
  showModal('New Encounter', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="encName" placeholder="Goblin Ambush"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="encDesc" rows="2"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Environment</label>
        <select class="form-select" id="encEnv">
          <option value="">Any</option><option value="forest">Forest</option><option value="dungeon">Dungeon</option>
          <option value="mountain">Mountain</option><option value="swamp">Swamp</option><option value="urban">Urban</option>
          <option value="underdark">Underdark</option><option value="coastal">Coastal</option><option value="arctic">Arctic</option>
          <option value="desert">Desert</option><option value="grassland">Grassland</option>
        </select></div>
      <div class="col-6"><label class="form-label">Difficulty</label>
        <select class="form-select" id="encDiff">
          <option value="easy">Easy</option><option value="medium" selected>Medium</option>
          <option value="hard">Hard</option><option value="deadly">Deadly</option>
        </select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEncounter()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
};

(window as any).saveEncounter = async function () {
  const name = (document.getElementById('encName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', '/api/encounters', {
    name, description: (document.getElementById('encDesc') as HTMLTextAreaElement).value,
    environment: (document.getElementById('encEnv') as HTMLSelectElement).value,
    difficulty: (document.getElementById('encDiff') as HTMLSelectElement).value,
  });
  hideModal();
  (window as any).showEncounterBuilder();
  toast('Encounter created');
};

(window as any).showEncounterDetail = async function (id: number) {
  showView('singleEncounter');
  const el = document.getElementById('singleEncounterContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading... ✧</div>';
  try {
    const e = await api('GET', `/api/encounters/${id}`);
    const monsters = e.monsters || [];
    const totalCount = monsters.reduce((s: number, m: any) => s + m.count, 0);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-start flex-wrap gap-2 mb-2">
        <div>
          <h1 class="h2 mb-0">${esc(e.name)}</h1>
          <p class="text-muted fst-italic mb-0 mt-1">
            <span class="badge badge-gold">${esc(e.difficulty)}</span>
            ${e.environment ? `<span class="badge badge-muted ms-1">${esc(e.environment)}</span>` : ''}
            <span class="badge badge-blood ms-1">${e.total_xp} XP</span>
          </p>
        </div>
    <div class="d-flex gap-2 flex-wrap">
      <button class="btn btn-gold btn-sm" onclick="showEncounterMonsterPicker(${id})"><i class="fa-solid fa-plus me-1"></i>Add Monster</button>
      <button class="btn btn-outline-primary btn-sm" id="compareBtn" onclick="toggleCompareMode()"><i class="fa-solid fa-arrow-right-arrow-left me-1"></i>Compare</button>
      <button class="btn btn-outline-primary btn-sm" onclick="showImport()"><i class="fa-solid fa-file-import me-1"></i>Import</button>
    </div>
      </div>
      ${e.description ? `<p class="text-muted">${esc(e.description)}</p>` : ''}
      <div class="ornament my-2">✧</div>
      <h5>Monsters <span class="text-muted small">(${totalCount} total)</span></h5>
      <div id="monsterList">
        ${monsters.length ? monsters.map((m: any) => `
          <div class="inv-item">
            <div>
              ${m.compendium_monster_id
                ? `<a href="javascript:void(0)" onclick="htmx.ajax('GET','/compendium/card/monster/${m.compendium_monster_id}',{target:'#cardContainer',swap:'beforeend'})" class="fw-bold text-decoration-none">${esc(m.name)}</a>`
                : `<span class="fw-bold">${esc(m.name)}</span>`}
              <span class="badge badge-blood ms-1">x${m.count}</span>
              <span class="badge badge-gold ms-1">CR ${esc(m.cr)}</span>
              <span class="badge badge-muted ms-1">${m.xp} XP</span>
              <span class="text-muted small ms-2">AC ${m.ac} · HP ${m.hp}</span>
            </div>
            <div class="d-flex gap-1">
              ${m.compendium_monster_id ? `<button class="btn btn-sm btn-outline-info" onclick="htmx.ajax('GET','/compendium/card/monster/${m.compendium_monster_id}',{target:'#cardContainer',swap:'beforeend'})" title="View Stats"><i class="fa-solid fa-eye"></i></button>` : ''}
              <button class="btn btn-sm btn-outline-danger" onclick="deleteMonster(${e.id}, ${m.id})"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>`).join('')
          : '<div class="empty-state"><i class="fa-solid fa-skull fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">No monsters yet. Click "Add Monster" to search the compendium.</p></div>'}
      </div>
      <div class="ornament my-2">✧</div>
      <h5>XP Budget</h5>
      <div class="row g-3">
        <div class="col-md-6">
          <div class="combat-stat"><div class="stat-label">Total Monster XP</div><div class="stat-value">${e.total_xp}</div></div>
        </div>
        <div class="col-md-6">
          <div class="combat-stat"><div class="stat-label">Difficulty</div><div class="stat-value text-capitalize">${esc(e.difficulty)}</div></div>
        </div>
      </div>
      ${e.notes ? `<div class="mt-3"><h6>Notes</h6><p class="text-muted small">${esc(e.notes)}</p></div>` : ''}`;
  } catch (err: any) {
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(err.message)}</p>
      <button class="btn btn-outline-secondary btn-sm" onclick="(window as any).showEncounterBuilder()">Back</button></div>`;
  }
};

(window as any).editEncounter = function (id: number) { (window as any).showEncounterBuilder(); };

(window as any).deleteEncounter = async function (id: number) {
  if (!confirm('Delete this encounter?')) return;
  await api('DELETE', `/api/encounters/${id}`);
  (window as any).showEncounterBuilder();
  toast('Encounter deleted');
};

(window as any).showEncounterMonsterPicker = function (eid: number) {
  showModal('Add Monster from Compendium', `<div hx-get="/htmx/compendium-monsters/picker/${eid}" hx-trigger="load" hx-swap="innerHTML">Loading compendium...</div>`);
};

(window as any).importCompendiumMonsterToEncounter = async function (monsterId: number, encounterId: number, _campaignId?: number) {
  try {
    await api('POST', `/api/encounters/${encounterId}/import/compendium`, { compendium_monster_id: monsterId, count: 1 });
    toast('Monster added to encounter');
    hideModal();
    (window as any).showEncounterDetail(encounterId);
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteMonster = async function (eid: number, mid: number) {
  if (!confirm('Remove this monster?')) return;
  await api('DELETE', `/api/encounter-monsters/${mid}`);
  (window as any).showEncounterDetail(eid);
  toast('Monster removed');
};

(window as any).showEncounterXPCalc = function () {
  showModal('XP Calculator', `
    <p class="text-muted small">Enter party levels and monster CRs to calculate encounter difficulty.</p>
    <div class="mb-3"><label class="form-label">Party Levels (comma-separated)</label>
      <input class="form-control" id="xpPartyLevels" placeholder="1, 1, 2, 3" value="1,1,1,1"></div>
    <div class="mb-3"><label class="form-label">Monsters (format: CR,count per line)</label>
      <textarea class="form-control" id="xpMonsters" rows="3" placeholder="1/4,3&#10;1,1&#10;0,2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doXPCalc()"><i class="fa-solid fa-calculator me-1"></i>Calculate</button>
    <div id="xpCalcResult" class="mt-3"></div>
  `);
};

(window as any).doXPCalc = async function () {
  const levels = (document.getElementById('xpPartyLevels') as HTMLInputElement).value.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n));
  const lines = (document.getElementById('xpMonsters') as HTMLTextAreaElement).value.split('\n').filter(l => l.trim());
  const monsters = lines.map(l => {
    const [cr, count] = l.split(',').map(s => s.trim());
    return { cr, count: parseInt(count) || 1, name: 'Custom' };
  });
  if (!levels.length || !monsters.length) { toast('Enter party levels and monsters', true); return; }
  try {
    const result = await api('POST', '/api/encounters/calculate-xp', { party_levels: levels, monsters });
    const el = document.getElementById('xpCalcResult')!;
    el.innerHTML = `
      <div class="card mt-2"><div class="card-body">
        <div class="d-flex justify-content-between"><span class="fw-bold">Total XP</span><span>${result.total_xp}</span></div>
        <div class="d-flex justify-content-between"><span class="fw-bold">Adjusted XP</span><span>${result.adjusted_xp}</span></div>
        <div class="d-flex justify-content-between"><span class="fw-bold">Difficulty</span>
          <span class="badge ${result.difficulty === 'deadly' ? 'bg-danger' : result.difficulty === 'hard' ? 'badge-blood' : result.difficulty === 'medium' ? 'badge-gold' : 'bg-success'}">${result.difficulty}</span></div>
        <hr>
        <div class="small text-muted">Party: ${result.party_size} · Monsters: ${result.monster_count} · Mult: ${result.size_multiplier}x</div>
        <div class="small text-muted">Thresholds: Easy ${result.thresholds.easy} / Med ${result.thresholds.medium} / Hard ${result.thresholds.hard} / Deadly ${result.thresholds.deadly}</div>
      </div></div>`;
  } catch (e: any) { toast(e.message, true); }
};
