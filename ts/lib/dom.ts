// @ts-ignore — bootstrap has no type declarations, used for modal handling
import * as bootstrap from 'bootstrap';

let modalEl: HTMLElement | null = null;
let loadingCount = 0;
let loadingTimer: ReturnType<typeof setTimeout> | null = null;

function getModal(): any {
  if (!modalEl) modalEl = document.getElementById('genericModal');
  let inst = bootstrap.Modal.getInstance(modalEl!);
  if (!inst) inst = new bootstrap.Modal(modalEl!, { backdrop: true, keyboard: true });
  return inst;
}

export function esc(s: string | null | undefined): string {
  if (!s) return '';
  const d = document.createElement('div'); d.textContent = s; return d.innerHTML;
}

export function attrEscape(s: string | null | undefined): string {
  if (!s) return '';
  return String(s).replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/'/g, '&#39;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

export function jsEscape(s: string | null | undefined): string {
  if (!s) return '';
  return String(s)
    .replace(/\\/g, '\\\\')
    .replace(/'/g, "\\'")
    .replace(/"/g, '\\"')
    .replace(/\n/g, '\\n')
    .replace(/\r/g, '\\r')
    .replace(/<\/script/gi, '<\\/script')
    .replace(/&/g, '\\x26')
    .replace(/</g, '\\x3c')
    .replace(/>/g, '\\x3e');
}

export function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

export function showModal(title: string, bodyHtml: string): void {
  const modal = getModal();
  document.getElementById('genericModalTitle')!.textContent = title;
  const body = document.getElementById('genericModalBody')!;
  body.innerHTML = bodyHtml;
  // Process any HTMX attributes (hx-trigger="load", hx-get, etc.) in the
  // newly inserted content. Without this, HTMX is unaware of elements
  // added via innerHTML and hx-trigger="load" never fires.
  const htmx = (window as any).htmx;
  if (htmx && typeof htmx.process === 'function') {
    htmx.process(body);
  }
  modal.show();
}

export function hideModal(): void {
  getModal().hide();
  document.querySelectorAll('.modal-backdrop').forEach(el => el.remove());
  document.body.classList.remove('modal-open');
  document.body.style.removeProperty('padding-right');
}

export type ToastOpts = { html?: boolean; isError?: boolean; duration?: number };
export function toast(msg: string, opts?: ToastOpts | boolean, duration = 5000): void {
  let html = false;
  let isError = false;
  let dur = duration;
  if (typeof opts === 'boolean') {
    isError = opts;
  } else if (opts && typeof opts === 'object') {
    html = !!opts.html;
    if (typeof opts.isError === 'boolean') isError = opts.isError;
    if (typeof opts.duration === 'number') dur = opts.duration;
  }
  // Back-compat: toast(msg, isError, duration) where second arg is boolean
  // If third arg meaning duration was passed positionally with boolean second, use it.
  // When opts is boolean, duration param already holds the positional duration.
  dur = dur ?? duration;
  const safe = html ? msg : esc(msg);
  const container = document.getElementById('toastContainer')!;
  const id = 'toast-' + Date.now();
  const bg = isError ? 'bg-danger' : 'bg-success';
  container.innerHTML += `
    <div class="toast align-items-center text-white ${bg} border-0 mb-2" id="${id}" role="alert">
      <div class="d-flex">
        <div class="toast-body">${safe}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
      </div>
    </div>`;
  const el = document.getElementById(id)!;
  new bootstrap.Toast(el, { autohide: true, delay: dur }).show();

  // Swipe-left to dismiss (touch only — mouse users unaffected).
  let startX = 0;
  let startY = 0;
  let swiping = false;
  el.addEventListener('touchstart', (e: TouchEvent) => {
    startX = e.touches[0].clientX;
    startY = e.touches[0].clientY;
    swiping = true;
  }, { passive: true });
  el.addEventListener('touchmove', (e: TouchEvent) => {
    if (!swiping) return;
    const dx = e.touches[0].clientX - startX;
    const dy = e.touches[0].clientY - startY;
    if (Math.abs(dx) > Math.abs(dy) && dx < 0) {
      el.style.transition = 'none';
      el.style.transform = `translateX(${dx}px)`;
      el.style.opacity = String(Math.max(0.2, 1 + dx / 120));
    }
  }, { passive: true });
  el.addEventListener('touchend', (e: TouchEvent) => {
    if (!swiping) return;
    swiping = false;
    const dx = e.changedTouches[0].clientX - startX;
    if (dx < -60) {
      el.style.transition = 'transform 0.22s ease, opacity 0.22s ease';
      el.style.transform = 'translateX(110%)';
      el.style.opacity = '0';
      window.setTimeout(() => el.remove(), 230);
    } else {
      el.style.transition = 'transform 0.2s ease';
      el.style.transform = '';
      el.style.opacity = '';
    }
  }, { passive: true });
  setTimeout(() => el.remove(), dur + 1000);
}

export function showLoading(): void {
  loadingCount++;
  const overlay = document.getElementById('loadingOverlay');
  if (!overlay || loadingCount > 1 || !overlay.classList.contains('d-none')) return;
  // 200ms debounce: skip the skeleton for fast requests to avoid flicker.
  loadingTimer = setTimeout(() => overlay.classList.remove('d-none'), 200);
}

export function hideLoading(): void {
  loadingCount = Math.max(0, loadingCount - 1);
  if (loadingCount > 0) return;
  if (loadingTimer) {
    clearTimeout(loadingTimer);
    loadingTimer = null;
  }
  const overlay = document.getElementById('loadingOverlay');
  if (overlay) overlay.classList.add('d-none');
}

export function $id<T extends HTMLElement>(id: string): T | null {
  return document.getElementById(id) as T | null;
}

export function $idStrict<T extends HTMLElement>(id: string): T {
  const el = document.getElementById(id) as T | null;
  if (!el) throw new Error(`Missing required element: #${id}`);
  return el;
}

export interface CompendiumPickerOptions {
  title: string;
  placeholder: string;
  search: (q: string) => Promise<any[]>;
  render: (item: any) => string;
  onPick: (item: any) => void;
}

export function openCompendiumPicker(opts: CompendiumPickerOptions): void {
  let results: any[] = [];
  let timer: ReturnType<typeof setTimeout> | null = null;
  const doSearch = async (q: string) => {
    try {
      results = await opts.search(q);
    } catch {
      results = [];
    }
    const list = document.getElementById('cpResults');
    if (!list) return;
    list.innerHTML = results.length
      ? results.map((r, i) => `<div class="cp-item" data-i="${i}" role="button" tabindex="0">${opts.render(r)}</div>`).join('')
      : '<div class="text-muted small fst-italic p-2">No results.</div>';
  };
  showModal(opts.title, `
    <input type="search" class="form-control mb-2" id="cpSearch" placeholder="${opts.placeholder}" autocomplete="off">
    <div id="cpResults" class="cp-list" style="max-height:50vh;overflow-y:auto"></div>
  `);
  const input = document.getElementById('cpSearch') as HTMLInputElement;
  const list = document.getElementById('cpResults')!;
  input.addEventListener('input', () => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => doSearch(input.value.trim()), 250);
  });
  list.addEventListener('click', (e) => {
    const row = (e.target as HTMLElement).closest('.cp-item') as HTMLElement | null;
    if (!row) return;
    const item = results[+row.dataset.i!];
    if (item) {
      hideModal();
      opts.onPick(item);
    }
  });
  doSearch('');
}
