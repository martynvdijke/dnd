// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api, getCsrfToken, getApiToken, clearApiToken } from '../lib/api';
import { currentChar, currentUser, setCurrentChar } from '../lib/state';
import { refreshChar } from '../lib/refresh';
import { renderError } from '../lib/errors';
import { renderSheet } from '../characters/sheet';
import { updateField } from '../characters/sheet';
import { showView } from '../navigation';
import { loadCharacters } from '../characters/list';
import { FilePicker } from '../file-picker';

expose('newChar', function () {
  showModal('New Character', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newName" placeholder="Character name"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="newRace" list="raceSuggestions"><datalist id="raceSuggestions"></datalist></div>
      <div class="col-6"><label class="form-label">Class</label><input class="form-control" id="newClass" list="classSuggestions"><datalist id="classSuggestions"></datalist></div>
    </div>
    <button class="btn btn-primary w-100" onclick="createChar()"><i class="fa-solid fa-plus me-1"></i>Create</button>
    <div class="text-center mt-2"><button class="btn btn-sm btn-outline-gold" onclick="generateRandomChar()"><i class="fa-solid fa-dice me-1"></i>Random Character</button></div>
  `);
  fetch('/api/compendium/races', { credentials: 'include' }).then(r => r.json()).then((races:any[]) => {
    document.getElementById('raceSuggestions')!.innerHTML = races.map((r:any) => `<option value="${esc(r.name)}">`).join('');
  }).catch(() => {});
  fetch('/api/compendium/classes', { credentials: 'include' }).then(r => r.json()).then((cls:any[]) => {
    document.getElementById('classSuggestions')!.innerHTML = cls.map((c:any) => `<option value="${esc(c.name)}">`).join('');
  }).catch(() => {});
});

expose('createChar', async function () {
  const name = (document.getElementById('newName') as HTMLInputElement).value || 'Unnamed';
  const race = (document.getElementById('newRace') as HTMLInputElement).value;
  const cls = (document.getElementById('newClass') as HTMLInputElement).value;
  try {
    const char = await api('POST', '/api/characters', { name, race, class: cls });
    hideModal();
    if (char.id) await (window as any).openChar(char.id);
    loadCharacters();
  } catch (e:any) {
    renderError(e);
  }
});

expose('showImport', function () {
  showModal('Import Character', `
    <p class="text-muted fst-italic small mb-3">Paste JSON or upload a file</p>
    <div class="mb-3"><label class="form-label">JSON</label><textarea class="form-control" id="importJson" rows="6" style="font-family:monospace;font-size:0.8rem"></textarea></div>
    <div class="mb-3"><label class="form-label">File</label><input class="form-control" type="file" id="importFile" accept=".json"></div>
    <button class="btn btn-primary w-100" onclick="doImport()"><i class="fa-solid fa-file-import me-1"></i>Import</button>
  `);
});

expose('doImport', async function () {
  const jsonEl = document.getElementById('importJson') as HTMLTextAreaElement;
  const fileEl = document.getElementById('importFile') as HTMLInputElement;
  try {
    let result;
    if (fileEl.files && fileEl.files[0]) {
      const form = new FormData();
      form.append('file', fileEl.files[0]);
      const res = await fetch('/api/characters/import', { method: 'POST', headers: { 'X-CSRF-Token': getCsrfToken(), ...(getApiToken() ? { 'Authorization': `Bearer ${getApiToken()}` } : {}) }, credentials: 'include', body: form });
      result = await res.json();
    } else if (jsonEl.value.trim()) {
      result = await api('POST', '/api/characters/import', JSON.parse(jsonEl.value));
    } else {
      toast('Provide JSON or a file', true);
      return;
    }
    toast(`Imported ${Array.isArray(result) ? result.length : 1} character(s)`);
    hideModal();
    loadCharacters();
  } catch (e:any) {
    toast('Import failed: ' + e.message, true);
  }
});

expose('exportChar', async function () {
  if (!currentChar) return;
  try {
    const data = await api('GET', `/api/characters/${currentChar.id}/export`);
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const a = document.createElement('a');
    const url = URL.createObjectURL(blob);
    a.href = url;
    a.download = currentChar.name.replace(/[^a-zA-Z0-9]/g, '_') + '.json';
    a.click();
    URL.revokeObjectURL(url);
  } catch (e:any) {
    renderError(e);
  }
});

expose('printChar', async function () {
  if (!currentChar) return;
  try {
    const res = await fetch(`/api/characters/${currentChar.id}/print`, {
      headers: { 'X-CSRF-Token': getCsrfToken() }, credentials: 'include',
    });
    const text = await res.text();
    const win = window.open('', '_blank');
    if (win) {
      win.document.write(`<pre style="font-family:monospace;font-size:12px;line-height:1.4">${esc(text)}</pre>`);
      win.document.close();
      win.print();
    }
  } catch (e:any) {
    renderError(e);
  }
});

expose('deleteChar', async function () {
  if (!currentChar) return;
  if (!confirm('Delete this character?')) return;
  try {
    await api('DELETE', `/api/characters/${currentChar.id}`);
    setCurrentChar(null);
    showView('characters');
    await loadCharacters();
    toast('Character deleted');
  } catch (e: any) {
    renderError(e);
  }
});

expose('logout', async function () {
  await api('POST', '/api/logout');
  clearApiToken();
  if (currentUser?.username) localStorage.removeItem(`villum-api-token-${currentUser.username}`);
  (window as any).clearSelection?.();
  window.location.href = '/login';
});

expose('uploadPortrait', async function () {
  const input = document.getElementById('portraitUpload') as HTMLInputElement;
  if (!input.files || !input.files[0]) { toast('Select an image', true); return; }
  const form = new FormData();
  form.append('image', input.files[0]);
  try {
    const res = await fetch('/api/upload', {
      method: 'POST', headers: { 'X-CSRF-Token': getCsrfToken(), ...(getApiToken() ? { 'Authorization': `Bearer ${getApiToken()}` } : {}) }, credentials: 'include', body: form,
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    await updateField('portrait_url', data.url);
    await refreshChar();
    renderSheet();
    toast('Portrait uploaded');
  } catch (e: any) { renderError(e); }
});

expose('browsePortrait', async function () {
  try {
    const url = await FilePicker.pick();
    await updateField('portrait_url', url);
    await refreshChar();
    renderSheet();
    toast('Portrait set');
  } catch (e: any) { renderError(e); }
});

expose('clearPortrait', async function () {
  await updateField('portrait_url', '');
  await refreshChar();
  renderSheet();
  toast('Portrait removed');
});
