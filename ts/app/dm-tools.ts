import { expose } from '../lib/expose';
import { esc, showModal } from '../lib/dom';
import { api } from '../lib/api';

// ─── DM Tools: generators ───

expose('showDmTools', function () {
  showModal('DM Tools', `
    <div class="row g-2 mb-3">
      <div class="col-6">
        <label class="form-label small mb-1">Treasure hoard</label>
        <div class="input-group input-group-sm">
          <select class="form-select" id="dmTreasureTier" data-testid="dm-treasure-tier">
            <option value="1">Tier 1 (CR 0-4)</option>
            <option value="2">Tier 2 (CR 5-10)</option>
            <option value="3">Tier 3 (CR 11-16)</option>
            <option value="4">Tier 4 (CR 17+)</option>
          </select>
          <button class="btn btn-outline-gold" data-testid="dm-tool-treasure" onclick="dmGenerateTreasure()">Roll</button>
        </div>
      </div>
      <div class="col-6">
        <label class="form-label small mb-1">Weather</label>
        <div class="input-group input-group-sm">
          <select class="form-select" id="dmWeatherBiome" data-testid="dm-weather-biome">
            <option value="temperate">Temperate</option>
            <option value="arctic">Arctic</option>
            <option value="desert">Desert</option>
            <option value="forest">Forest</option>
            <option value="mountain">Mountain</option>
            <option value="swamp">Swamp</option>
            <option value="coast">Coast</option>
            <option value="underdark">Underdark</option>
          </select>
          <select class="form-select" id="dmWeatherSeason" data-testid="dm-weather-season">
            <option value="Spring">Spring</option>
            <option value="Summer">Summer</option>
            <option value="Autumn">Autumn</option>
            <option value="Winter">Winter</option>
          </select>
          <button class="btn btn-outline-gold" data-testid="dm-tool-weather" onclick="dmGenerateWeather()">Roll</button>
        </div>
      </div>
      <div class="col-6">
        <label class="form-label small mb-1">Name</label>
        <div class="input-group input-group-sm">
          <select class="form-select" id="dmNameRace" data-testid="dm-name-race">
            <option value="human">Human</option>
            <option value="elf">Elf</option>
            <option value="dwarf">Dwarf</option>
            <option value="halfling">Halfling</option>
            <option value="orc">Orc</option>
            <option value="dragonborn">Dragonborn</option>
            <option value="tiefling">Tiefling</option>
            <option value="gnome">Gnome</option>
          </select>
          <button class="btn btn-outline-gold" data-testid="dm-tool-name" onclick="dmGenerateName()">Roll</button>
        </div>
      </div>
      <div class="col-6">
        <label class="form-label small mb-1">Adventuring-day XP budget</label>
        <div class="input-group input-group-sm">
          <input class="form-control" id="dmBudgetLevels" data-testid="dm-budget-levels" placeholder="e.g. 3,3,3,3" value="3,3,3,3">
          <button class="btn btn-outline-gold" data-testid="dm-tool-budget" onclick="dmDailyBudget()">Calc</button>
        </div>
      </div>
    </div>
    <div class="border rounded p-2 mb-3">
      <label class="form-label small mb-1">One-shot generators</label>
      <div class="d-flex flex-wrap gap-1">
        <button class="btn btn-sm btn-outline-gold" data-testid="dm-tool-hook" onclick="dmGenerateHook()">Adventure Hook</button>
        <button class="btn btn-sm btn-outline-gold" data-testid="dm-tool-dressing" onclick="dmGenerateDungeonDressing()">Dungeon Dressing</button>
        <button class="btn btn-sm btn-outline-gold" data-testid="dm-tool-tavern" onclick="dmGenerateTavern()">Tavern</button>
        <button class="btn btn-sm btn-outline-gold" data-testid="dm-tool-urban" onclick="dmGenerateUrbanEncounter()">Urban Encounter</button>
        <button class="btn btn-sm btn-outline-gold" data-testid="dm-tool-road" onclick="dmGenerateRoadEncounter()">Road Encounter</button>
      </div>
    </div>
    <div id="dmToolsResult" class="small"></div>
  `);
});

