import { esc, capitalize, showModal, toast } from './lib/dom';
import { api } from './lib/api';
import type { CompendiumEntry, CompendiumSchema, CompendiumSearchResult } from './lib/api-types';
import { showView } from './navigation';
import { expose } from './lib/expose';

// ─── Compendium Global Search Navigation ───

expose('compendiumNavigateToSearchResult', function (type: string, _id: number): void {
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
    window.loadCompendiumTab(tab);
  }
  const compView = document.getElementById('compendiumView');
  if (compView) {
    compView.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
});

// ─── Compendium ───

expose('showCompendium', function (): void {
  showView('compendium');
  void loadCompendiumMonsters();
  void loadCompendiumSchemaTypes();
});

expose('loadCompendiumTab', function (tab: string): void {
  document.getElementById('compTabRaces')!.classList.remove('active');
  document.getElementById('compTabClasses')!.classList.remove('active');
  document.getElementById('compTabSpells')!.classList.remove('active');
  document.getElementById('compTabEquipment')!.classList.remove('active');
  document.getElementById('compTabMonsters')!.classList.remove('active');
  const tabEl = document.getElementById('compTab' + capitalize(tab));
  if (tabEl) tabEl.classList.add('active');
  (['races', 'classes', 'spells', 'equipment', 'monsters'] as const).forEach((s) => {
    const el = document.getElementById('comp' + capitalize(s));
    if (el) el.style.display = s === tab ? 'block' : 'none';
  });
  if (tab === 'races') void loadCompendiumRaces();
  if (tab === 'classes') void loadCompendiumClasses();
  if (tab === 'spells') void loadCompendiumSpells();
  if (tab === 'equipment') void loadCompendiumEquipment();
  if (tab === 'monsters') void loadCompendiumMonsters();
});

// ─── Dynamic Schema Tabs (imported entries) ───

let activeSchemaTab: string | null = null;

async function loadCompendiumSchemaTypes(): Promise<void> {
  try {
    const resp = await api<{ schemas: CompendiumSchema[] }>('GET', '/api/compendium/entries-by-schema');
    const schemas: CompendiumSchema[] = resp.schemas || [];
    const section = document.getElementById('compSchemaSection');
    const tabsEl = document.getElementById('compSchemaTabs');
    const contentEl = document.getElementById('compSchemaContent');
    if (!section || !tabsEl || !contentEl) return;

    if (schemas.length === 0) {
      section.style.display = 'none';
      return;
    }

    tabsEl.innerHTML = schemas.map((s, i) => `
      <li class="nav-item">
        <button class="nav-link ${i === 0 ? 'active' : ''}"
          id="compSchemaTab-${s.type_name}"
          onclick="loadCompendiumSchemaTab('${s.type_name}')"
          data-schema-id="${s.id}">
          ${esc(s.display_name)} <span class="badge bg-secondary">${s.entry_count}</span>
        </button>
      </li>`).join('');

    contentEl.innerHTML = schemas.map((s, i) => `
      <div id="compSchemaContent-${s.type_name}" style="display:${i === 0 ? 'block' : 'none'}">
        ${renderSchemaEntries(s)}
      </div>`).join('');

    section.style.display = 'block';
    activeSchemaTab = schemas[0]?.type_name || null;
    void activeSchemaTab;
  } catch { /* no schema entries — hide section */ }
}

expose('loadCompendiumSchemaTab', function (typeName: string): void {
  const tabsEl = document.getElementById('compSchemaTabs');
  if (!tabsEl) return;
  tabsEl.querySelectorAll('.nav-link').forEach((b) => b.classList.remove('active'));
  const tab = document.getElementById('compSchemaTab-' + typeName);
  if (tab) tab.classList.add('active');

  const contentEl = document.getElementById('compSchemaContent');
  if (!contentEl) return;
  contentEl.querySelectorAll('[id^="compSchemaContent-"]').forEach((el) => {
    (el as HTMLElement).style.display = 'none';
  });
  const pane = document.getElementById('compSchemaContent-' + typeName);
  if (pane) pane.style.display = 'block';
  activeSchemaTab = typeName;
});

