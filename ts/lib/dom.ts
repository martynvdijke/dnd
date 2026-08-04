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

export function toast(msg: string, isError = false, duration = 5000): void {
  const container = document.getElementById('toastContainer')!;
  const id = 'toast-' + Date.now();
  const bg = isError ? 'bg-danger' : 'bg-success';
  container.innerHTML += `
    <div class="toast align-items-center text-white ${bg} border-0 mb-2" id="${id}" role="alert">
      <div class="d-flex">
        <div class="toast-body">${msg}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
      </div>
    </div>`;
  const el = document.getElementById(id)!;
  new bootstrap.Toast(el, { autohide: true, delay: duration }).show();
  setTimeout(() => el.remove(), duration + 1000);
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