expose('dmGenerateTreasure', async function () {
  const tier = (document.getElementById('dmTreasureTier') as HTMLSelectElement).value;
  const el = document.getElementById('dmToolsResult')!;
  try {
    const t = await api<any>('GET', `/api/generate/treasure?tier=${tier}`);
    const c = t.coins || {};
    el.innerHTML = `
      <div class="dash-card">
        <h6>Treasure Hoard — Tier ${t.tier}</h6>
        <div>Coins: ${c.cp || 0} cp · ${c.sp || 0} sp · ${c.ep || 0} ep · ${c.gp || 0} gp · ${c.pp || 0} pp</div>
        <div>Gems: ${(t.gems || []).map((g: string) => esc(g)).join(', ') || '—'}</div>
        <div>Art: ${(t.art_objects || []).map((a: string) => esc(a)).join(', ') || '—'}</div>
        <div>Magic: ${(t.magic_items || []).map((m: string) => esc(m)).join(', ') || '—'}</div>
        <div class="fw-bold mt-1">Total value ≈ ${t.total_gp_value} gp</div>
      </div>`;
  } catch (e: any) {
    el.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

expose('dmGenerateWeather', async function () {
  const biome = (document.getElementById('dmWeatherBiome') as HTMLSelectElement).value;
  const season = (document.getElementById('dmWeatherSeason') as HTMLSelectElement).value;
  const el = document.getElementById('dmToolsResult')!;
  try {
    const w = await api<any>('GET', `/api/generate/weather?biome=${biome}&season=${season}`);
    el.innerHTML = `
      <div class="dash-card">
        <h6>Weather — ${esc(w.biome)} ${esc(w.season)}</h6>
        <div>${esc(w.temperature)} · ${esc(w.sky)} · ${esc(w.precipitation)} · Wind: ${esc(w.wind)}</div>
        <div class="mt-1">${esc(w.description)}</div>
      </div>`;
  } catch (e: any) {
    el.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

expose('dmGenerateName', async function () {
  const race = (document.getElementById('dmNameRace') as HTMLSelectElement).value;
  const el = document.getElementById('dmToolsResult')!;
  try {
    const n = await api<any>('GET', `/api/generate/name?race=${race}`);
    el.innerHTML = `<div class="dash-card"><h6>Name</h6><div class="fs-5">${esc(n.name)}</div></div>`;
  } catch (e: any) {
    el.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

expose('dmDailyBudget', async function () {
  const levels = (document.getElementById('dmBudgetLevels') as HTMLInputElement).value;
  const el = document.getElementById('dmToolsResult')!;
  try {
    const b = await api<any>('GET', `/api/encounters/daily-budget?levels=${encodeURIComponent(levels)}`);
    el.innerHTML = `
      <div class="dash-card">
        <h6>Adventuring-Day Budget</h6>
        <div class="dash-value">${b.party_budget} XP</div>
        <div class="text-muted">${esc(b.suggested_encounters || 'No valid levels supplied')}</div>
      </div>`;
  } catch (e: any) {
    el.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

// ─── DM Tools: one-shot generators ───

function dmOneShotResult(targetId: string, html: string): void {
  const el = document.getElementById(targetId);
  if (el) el.innerHTML = html;
}

function dmOneShotError(targetId: string, message: string): void {
  dmOneShotResult(targetId, `<p class="text-danger">${esc(message)}</p>`);
}

expose('dmGenerateHook', async function (targetId = 'dmToolsResult') {
  try {
    const h = await api<any>('GET', '/api/generate/adventure-hook');
    dmOneShotResult(targetId, `
      <div class="dash-card">
        <h6>${esc(h.hook_name || 'Adventure Hook')} <span class="badge bg-secondary">${esc(h.hook_type || '')}</span></h6>
        <div><strong>Villain:</strong> ${esc(h.villain || '—')}</div>
        <div><strong>MacGuffin:</strong> ${esc(h.macguffin || '—')}</div>
        <div><strong>Stakes:</strong> ${esc(h.stakes || '—')}</div>
        <div><strong>Location:</strong> ${esc(h.location_hint || '—')}</div>
        <div><strong>Twist:</strong> ${esc(h.twist || '—')}</div>
      </div>`);
  } catch (e: any) { dmOneShotError(targetId, e.message); }
});

expose('dmGenerateDungeonDressing', async function (targetId = 'dmToolsResult') {
  try {
    const d = await api<any>('GET', '/api/generate/dungeon-dressing');
    dmOneShotResult(targetId, `
      <div class="dash-card">
        <h6>Dungeon Dressing</h6>
        <div><strong>Room:</strong> ${esc(d.room_type || '—')} · ${esc(d.size || '')} · ${esc(d.shape || '')}</div>
        <div><strong>Floor:</strong> ${esc(d.floor || '—')} · <strong>Walls:</strong> ${esc(d.walls || '—')} · <strong>Ceiling:</strong> ${esc(d.ceiling || '—')}</div>
        <div><strong>Sound:</strong> ${esc(d.sound || '—')} · <strong>Smell:</strong> ${esc(d.smell || '—')}</div>
        <div><strong>Light:</strong> ${esc(d.light || '—')} · <strong>Temperature:</strong> ${esc(d.temperature || '—')}</div>
        <div><strong>Debris:</strong> ${esc(d.debris || '—')}</div>
      </div>`);
  } catch (e: any) { dmOneShotError(targetId, e.message); }
});

expose('dmGenerateTavern', async function (targetId = 'dmToolsResult') {
  try {
    const t = await api<any>('GET', '/api/generate/tavern');
    const clientele = (t.clientele || []).map((c: string) => esc(c)).join(', ') || '—';
    const rumors = (t.rumors || []).map((r: string) => esc(r)).join(' · ') || '—';
    dmOneShotResult(targetId, `
      <div class="dash-card">
        <h6>${esc(t.name || 'Tavern')}</h6>
        <div><strong>Proprietor:</strong> ${esc(t.proprietor || '—')} — ${esc(t.proprietor_trait || '')}</div>
        <div><strong>Clientele:</strong> ${clientele}</div>
        <div><strong>Drink:</strong> ${esc(t.specialty_drink || '—')} — ${esc(t.drink_description || '')}</div>
        <div><strong>Atmosphere:</strong> ${esc(t.atmosphere || '—')}</div>
        <div><strong>Prices:</strong> ${esc(t.prices || '—')}</div>
        <div><strong>Rumors:</strong> ${rumors}</div>
      </div>`);
  } catch (e: any) { dmOneShotError(targetId, e.message); }
});

expose('dmGenerateUrbanEncounter', async function (targetId = 'dmToolsResult') {
  try {
    const u = await api<any>('GET', '/api/generate/urban-encounter');
    dmOneShotResult(targetId, `
      <div class="dash-card">
        <h6>Urban Encounter <span class="badge bg-secondary">${esc(u.theme || '')}</span></h6>
        <div><strong>NPC:</strong> ${esc(u.npc || '—')}</div>
        <div>${esc(u.description || '')}</div>
        <div><strong>Complication:</strong> ${esc(u.complication || '—')}</div>
        <div><strong>Resolution:</strong> ${esc(u.possible_resolution || '—')}</div>
      </div>`);
  } catch (e: any) { dmOneShotError(targetId, e.message); }
});

expose('dmGenerateRoadEncounter', async function (targetId = 'dmToolsResult') {
  try {
    const r = await api<any>('GET', '/api/generate/road-encounter');
    dmOneShotResult(targetId, `
      <div class="dash-card">
        <h6>Road Encounter <span class="badge bg-secondary">${esc(r.terrain || '')}</span> <span class="badge bg-secondary">${esc(r.encounter_type || '')}</span></h6>
        <div>${esc(r.description || '')}</div>
        <div><strong>Creatures:</strong> ${esc(r.creatures || '—')}</div>
        <div><strong>Loot hint:</strong> ${esc(r.loot_hint || '—')}</div>
        <div><strong>Complication:</strong> ${esc(r.complication || '—')}</div>
      </div>`);
  } catch (e: any) { dmOneShotError(targetId, e.message); }
});

// ─── DM Screen: live campaign aggregate ───

let dmScreenTimer: number | null = null;

function stopDmScreenRefresh(): void {
  if (dmScreenTimer !== null) {
    clearInterval(dmScreenTimer);
    dmScreenTimer = null;
  }
}

async function renderDmScreen(campaignId: number): Promise<void> {
  const el = document.getElementById('dmScreenContent');
  if (!el) {
    stopDmScreenRefresh();
    return;
  }
  try {
    const d = await api<any>('GET', `/api/campaigns/${campaignId}/dm-screen`);
    const hpPct = (h: number, m: number) => (m > 0 ? Math.round((h / m) * 100) : 0);
    const combat = d.combat || {};
    const session = d.session;
    el.innerHTML = `
      <div class="dash-grid">
        <div class="dash-card">
          <h6>${esc(d.campaign?.name || 'Campaign')}${d.campaign?.party_name ? ` — ${esc(d.campaign.party_name)}` : ''}</h6>
          <div class="small text-muted">${esc(d.campaign?.dm_notes || 'No campaign notes.')}</div>
        </div>
        <div class="dash-card">
          <h6>Current Turn</h6>
          <div class="dash-value">${esc(combat.current_turn || '—')}</div>
          <div class="small text-muted">${combat.active ? `${(combat.entries || []).length} in combat` : 'No active combat'}</div>
        </div>
        <div class="dash-card">
          <h6>Session Plan</h6>
          ${session ? `
            <div class="fw-bold">${esc(session.title)}</div>
            <div class="small text-muted">${esc(session.dm_notes || '')}</div>
          ` : '<div class="text-muted small">No active session plan.</div>'}
        </div>
        <div class="dash-card">
          <h6>Party</h6>
          ${(d.party || []).map((p: any) => `
            <div class="d-flex justify-content-between align-items-center border-bottom py-1">
              <div>
                <span class="fw-bold">${esc(p.name)}</span>
                ${(p.conditions || []).length ? `<span class="badge badge-muted ms-1">${(p.conditions || []).map((c: string) => esc(c)).join(', ')}</span>` : ''}
              </div>
              <span class="small">${p.hp_current}/${p.hp_max} HP · AC ${p.ac}</span>
            </div>
            <div class="dash-hp-bar"><div class="dash-hp-bar-fill${hpPct(p.hp_current, p.hp_max) < 30 ? ' low-hp' : ''}" style="width:${hpPct(p.hp_current, p.hp_max)}%"></div></div>
          `).join('') || '<div class="text-muted small">No characters.</div>'}
        </div>
        <div class="dash-card">
          <h6>Ambience</h6>
          <div>${d.ambience?.track ? `${esc(d.ambience.track)} (${esc(d.ambience.action)})` : 'Silent'}</div>
          <button class="btn btn-sm btn-outline-light mt-2" onclick="showAmbienceControls(${campaignId})"><i class="fa-solid fa-music me-1"></i>Ambience</button>
        </div>
      </div>
      <div class="text-center mt-3">
        <button class="btn btn-sm btn-outline-light" data-testid="dm-screen-refresh" onclick="refreshDmScreen(${campaignId})"><i class="fa-solid fa-rotate me-1"></i>Refresh</button>
      </div>`;
  } catch (e: any) {
    el.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
}

expose('refreshDmScreen', function (campaignId: number) {
  void renderDmScreen(campaignId);
});

expose('showDmScreen', async function (campaignId: number) {
  showModal('DM Screen', `<div id="dmScreenContent" data-testid="dm-screen-content"><div class="ornament">✧ Loading DM screen... ✧</div></div>`);
  await renderDmScreen(campaignId);
  stopDmScreenRefresh();
  // ponytail: 15s polling instead of a WS subscription; upgrade if noisy.
  dmScreenTimer = window.setInterval(() => { void renderDmScreen(campaignId); }, 15000);
  document.getElementById('genericModal')?.addEventListener('hidden.bs.modal', stopDmScreenRefresh, { once: true });
});