export function renderSchemaEntries(schema: { id: number; entry_count: number; entries?: Array<{ id?: number; data?: unknown; [k: string]: unknown }>; type_name?: string; display_name?: string; [k: string]: unknown }): string {
  const entries = schema.entries || [];
  return entries.map((e) => {
    const data = e.data as Record<string, unknown>;
    const name = esc((data?.['name'] as string) || (data?.['Name'] as string) || 'Unnamed');
    const preview = entryPreview(data, name);
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

export function entryPreview(data: Record<string, unknown> | null | undefined, _name: string): string {
  if (!data) return '';
  const parts: string[] = [];
  for (const [key, val] of Object.entries(data)) {
    if (key.toLowerCase() === 'name') continue;
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

async function loadCompendiumRaces(): Promise<void> {
  try {
    const resp = await fetch('/htmx/compendium/races', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compRaces');
    if (container) {
      container.innerHTML = html;
      window.htmx.process(container);
    }
  } catch { /* ignore */ }
}

async function loadCompendiumClasses(): Promise<void> {
  try {
    const resp = await fetch('/htmx/compendium/classes', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compClasses');
    if (container) {
      container.innerHTML = html;
      window.htmx.process(container);
    }
  } catch { /* ignore */ }
}

async function loadCompendiumSpells(): Promise<void> {
  try {
    const resp = await fetch('/htmx/compendium/spells', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compSpells');
    if (container) {
      container.innerHTML = html;
      window.htmx.process(container);
    }
  } catch { /* ignore */ }
}

async function loadCompendiumEquipment(): Promise<void> {
  try {
    const resp = await fetch('/htmx/compendium/equipment', { credentials: 'include' });
    const html = await resp.text();
    const container = document.getElementById('compEquipment');
    if (container) {
      container.innerHTML = html;
      window.htmx.process(container);
    }
  } catch { /* ignore */ }
}

async function loadCompendiumMonsters(): Promise<void> {
  try {
    const resp = await fetch('/htmx/compendium-monsters', { credentials: 'include' });
    const html = await resp.text();
    document.getElementById('compMonsters')!.innerHTML = html;
  } catch { /* ignore */ }
}

// ─── Player Entry Detail + Compendium Links ───

export function humanizeKey(key: string): string {
  return key.replace(/[_-]+/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

async function showCompendiumEntryDetail(schemaId: number, entryId: number): Promise<void> {
  try {
    const entry = await api<CompendiumEntry>('GET', `/api/compendium/schemas/${schemaId}/entries/${entryId}`);
    const data = (entry?.data as Record<string, unknown>) || {};
    const name = ((data['name'] as string) || (data['Name'] as string) || 'Compendium Entry') as string;
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
  } catch (e: unknown) {
    toast(e instanceof Error ? e.message : 'Could not load entry', true);
  }
}
expose('showCompendiumEntryDetail', showCompendiumEntryDetail);

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
      a.dataset['schema'] = m[1]!;
      a.dataset['name'] = m[2]!;
      a.textContent = m[2]!;
      frag.appendChild(a);
      last = m.index + m[0].length;
    }
    frag.appendChild(document.createTextNode(textNode.nodeValue!.slice(last)));
    textNode.parentNode?.replaceChild(frag, textNode);
  }
}

async function openCompendiumLink(schema: string, name: string): Promise<void> {
  const norm = (x: string): string => (x || '').toLowerCase();
  try {
    const hits = await api<CompendiumSearchResult[]>('GET', `/api/compendium/search?q=${encodeURIComponent(name)}`);
    const schemas = await api<CompendiumSchema[]>('GET', '/api/compendium/schemas');
    const def = (schemas || []).find((s) => norm(s.type_name) === norm(schema));
    const matching = (hits || []).filter((r) => norm(r.type) === norm(schema));
    if (def) {
      for (const h of matching) {
        try {
          const v = await api<CompendiumEntry>('GET', `/api/compendium/schemas/${def.id}/entries/${h.id}`);
          if (v && v.id) { await showCompendiumEntryDetail(def.id, h.id); return; }
        } catch { /* 404 — try next */ }
      }
    }
    const bySchema = await api<{ schemas: CompendiumSchema[] }>('GET', '/api/compendium/entries-by-schema');
    const es = ((bySchema && bySchema.schemas) || []).find((s) => norm(s.type_name) === norm(schema));
    const entry = es && (es.entries || []).find((e) => norm((e.data as Record<string, unknown>)?.['name'] as string) === norm(name));
    if (entry && es) {
      await showCompendiumEntryDetail(es.id, entry.id);
      return;
    }
    toast('Compendium entry not found: ' + name, true);
  } catch (e: unknown) {
    toast(e instanceof Error ? e.message : 'Could not open compendium entry', true);
  }
}

// Delegate clicks once (module import time).
document.addEventListener('click', (ev) => {
  const target = ev.target as HTMLElement;
  const link = target.closest?.('a.compendium-link') as HTMLAnchorElement | null;
  if (!link) return;
  ev.preventDefault();
  const schema = link.dataset['schema'] || '';
  const name = link.dataset['name'] || link.textContent || '';
  void openCompendiumLink(schema, name);
});

// Parse [[compendium:...]] links in htmx-swapped content
document.body.addEventListener('htmx:afterSwap', (e: Event) => {
  const el = (e as CustomEvent).detail?.elt as Element | undefined;
  if (el) parseCompendiumLinksInto(el);
});

expose('parseCompendiumLinksInto', parseCompendiumLinksInto);
