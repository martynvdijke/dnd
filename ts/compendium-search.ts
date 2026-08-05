// compendium-search.ts — reusable compendium-first search modal (compendium-first change)
// Queries the DM-scoped FTS5 search endpoint and returns the selected entry via a
// Promise. "Create Custom" resolves with null so callers can fall back to their
// existing custom-creation forms.
import { expose } from './lib/expose';
import { esc, hideModal } from './lib/dom';
import { api } from './lib/api';
import * as bootstrap from 'bootstrap';

export interface CompendiumSearchOptions {
  /** Client-side filter on the entry type (e.g. 'monster', 'spell', 'equipment'). */
  schemaType?: string;
  /** Optional server-side schema_id filter (compendium_entries.schema_id). */
  schemaId?: number;
  title?: string;
  context?: string;
  customLabel?: string;
}

export interface CompendiumSearchResult {
  id: number;
  type: number; // schema_id
  type_name: string;
  name: string;
  snippet?: string;
}

function badgeClassFor(typeName: string): string {
  const t = (typeName || '').toLowerCase();
  if (t.includes('monster')) return 'bg-danger';
  if (t.includes('spell')) return 'bg-info';
  if (t.includes('equip') || t.includes('item')) return 'bg-success';
  if (t.includes('feat')) return 'badge-gold';
  return 'bg-secondary';
}

// Bootstrap silently swallows Modal.show() while a hide transition is running
// (rapid close → reopen). Queue the show until the transition completes.
function openGenericModal(): void {
  const modalEl = document.getElementById('genericModal');
  if (!modalEl) return;
  const inst = bootstrap.Modal.getOrCreateInstance(modalEl);
  if (modalEl.classList.contains('show')) return;
  if ((inst as any)._isTransitioning) {
    const onHidden = () => {
      modalEl.removeEventListener('hidden.bs.modal', onHidden);
      inst.show();
    };
    modalEl.addEventListener('hidden.bs.modal', onHidden);
    return;
  }
  inst.show();
}

function openModalBody(title: string, bodyHtml: string): void {
  const titleEl = document.getElementById('genericModalTitle');
  if (titleEl) titleEl.textContent = title;
  const body = document.getElementById('genericModalBody');
  if (body) body.innerHTML = bodyHtml;
  openGenericModal();
}

/**
 * Shows the compendium search modal and resolves with the selected entry,
 * or `null` when the user chooses "Create Custom" (or dismisses the modal).
 */
export async function compendiumSearchModal(options: CompendiumSearchOptions = {}): Promise<CompendiumSearchResult | null> {
  return new Promise((resolve) => {
    let resolved = false;
    const finish = (value: CompendiumSearchResult | null) => {
      if (resolved) return;
      resolved = true;
      modalEl.removeEventListener('hidden.bs.modal', onHidden);
      // On "Create Custom" the caller immediately opens its own modal; hiding
      // first would leave that modal's show() swallowed by the hide transition.
      if (value !== null) hideModal();
      resolve(value);
    };
    const onHidden = () => finish(null);

    const title = options.title || 'Search Compendium';
    openModalBody(
      title,
      `
      <p class="text-muted small">${esc(options.context || 'Search the compendium and pick an entry, or create a custom one.')}</p>
      <div class="mb-2"><input class="form-control form-control-sm" id="csSearch" placeholder="Search entries..." autofocus></div>
      <div id="csResults" style="max-height:45vh;overflow-y:auto" class="mb-2">
        <div class="text-muted small py-2">Type to search the compendium...</div>
      </div>
      <button class="btn btn-outline-secondary w-100" id="csCustomBtn" type="button"><i class="fa-solid fa-pen me-1"></i>${esc(options.customLabel || 'Create Custom')}</button>
    `,
    );

    const modalEl = document.getElementById('genericModal')!;
    modalEl.addEventListener('hidden.bs.modal', onHidden);

    const input = document.getElementById('csSearch') as HTMLInputElement;
    const resultsEl = document.getElementById('csResults') as HTMLDivElement;
    let timer: ReturnType<typeof setTimeout> | null = null;

    const doSearch = async (q: string) => {
      if (!q.trim()) {
        resultsEl.innerHTML = '<div class="text-muted small py-2">Type to search the compendium...</div>';
        return;
      }
      resultsEl.innerHTML = '<div class="text-muted small py-2"><i class="fa-solid fa-spinner fa-spin me-1"></i>Searching...</div>';
      try {
        let results: CompendiumSearchResult[] = await api(
          'GET',
          `/api/compendium-search?q=${encodeURIComponent(q)}` + (options.schemaId ? `&schema_id=${options.schemaId}` : ''),
        );
        if (options.schemaType) {
          const t = options.schemaType.toLowerCase();
          results = results.filter((r) => (r.type_name || '').toLowerCase().includes(t));
        }
        if (!results.length) {
          resultsEl.innerHTML = '<div class="text-muted small py-2">No entries found.</div>';
          return;
        }
        resultsEl.innerHTML = results
          .map(
            (r, i) => `
          <div class="compendium-search-result cs-item" data-i="${i}" role="button">
            <div class="d-flex justify-content-between align-items-center gap-2">
              <span class="fw-bold">${esc(r.name || 'Unnamed')}</span>
              <span class="badge ${badgeClassFor(r.type_name)} flex-shrink-0">${esc(r.type_name || 'entry')}</span>
            </div>
            ${r.snippet ? `<div class="small text-muted">${esc(r.snippet)}</div>` : ''}
          </div>`,
          )
          .join('');
        (resultsEl as any)._results = results;
      } catch (e: any) {
        resultsEl.innerHTML = `<div class="text-danger small py-2">${esc(e.message || 'Search failed')}</div>`;
      }
    };

    input.addEventListener('input', () => {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => doSearch(input.value), 250);
    });

    resultsEl.addEventListener('click', (ev) => {
      const item = (ev.target as HTMLElement).closest('.cs-item') as HTMLElement;
      if (!item) return;
      const idx = parseInt(item.dataset.i || '-1', 10);
      const results: CompendiumSearchResult[] = (resultsEl as any)._results || [];
      const r = results[idx];
      if (r) finish(r);
    });

    document.getElementById('csCustomBtn')!.addEventListener('click', () => finish(null));
  });
}

expose('compendiumSearchModal', compendiumSearchModal);
