// @ts-nocheck — extracted from app.ts monolith

import { esc, capitalize } from './lib/dom';
import { api } from './lib/api';
import { showView } from './navigation';

// ─── Compendium ───

(window as any).showCompendium = function () {
  showView('compendium');
  loadCompendiumMonsters();
};

(window as any).loadCompendiumTab = function (tab: string) {
  document.getElementById('compTabRaces')!.classList.remove('active');
  document.getElementById('compTabClasses')!.classList.remove('active');
  document.getElementById('compTabSpells')!.classList.remove('active');
  document.getElementById('compTabEquipment')!.classList.remove('active');
  document.getElementById('compTabMonsters')!.classList.remove('active');
  const tabEl = document.getElementById('compTab' + capitalize(tab));
  if (tabEl) tabEl.classList.add('active');
  ['races', 'classes', 'spells', 'equipment', 'monsters'].forEach(s => {
    const el = document.getElementById('comp' + capitalize(s));
    if (el) el.style.display = s === tab ? 'block' : 'none';
  });
  if (tab === 'races') loadCompendiumRaces();
  if (tab === 'classes') loadCompendiumClasses();
  if (tab === 'spells') loadCompendiumSpells();
  if (tab === 'equipment') loadCompendiumEquipment();
  if (tab === 'monsters') loadCompendiumMonsters();
};

async function loadCompendiumRaces() {
  try {
    const races = await api('GET', '/api/compendium/races');
    document.getElementById('compRaces')!.innerHTML = races.map((r: any) => `
      <div class="card mb-2">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between"><strong>${esc(r.name)}</strong>
            <span><span class="badge badge-gold">${esc(r.size)}</span> Speed: ${r.speed}</span></div>
          <p class="mb-0 mt-1 small text-muted">${esc(r.description)}</p>
        </div>
      </div>`).join('');
  } catch {}
}

async function loadCompendiumClasses() {
  try {
    const cls = await api('GET', '/api/compendium/classes');
    document.getElementById('compClasses')!.innerHTML = cls.map((c: any) => `
      <div class="card mb-2">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between"><strong>${esc(c.name)}</strong>
            <span class="text-muted small">d${c.hit_die} · ${esc(c.primary_ability)}</span></div>
          <p class="mb-0 mt-1 small text-muted">${esc(c.description)}</p>
        </div>
      </div>`).join('');
  } catch {}
}

async function loadCompendiumSpells() {
  try {
    const spells = await api('GET', '/api/compendium/spells');
    document.getElementById('compSpells')!.innerHTML = spells.map((s: any) => `
      <div class="inv-item">
        <div><span class="fw-bold">${esc(s.name)}</span> <span class="text-muted small">Lv${s.level} ${esc(s.school)}</span></div>
        <div class="text-muted small">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.duration)}</div>
      </div>`).join('');
  } catch {}
}

async function loadCompendiumEquipment() {
  try {
    const items = await api('GET', '/api/compendium/equipment');
    document.getElementById('compEquipment')!.innerHTML = items.map((i: any) => `
      <div class="inv-item">
        <span class="fw-bold">${esc(i.name)}</span>
        <span class="text-muted small">${esc(i.category)}${i.weight ? ' · ' + i.weight + 'lb' : ''}</span>
      </div>`).join('');
  } catch {}
}

async function loadCompendiumMonsters() {
  try {
    const resp = await fetch('/htmx/compendium-monsters', { credentials: 'include' });
    const html = await resp.text();
    document.getElementById('compMonsters')!.innerHTML = html;
  } catch {}
}
