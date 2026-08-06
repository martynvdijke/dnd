// Centralized auto-save module — dirty tracking, scheduler, save button state.
// Extracted from characters/sheet.ts (improve-auto-save).
import { expose } from './expose';
import { currentChar, setCurrentChar } from './state';
import { showModal, toast } from './dom';
import { api } from './api';

let sheetDirty = false;
let saving = false;
let schedulerHandle: number | null = null;
let saveTimer: number | null = null;

const INTERVAL_KEY = 'villum-autosave-interval';

export function getAutoSaveInterval(): number {
  const raw = localStorage.getItem(INTERVAL_KEY);
  const n = parseInt(raw || '', 10);
  return !isNaN(n) && n >= 5 && n <= 300 ? n : 12;
}

export function setAutoSaveInterval(seconds: number): void {
  const s = Math.min(300, Math.max(5, Math.round(seconds)));
  localStorage.setItem(INTERVAL_KEY, String(s));
  if (schedulerHandle !== null) {
    window.clearInterval(schedulerHandle);
    schedulerHandle = window.setInterval(autoSaveScheduler, s * 1000);
  }
  window.dispatchEvent(new CustomEvent('villum-savestate'));
}

export function isDirty(): boolean {
  return sheetDirty;
}

export function isSaving(): boolean {
  return saving;
}

export function markDirty(): void {
  sheetDirty = true;
  ensureScheduler();
  window.dispatchEvent(new CustomEvent('villum-savestate'));
}

export function markClean(): void {
  sheetDirty = false;
  window.dispatchEvent(new CustomEvent('villum-savestate'));
}

function ensureScheduler(): void {
  if (schedulerHandle !== null) return;
  schedulerHandle = window.setInterval(autoSaveScheduler, getAutoSaveInterval() * 1000);
}

export async function saveCharacter(): Promise<void> {
  if (!currentChar || !currentChar.id || saving) return;
  saving = true;
  window.dispatchEvent(new CustomEvent('villum-savestate'));
  try {
    const updated = await api('PUT', `/api/characters/${currentChar.id}`, currentChar);
    setCurrentChar(updated);
    markClean();
  } catch (e: any) {
    toast(e.message || 'Save failed', true);
  } finally {
    saving = false;
    window.dispatchEvent(new CustomEvent('villum-savestate'));
  }
}

function autoSaveScheduler(): void {
  if (sheetDirty && !saving) {
    saveCharacter();
  }
}

export function startAutoSave(): void {
  ensureScheduler();
  window.dispatchEvent(new CustomEvent('villum-savestate'));
}

export function stopAutoSave(): void {
  if (schedulerHandle !== null) {
    window.clearInterval(schedulerHandle);
    schedulerHandle = null;
  }
  if (saveTimer !== null) {
    window.clearTimeout(saveTimer);
    saveTimer = null;
  }
}

export function openAutoSaveSettings(): void {
  const current = getAutoSaveInterval();
  showModal('Auto-save settings', `<div class="mb-2">
    <label class="form-label" for="autosaveInterval">Auto-save interval (seconds)</label>
    <input type="range" id="autosaveInterval" min="5" max="300" step="5" class="form-range" value="${current}">
    <div class="d-flex justify-content-between"><span class="text-muted small">5s</span><span id="autosaveIntervalLabel" class="fw-bold">${current}s</span><span class="text-muted small">300s</span></div>
  </div>`);
  const slider = document.getElementById('autosaveInterval') as HTMLInputElement;
  const label = document.getElementById('autosaveIntervalLabel');
  const update = () => {
    const v = parseInt(slider.value, 10);
    setAutoSaveInterval(v);
    if (label) label.textContent = v + 's';
  };
  slider.addEventListener('input', update);
}

expose('saveCharacter', saveCharacter);
expose('isDirty', isDirty);
expose('isSaving', isSaving);
expose('setAutoSaveInterval', setAutoSaveInterval);
expose('openAutoSaveSettings', openAutoSaveSettings);
