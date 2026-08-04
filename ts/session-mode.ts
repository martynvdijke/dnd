import type { SessionModeState } from './types';
import { updateFabForView } from './fab';
import { getCurrentView } from './navigation';
import { expose } from './lib/expose';

let state: SessionModeState = 'normal';

const STORAGE_KEY = 'villum-session-mode';
const AUTO_DISABLE_KEY = 'villum-session-auto-disable';

export function isSessionMode(): boolean {
  return state === 'session';
}

export function getSessionState(): SessionModeState {
  return state;
}

export function toggleSessionMode(): void {
  if (state === 'normal') {
    activateSessionMode();
  } else {
    deactivateSessionMode();
  }
}

export function activateSessionMode(): void {
  state = 'session';
  sessionStorage.setItem(STORAGE_KEY, 'session');
  document.body.classList.add('session-mode');
  showSessionToast('Session Mode activated');
  updateFabForView(getCurrentView(), true);
}

export function deactivateSessionMode(): void {
  state = 'normal';
  sessionStorage.removeItem(STORAGE_KEY);
  document.body.classList.remove('session-mode');
  updateFabForView(getCurrentView(), false);
}

export function showSessionToast(msg: string): void {
  const container = document.getElementById('toastContainer');
  if (!container) return;
  const id = 'toast-' + Date.now();
  container.innerHTML += `
    <div class="toast align-items-center text-white bg-primary border-0 mb-2" id="${id}" role="alert">
      <div class="d-flex">
        <div class="toast-body">${msg}</div>
        <button type="button" class="btn btn-sm btn-outline-light me-2 undo-session-btn" onclick="deactivateSessionMode();this.closest('.toast').remove()">Undo</button>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
      </div>
    </div>`;
  const el = document.getElementById(id);
  if (el) {
    setTimeout(() => el.remove(), 5000);
  }
}

export function isAutoActivateDisabled(): boolean {
  return localStorage.getItem(AUTO_DISABLE_KEY) === 'true';
}

export function setAutoActivateDisabled(disabled: boolean): void {
  localStorage.setItem(AUTO_DISABLE_KEY, disabled ? 'true' : 'false');
}

export function handleAutoActivate(): void {
  if (isAutoActivateDisabled()) return;
  if (state === 'session') return;
  activateSessionMode();
}

export function restoreSessionMode(): void {
  const saved = sessionStorage.getItem(STORAGE_KEY);
  if (saved === 'session') {
    activateSessionMode();
  }
}

expose('toggleSessionMode', toggleSessionMode);
expose('activateSessionMode', activateSessionMode);
expose('deactivateSessionMode', deactivateSessionMode);
expose('isSessionMode', isSessionMode);
