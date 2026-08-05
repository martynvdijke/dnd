// Extracted from app.ts — NPCs section (address-tech-debt-and-ux)
import { expose } from '../lib/expose';
import { currentChar, allNPCs, setAllNPCs } from '../lib/state';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api, getCsrfToken } from '../lib/api';
import { FilePicker } from '../file-picker';

// ─── NPCs ───

async function renderNPCs() {
  const el = document.getElementById('npcsSection')!;
  try {
    const links = await api('GET', `/api/characters/${currentChar.id}/npcs`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Related NPCs</h5>
        <button class="btn btn-primary btn-sm" onclick="showLinkNPC()"><i class="fa-solid fa-link me-1"></i>Link NPC</button>
      </div>
      <div class="mt-2">${links.length ? links.map((n:any) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(n.npc_name)}</span>
            ${n.npc_race_color ? `<span class="badge ms-1" style="background:${n.npc_race_color};color:#fff">${esc(n.npc_race)}</span>` : `<span class="text-muted small">${esc(n.npc_race)}</span>`}
            <span class="text-muted small">${esc(n.npc_class)}</span>
            ${!n.npc_is_alive ? '<span class="badge badge-blood ms-1">Deceased</span>' : ''}</div>
          <div>
            <span class="badge badge-gold">${esc(n.relationship)}</span>
            ${n.interaction_count > 0 ? `<span class="badge badge-blood ms-1">${n.interaction_count} talks</span>` : ''}
            <button class="btn btn-sm btn-outline-primary" onclick="logNPCInteraction(${n.id})"><i class="fa-solid fa-comment"></i></button>
            <button class="btn btn-sm btn-outline-danger" onclick="unlinkNPC(${n.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </div>`).join('')
        : '<div class="empty-state"><i class="fa-solid fa-user-group fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No NPCs Linked</p><p class="small text-muted">Link NPCs to track relationships and interactions.</p></div>'}</div>
      <hr class="my-3">
      <div class="d-flex justify-content-between align-items-center"><h5>All NPCs</h5>
        <button class="btn btn-outline-primary btn-sm" onclick="showCreateNPC()"><i class="fa-solid fa-plus me-1"></i>New NPC</button>
      </div>
      <div class="mt-2">${allNPCs.map((n:any) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(n.name)}</span>
            ${n.race_color ? `<span class="badge ms-1" style="background:${n.race_color};color:#fff">${esc(n.race)}</span>` : `<span class="text-muted small">${esc(n.race)}</span>`}
            <span class="text-muted small">${esc(n.class)}</span></div>
          <div class="text-muted small">HP: ${n.hp_current}/${n.hp_max}</div>
        </div>`).join('')}&nbsp;</div>`;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load NPCs. Try again later.</p></div>'; }
}
expose('renderNPCs', renderNPCs);

expose('showLinkNPC', function () {
  showModal('Link NPC', `
    <div class="mb-3"><label class="form-label">NPC</label>
      <select class="form-select" id="linkNPCId">${allNPCs.map((n:any) => `<option value="${n.id}">${esc(n.name)} (${esc(n.race)} ${esc(n.class)})</option>`).join('')}</select></div>
    <div class="mb-3"><label class="form-label">Relationship</label>
      <select class="form-select" id="linkNPCRel">
        <option value="ally">Ally</option><option value="enemy">Enemy</option><option value="family">Family</option>
        <option value="contact">Contact</option><option value="acquaintance">Acquaintance</option>
        <option value="pet">Pet/Mount</option><option value="deity">Deity/Patron</option><option value="other">Other</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="linkNPCNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveLinkNPC()"><i class="fa-solid fa-link me-1"></i>Link</button>
  `);
});

expose('saveLinkNPC', async function () {
  await api('POST', `/api/characters/${currentChar.id}/npcs`, {
    npc_id: +(document.getElementById('linkNPCId') as HTMLSelectElement).value,
    relationship: (document.getElementById('linkNPCRel') as HTMLSelectElement).value,
    notes: (document.getElementById('linkNPCNotes') as HTMLTextAreaElement).value,
  });
  await renderNPCs();
  hideModal();
  toast('NPC linked');
});

expose('logNPCInteraction', async function (id:number) {
  await api('POST', `/api/npcs/link/${id}/interact`, {});
  await renderNPCs();
  toast('Interaction logged');
});

expose('unlinkNPC', async function (id:number) {
  await api('DELETE', `/api/npcs/link/${id}`);
  await renderNPCs();
});

expose('showCreateNPC', function () {
  showModal('New NPC', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newNPCName"></div>
    <div class="mb-3">
      <label class="form-label">Portrait</label>
      <input type="hidden" id="newNPCPortraitUrl">
      <div class="d-flex align-items-center gap-2">
        <input type="file" class="form-control form-control-sm" id="newNPCPortraitUpload" accept="image/*">
        <button class="btn btn-primary btn-sm" onclick="uploadNewNPCPortrait()"><i class="fa-solid fa-upload me-1"></i>Upload</button>
        <button class="btn btn-outline-info btn-sm" onclick="browseNewNPCPortrait()"><i class="fa-solid fa-image me-1"></i>Browse</button>
      </div>
      <button class="btn btn-outline-primary btn-sm mt-2 ai-generate-btn" type="button" data-ai-mode="image" data-ai-target="newNPCPortraitUrl" data-ai-hint="Fantasy portrait of this NPC" data-ai-title="Generate NPC Portrait"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate with AI</button>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="newNPCRace"></div>
      <div class="col-6"><label class="form-label">Class</label><input class="form-control" id="newNPCClass"></div>
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newNPCDesc" rows="3"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveNewNPC()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
});

let newNPCPortraitUrl = '';

expose('uploadNewNPCPortrait', async function () {
  const input = document.getElementById('newNPCPortraitUpload') as HTMLInputElement;
  if (!input.files || !input.files[0]) { toast('Select an image', true); return; }
  const form = new FormData();
  form.append('image', input.files[0]);
  try {
    const res = await fetch('/api/upload', {
      method: 'POST', headers: { 'X-CSRF-Token': getCsrfToken() }, credentials: 'include', body: form,
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    newNPCPortraitUrl = data.url;
    toast('Image uploaded');
  } catch (e: any) { toast(e.message, true); }
});

expose('browseNewNPCPortrait', async function () {
  try {
    newNPCPortraitUrl = await FilePicker.pick();
    toast('Image selected');
  } catch (e: any) { toast(e.message, true); }
});

expose('saveNewNPC', async function () {
  const aiUrl = (document.getElementById('newNPCPortraitUrl') as HTMLInputElement)?.value || '';
  await api('POST', '/api/npcs', {
    name: (document.getElementById('newNPCName') as HTMLInputElement).value,
    race: (document.getElementById('newNPCRace') as HTMLInputElement).value,
    class: (document.getElementById('newNPCClass') as HTMLInputElement).value,
    description: (document.getElementById('newNPCDesc') as HTMLTextAreaElement).value,
    portrait_url: newNPCPortraitUrl || aiUrl,
  });
  newNPCPortraitUrl = '';
  setAllNPCs(await api('GET', '/api/npcs'));
  await renderNPCs();
  hideModal();
  toast('NPC created');
});
