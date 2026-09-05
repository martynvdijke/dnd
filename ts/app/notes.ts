// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, capitalize, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { currentChar } from '../lib/state';

async function renderNotes() {
  const el = document.getElementById('notesSection')!;
  if (!currentChar) return;
  try {
    const notes = await api('GET', `/api/notes?character_id=${currentChar.id}`);
    const groups: Record<string, any[]> = { general: [], backstory: [], quest: [], lore: [], dm: [], other: [] };
    notes.forEach((n: any) => { if (groups[n.category]) groups[n.category].push(n); else groups.other.push(n); });
    let html = `
      <div class="d-flex justify-content-between align-items-center">
        <h5>Notes</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddNote()"><i class="fa-solid fa-plus me-1"></i>New Note</button>
      </div>`;
    for (const [cat, items] of Object.entries(groups)) {
      if (!items.length) continue;
      html += `<h6 class="mt-3 text-muted">${capitalize(cat)}</h6>`;
      for (const n of items) {
        const visIcon = n.visibility === 'dm' ? '<i class="fa-solid fa-eye-slash ms-1 text-muted" title="DM only"></i>' : '';
        html += `<div class="card mb-2">
          <div class="card-body py-2 px-3">
            <div class="d-flex justify-content-between align-items-start">
              <div><span class="fw-bold">${esc(n.title)}</span> ${visIcon}
                <span class="badge badge-muted ms-1">${esc(n.visibility)}</span></div>
              <div class="d-flex gap-1">
                <button class="btn btn-sm btn-outline-primary" onclick="editNote(${n.id})"><i class="fa-solid fa-pen"></i></button>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteNote(${n.id})"><i class="fa-solid fa-trash"></i></button>
              </div>
            </div>
            <div class="mt-1 small text-muted" style="white-space:pre-wrap">${esc(n.content).substring(0, 300)}</div>
          </div>
        </div>`;
      }
    }
    if (!notes.length) html += '<div class="empty-state"><i class="fa-solid fa-note-sticky fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">No notes yet. Keep track of campaign information, backstory details, and DM secrets.</p></div>';
    el.innerHTML = html;
  } catch { el.innerHTML = '<div class="empty-state"><p class="small text-muted">Could not load notes.</p></div>'; }
}
expose('showAddNote', function () {
  showModal('New Note', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="noteTitle" placeholder="Note title"></div>
    <div class="mb-3"><label class="form-label">Content</label><textarea class="form-control" id="noteContent" rows="6"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Visibility</label>
        <select class="form-select" id="noteVis">
          <option value="player">Player Only</option><option value="both">Player & DM</option>
          <option value="dm">DM Only</option>
        </select></div>
      <div class="col-6"><label class="form-label">Category</label>
        <select class="form-select" id="noteCat">
          <option value="general">General</option><option value="backstory">Backstory</option>
          <option value="quest">Quest</option><option value="lore">Lore</option>
          <option value="dm">DM</option><option value="other">Other</option>
        </select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveNote()"><i class="fa-solid fa-plus me-1"></i>Create Note</button>
  `);
});
expose('saveNote', async function () {
  await api('POST', '/api/notes', {
    character_id: currentChar.id,
    title: (document.getElementById('noteTitle') as HTMLInputElement).value,
    content: (document.getElementById('noteContent') as HTMLTextAreaElement).value,
    visibility: (document.getElementById('noteVis') as HTMLSelectElement).value,
    category: (document.getElementById('noteCat') as HTMLSelectElement).value,
  });
  hideModal();
  renderNotes();
  toast('Note created');
});
expose('editNote', async function (id: number) {
  const notes = await api('GET', `/api/notes?character_id=${currentChar.id}`);
  const n = notes.find((x: any) => x.id === id);
  if (!n) return;
  showModal('Edit Note', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="noteTitle" value="${esc(n.title)}"></div>
    <div class="mb-3"><label class="form-label">Content</label><textarea class="form-control" id="noteContent" rows="6">${esc(n.content)}</textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Visibility</label>
        <select class="form-select" id="noteVis">${['player','both','dm'].map(v => `<option value="${v}"${v===n.visibility?' selected':''}>${capitalize(v)}</option>`).join('')}</select></div>
      <div class="col-6"><label class="form-label">Category</label>
        <select class="form-select" id="noteCat">${['general','backstory','quest','lore','dm','other'].map(c => `<option value="${c}"${c===n.category?' selected':''}>${capitalize(c)}</option>`).join('')}</select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEditNote(${id})"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
});
expose('saveEditNote', async function (id: number) {
  await api('PUT', `/api/notes/${id}`, {
    title: (document.getElementById('noteTitle') as HTMLInputElement).value,
    content: (document.getElementById('noteContent') as HTMLTextAreaElement).value,
    visibility: (document.getElementById('noteVis') as HTMLSelectElement).value,
    category: (document.getElementById('noteCat') as HTMLSelectElement).value,
  });
  hideModal();
  renderNotes();
  toast('Note updated');
});
expose('deleteNote', async function (id: number) {
  if (!confirm('Delete this note?')) return;
  await api('DELETE', `/api/notes/${id}`);
  renderNotes();
  toast('Note deleted');
});
expose('renderNotes', renderNotes);
