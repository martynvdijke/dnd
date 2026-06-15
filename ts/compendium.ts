// @ts-nocheck — extracted from app.ts monolith

import { esc, capitalize } from './lib/dom';
import { api } from './lib/api';
import { showView } from './navigation';

// ─── Compendium ───

(window as any).showCompendium = function () {
  showView('compendium');
  loadCompendiumMonsters();
  loadCompendiumSchemaTypes();
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

// ─── Dynamic Schema Tabs (imported entries) ───

let activeSchemaTab: string | null = null;

async function loadCompendiumSchemaTypes() {
  try {
    const resp = await api('GET', '/api/compendium/entries-by-schema');
    const schemas: any[] = resp.schemas || [];
    const section = document.getElementById('compSchemaSection');
    const tabsEl = document.getElementById('compSchemaTabs');
    const contentEl = document.getElementById('compSchemaContent');
    if (!section || !tabsEl || !contentEl) return;

    if (schemas.length === 0) {
      section.style.display = 'none';
      return;
    }

    // Build tab buttons
    tabsEl.innerHTML = schemas.map((s, i) => `
      <li class="nav-item">
        <button class="nav-link ${i === 0 ? 'active' : ''}"
          id="compSchemaTab-${s.type_name}"
          onclick="loadCompendiumSchemaTab('${s.type_name}')"
          data-schema-id="${s.id}">
          ${esc(s.display_name)} <span class="badge bg-secondary">${s.entry_count}</span>
        </button>
      </li>`).join('');

    // Build content areas
    contentEl.innerHTML = schemas.map((s, i) => `
      <div id="compSchemaContent-${s.type_name}" style="display:${i === 0 ? 'block' : 'none'}">
        ${renderSchemaEntries(s)}
      </div>`).join('');

    // Show the section
    section.style.display = 'block';
    activeSchemaTab = schemas[0]?.type_name || null;
  } catch { /* no schema entries — hide section */ }
}

(window as any).loadCompendiumSchemaTab = function (typeName: string) {
  const tabsEl = document.getElementById('compSchemaTabs');
  if (!tabsEl) return;
  tabsEl.querySelectorAll('.nav-link').forEach(b => b.classList.remove('active'));
  const tab = document.getElementById('compSchemaTab-' + typeName);
  if (tab) tab.classList.add('active');

  const contentEl = document.getElementById('compSchemaContent');
  if (!contentEl) return;
  contentEl.querySelectorAll('[id^="compSchemaContent-"]').forEach(el => {
    (el as HTMLElement).style.display = 'none';
  });
  const pane = document.getElementById('compSchemaContent-' + typeName);
  if (pane) pane.style.display = 'block';
  activeSchemaTab = typeName;
};

export function renderSchemaEntries(schema: any): string {
  const entries = schema.entries || [];
  return entries.map((e: any) => {
    const name = esc(e.data?.name || e.data?.Name || 'Unnamed');
    const preview = entryPreview(e.data, name);
    return `
      <div class="card mb-2">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between">
            <strong>${name}</strong>
          </div>
          <p class="mb-0 mt-1 small text-muted">${preview}</p>
        </div>
      </div>`;
  }).join('')
  + (schema.entry_count > 20
    ? `<div class="mt-2">
        <a href="/api/compendium/schemas/${schema.id}/entries"
           class="btn btn-sm btn-outline-secondary"
           target="_blank">View All ${schema.entry_count} entries</a>
       </div>`
    : '');
}

export function entryPreview(data: any, name: string): string {
  if (!data) return '';
  const parts: string[] = [];
  for (const [key, val] of Object.entries(data)) {
    if (key.toLowerCase() === 'name' || key.toLowerCase() === 'name') continue;
    if (typeof val === 'string' && val.length > 0 && val.length < 80) {
      parts.push(`${key}: ${esc(val)}`);
    } else if (typeof val === 'number') {
      parts.push(`${key}: ${val}`);
    }
    if (parts.length >= 3) break;
  }
  return parts.join(' · ') || '';
}

// ─── Legacy Tab Loaders ───

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
