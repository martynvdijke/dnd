// @ts-nocheck — extracted from app.ts monolith

import { esc, capitalize, showModal, toast } from './lib/dom';
import { api } from './lib/api';
import type { CompendiumEntry, CompendiumSchema, CompendiumSearchResult } from './lib/api-types';
import { showView } from './navigation';
import { expose } from './lib/expose';

// ─── Compendium Global Search Navigation ───

expose('compendiumNavigateToSearchResult', function (type: string, id: number) {
  const tabMap: Record<string, string> = {
    spell: 'spells',
    equipment: 'equipment',
    monster: 'monsters',
    race: 'races',
    class: 'classes',
    feat: 'feats',
    background: 'backgrounds',
  };
  const tab = tabMap[type];
  if (tab) {
    (window as any).loadCompendiumTab(tab);
  }
  // Scroll to top of the compendium view for a fresh look
  const compView = document.getElementById('compendiumView');
  if (compView) {
    compView.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
});

// ─── Compendium ───

expose('showCompendium', function () {
  showView('compendium');
  loadCompendiumMonsters();
  loadCompendiumSchemaTypes();
});

expose('loadCompendiumTab', function (tab: string) {
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
});

// ─── Dynamic Schema Tabs (imported entries) ───

let activeSchemaTab: string | null = null;

async function loadCompendiumSchemaTypes() {
  try {
    const resp = await api<{ schemas: CompendiumSchema[] }>('GET', '/api/compendium/entries-by-schema');
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

expose('loadCompendiumSchemaTab', function (typeName: string) {
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
});

export function renderSchemaEntries(schema: any): string {
  const entries = schema.entries || [];
  return entries.map((e: any) => {
    const name = esc(e.data?.name || e.data?.Name || 'Unnamed');
    const preview = entryPreview(e.data, name);
    return `
      <div class="card mb-2 schema-entry-card" role="button"
           onclick="showCompendiumEntryDetail(${schema.id}, ${e.id})"
           style="cursor:pointer">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between">
            <strong>${name}</strong>
            <span class="badge bg-secondary align-self-center"><i class="fa-solid fa-book me-1"></i>View</span>
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
    const resp = await fetch('/htmx/compendium/races', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compRaces');
    if (container) {
      container.innerHTML = html;
      if (typeof (window as any).htmx !== 'undefined') {
        (window as any).htmx.process(container);
      }
    }
  } catch {}
}

async function loadCompendiumClasses() {
  try {
    const resp = await fetch('/htmx/compendium/classes', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compClasses');
    if (container) {
      container.innerHTML = html;
      if (typeof (window as any).htmx !== 'undefined') {
        (window as any).htmx.process(container);
      }
    }
  } catch {}
}

async function loadCompendiumSpells() {
  try {
    const resp = await fetch('/htmx/compendium/spells', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compSpells');
    if (container) {
      container.innerHTML = html;
      if (typeof (window as any).htmx !== 'undefined') {
        (window as any).htmx.process(container);
      }
    }
  } catch {}
}

async function loadCompendiumEquipment() {
  try {
    const resp = await fetch('/htmx/compendium/equipment', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compEquipment');
    if (container) {
      container.innerHTML = html;
      if (typeof (window as any).htmx !== 'undefined') {
        (window as any).htmx.process(container);
      }
    }
  } catch {}
}

async function loadCompendiumMonsters() {
  try {
    const resp = await fetch('/htmx/compendium-monsters', { credentials: 'include' });
    const html = await resp.text();
    document.getElementById('compMonsters')!.innerHTML = html;
  } catch {}
}

// ─── Player Entry Detail + Compendium Links (player-compendium-access) ───

export function humanizeKey(key: string): string {
  return key.replace(/[_-]+/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

expose('showCompendiumEntryDetail', async function (schemaId: number, entryId: number) {
  try {
    const entry = await api<CompendiumEntry>('GET', `/api/compendium/schemas/${schemaId}/entries/${entryId}`);
    const data = entry?.data || {};
    const name = (data.name || data.Name || 'Compendium Entry') as string;
    const rows = Object.entries(data)
      .filter(([k]) => k.toLowerCase() !== 'name')
      .map(([k, v]) => {
        const label = esc(humanizeKey(k));
        if (typeof v === 'string' && v.length > 80) {
          return `<div class="mb-2"><div class="fw-bold small">${label}</div><pre style="white-space:pre-wrap;margin:0">${esc(v)}</pre></div>`;
        }
        const val = typeof v === 'object' && v !== null ? JSON.stringify(v) : String(v);
        return `<div class="mb-2"><span class="fw-bold small">${label}:</span> <span>${esc(val)}</span></div>`;
      })
      .join('');
    showModal(esc(name), rows || '<p class="text-muted mb-0">No details.</p>');
  } catch (e: any) {
    toast(e.message || 'Could not load entry', true);
  }
});

const COMPENDIUM_LINK_RE = /\[\[compendium:([A-Za-z0-9_-]+):([^\]]+)\]\]/g;

export function parseCompendiumLinksInto(rootEl: Element): void {
  if (!rootEl || rootEl.querySelector('a.compendium-link')) return;
  const walker = document.createTreeWalker(rootEl, NodeFilter.SHOW_TEXT);
  const nodes: Text[] = [];
  let n: Node | null;
  while ((n = walker.nextNode())) nodes.push(n as Text);
  for (const textNode of nodes) {
    if (!COMPENDIUM_LINK_RE.test(textNode.nodeValue || '')) continue;
    COMPENDIUM_LINK_RE.lastIndex = 0;
    const frag = document.createDocumentFragment();
    let last = 0;
    let m: RegExpExecArray | null;
    while ((m = COMPENDIUM_LINK_RE.exec(textNode.nodeValue || '')) !== null) {
      frag.appendChild(document.createTextNode(textNode.nodeValue!.slice(last, m.index)));
      const a = document.createElement('a');
      a.href = '#';
      a.className = 'compendium-link';
      a.dataset.schema = m[1];
      a.dataset.name = m[2];
      a.textContent = m[2];
      frag.appendChild(a);
      last = m.index + m[0].length;
    }
    frag.appendChild(document.createTextNode(textNode.nodeValue!.slice(last)));
    textNode.parentNode?.replaceChild(frag, textNode);
  }
}

async function openCompendiumLink(schema: string, name: string): Promise<void> {
  const norm = (x: string) => (x || '').toLowerCase();
  try {
    const hits: CompendiumSearchResult[] = await api<CompendiumSearchResult[]>('GET', `/api/compendium/search?q=${encodeURIComponent(name)}`);
    const schemas: CompendiumSchema[] = await api<CompendiumSchema[]>('GET', '/api/compendium/schemas');
    const def = (schemas || []).find((s) => norm(s.type_name) === norm(schema));
    const matching = (hits || []).filter((r) => norm(r.type) === norm(schema));
    if (def) {
      // Legacy search hits can carry legacy table ids — try each match until
      // one resolves to a real compendium_entries row.
      for (const h of matching) {
        try {
          const v: CompendiumEntry = await api<CompendiumEntry>('GET', `/api/compendium/schemas/${def.id}/entries/${h.id}`);
          if (v && v.id) { await showCompendiumEntryDetail(def.id, h.id); return; }
        } catch (e) { /* 404 etc — try next hit */ }
      }
    }
    // Fallback: entries-by-schema scan (authoritative unified ids).
    const bySchema = await api<{ schemas: CompendiumSchema[] }>('GET', '/api/compendium/entries-by-schema');
    const es = ((bySchema && bySchema.schemas) || []).find((s: any) => norm(s.type_name) === norm(schema));
    const entry = es && (es.entries || []).find((e: any) => norm(e.data && e.data.name) === norm(name));
    if (entry) {
      await showCompendiumEntryDetail(es.id, entry.id);
      return;
    }
    toast('Compendium entry not found: ' + name, true);
  } catch (e: any) {
    toast(e.message || 'Could not open compendium entry', true);
  }
}

// Delegate clicks once (module import time).
document.addEventListener('click', (ev) => {
  const target = ev.target as HTMLElement;
  const link = target.closest?.('a.compendium-link') as HTMLAnchorElement | null;
  if (!link) return;
  ev.preventDefault();
  const schema = link.dataset.schema || '';
  const name = link.dataset.name || link.textContent || '';
  openCompendiumLink(schema, name);
});

// Parse [[compendium:...]] links in htmx-swapped content (act/scene descriptions etc.).
document.body.addEventListener('htmx:afterSwap', (e: any) => {
  const el = e?.detail?.elt;
  if (el && el.querySelector) parseCompendiumLinksInto(el);
});

expose('parseCompendiumLinksInto', parseCompendiumLinksInto);
