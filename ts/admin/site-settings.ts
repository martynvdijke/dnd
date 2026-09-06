// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { api } from './state';
import { toast } from '../lib/dom';

async function loadEinkSetting() {
  try {
    const res = await api('GET', '/api/admin/settings/eink');
    const el = document.getElementById('einkEnabled') as HTMLInputElement | null;
    if (el) el.checked = !!res.enabled;
  } catch {
    toast('Failed to load e-ink setting', true);
  }
}
async function saveEinkSetting() {
  const el = document.getElementById('einkEnabled') as HTMLInputElement | null;
  const enabled = !!el && el.checked;
  try {
    await api('PUT', '/api/admin/settings/eink', { enabled });
    toast(enabled ? 'E-ink mode enabled site-wide' : 'E-ink mode disabled');
  } catch {
    toast('Failed to save e-ink setting', true);
  }
}
expose('loadEinkSetting', loadEinkSetting);
expose('saveEinkSetting', saveEinkSetting);

async function loadAISetting() {
  try {
    const res = await api('GET', '/api/admin/settings/ai');
    const el = document.getElementById('aiEnabled') as HTMLInputElement | null;
    if (el) el.checked = !!res.enabled;
  } catch {
    toast('Failed to load AI setting', true);
  }
}
async function saveAISetting() {
  const el = document.getElementById('aiEnabled') as HTMLInputElement | null;
  const enabled = !!el && el.checked;
  try {
    await api('PUT', '/api/admin/settings/ai', { enabled });
    document.body.classList.toggle('ai-disabled', !enabled);
    toast(enabled ? 'AI features enabled site-wide' : 'AI features disabled site-wide');
  } catch {
    toast('Failed to save AI setting', true);
  }
}
expose('loadAISetting', loadAISetting);
expose('saveAISetting', saveAISetting);

async function loadAutoSaveSetting() {
  try {
    const res = await api('GET', '/api/admin/settings/autosave');
    const el = document.getElementById('autosaveInterval') as HTMLInputElement | null;
    if (el) el.value = String(res.interval ?? 12);
  } catch {
    toast('Failed to load auto-save setting', true);
  }
}
async function saveAutoSaveSetting() {
  const el = document.getElementById('autosaveInterval') as HTMLInputElement | null;
  const interval = Number(el?.value || 12);
  try {
    const res = await api('PUT', '/api/admin/settings/autosave', { interval });
    if (el) el.value = String(res.interval);
    toast('Auto-save setting saved site-wide');
  } catch {
    toast('Failed to save auto-save setting', true);
  }
}
expose('loadAutoSaveSetting', loadAutoSaveSetting);
expose('saveAutoSaveSetting', saveAutoSaveSetting);
