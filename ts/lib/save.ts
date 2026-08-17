// Centralized auto-save module — dirty tracking, scheduler, save button state.
// Extracted from characters/sheet.ts (improve-auto-save).
import { expose } from './expose';
import { currentChar, setCurrentChar } from './state';
import { toast } from './dom';
import { api } from './api';

let sheetDirty = false;
let saving = false;
let schedulerHandle: number | null = null;
let saveTimer: number | null = null;

let serverInterval = 12;
let settingsLoaded = false;

export function getAutoSaveInterval(): number {
  return serverInterval;
}

async function loadAutoSaveInterval(): Promise<void> {
  try {
    const result = await api('GET', '/api/settings/autosave');
    const value = Number(result?.interval);
    serverInterval = Number.isInteger(value) && value >= 5 && value <= 300 ? value : 12;
  } catch {
    serverInterval = 12;
  }
  settingsLoaded = true;
  if (schedulerHandle !== null) {
    window.clearInterval(schedulerHandle);
    schedulerHandle = window.setInterval(autoSaveScheduler, getAutoSaveInterval() * 1000);
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
  if (!settingsLoaded) void loadAutoSaveInterval();
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

expose('saveCharacter', saveCharacter);
expose('isDirty', isDirty);
expose('isSaving', isSaving);
