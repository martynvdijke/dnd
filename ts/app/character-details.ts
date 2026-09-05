// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, capitalize, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { currentChar } from '../lib/state';
import { refreshChar } from '../lib/refresh';
import { renderError } from '../lib/errors';
import { renderSheet, updateField } from '../characters/sheet';

function renderFeatures() {
  const feats = currentChar.features || [];
  document.getElementById('featuresSection')!.innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5>Features & Proficiencies</h5>
      <button class="btn btn-primary btn-sm" onclick="addFeature()"><i class="fa-solid fa-plus me-1"></i>Add Feature</button>
    </div>
    <div class="mt-2">
      ${feats.map((f:any) => `
        <div class="card mb-2">
          <div class="card-body py-2 px-3">
            <div class="d-flex justify-content-between align-items-center">
              <div><span class="fw-bold">${esc(f.name)}</span>
                <span class="badge badge-blood ms-1">Lv ${f.level_gained}</span>
                ${f.source ? `<span class="badge badge-gold ms-1">${esc(f.source)}</span>` : ''}
                <p class="mb-0 mt-1 small text-muted">${esc(f.description)}</p></div>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteFeature(${f.id})"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>
        </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-star fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Features Yet</p><p class="small text-muted">Track class, race, and feat features here.</p></div>'}
    </div>`;
}
expose('addFeature', function () {
  showModal('Add Feature', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="featName"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="featDesc" rows="3"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Source</label><input class="form-control" id="featSource" placeholder="Class, Race, etc."></div>
      <div class="col-6"><label class="form-label">Level Gained</label><input class="form-control" id="featLevel" type="number" value="1"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveFeature(this)">Add Feature</button>
  `);
});
expose('saveFeature', async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/features`, {
    name: (document.getElementById('featName') as HTMLInputElement).value,
    description: (document.getElementById('featDesc') as HTMLTextAreaElement).value,
    source: (document.getElementById('featSource') as HTMLInputElement).value,
    level_gained: +(document.getElementById('featLevel') as HTMLInputElement).value || 1,
  });
  hideModal();
  await refreshChar();
  renderFeatures();
  toast('Feature added');
});
expose('deleteFeature', async function (id:number) {
  await api('DELETE', `/api/features/${id}`);
  await refreshChar();
  renderFeatures();
  toast('Feature removed');
});

expose('addProf', function () {
  showModal('Add Proficiency', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="profName"></div>
    <div class="mb-3"><label class="form-label">Type</label>
      <select class="form-select" id="profType">
        <option value="skill">Skill</option><option value="save">Saving Throw</option><option value="tool">Tool</option>
        <option value="weapon">Weapon</option><option value="armor">Armor</option><option value="language">Language</option>
        <option value="other">Other</option>
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveProf(this)">Add Proficiency</button>
  `);
});
expose('saveProf', async function (btn:HTMLElement) {
  await api('POST', '/api/proficiencies', {
    character_id: currentChar.id,
    type: (document.getElementById('profType') as HTMLSelectElement).value,
    name: (document.getElementById('profName') as HTMLInputElement).value,
  });
  hideModal();
  await refreshChar();
  renderSheet();
  toast('Proficiency added');
});
expose('deleteProf', async function (id:number) {
  await api('DELETE', `/api/proficiencies/${id}`);
  await refreshChar();
  renderSheet();
  toast('Proficiency removed');
});

function renderDetails() {
  const c = currentChar;
  const el = document.getElementById('detailsSection')!;
  el.innerHTML = `
    <div class="row g-3">
      <div class="col-md-12 mb-2">
        <label class="form-label">Portrait</label>
        <div class="d-flex align-items-center gap-2">
          ${c.portrait_url ? `<img src="${esc(c.portrait_url)}" class="character-portrait-lg me-2" alt="">` : ''}
          <input type="file" class="form-control form-control-sm" id="portraitUpload" accept="image/*">
          <button class="btn btn-primary btn-sm" onclick="uploadPortrait()"><i class="fa-solid fa-upload me-1"></i>Upload</button>
          <button class="btn btn-outline-info btn-sm" onclick="browsePortrait()"><i class="fa-solid fa-image me-1"></i>Browse</button>
          ${c.portrait_url ? `<button class="btn btn-outline-danger btn-sm" onclick="clearPortrait()"><i class="fa-solid fa-xmark"></i></button>` : ''}
        </div>
      </div>
    </div>
    <div class="row g-3">
      <div class="col-md-4">
        <label class="form-label">Race ${c.compendium_race_id ? '<span class="badge badge-compendium" title="Linked from compendium"><i class="fa-solid fa-book me-1"></i></span>' : ''}</label>
        <div class="input-group input-group-sm">
          <input class="form-control form-control-sm" value="${esc(c.race)}" oninput="autoSaveField('race',this)">
          <button class="btn btn-outline-gold" onclick="linkCharIdentity('race')" title="Link race from compendium"><i class="fa-solid fa-book"></i></button>
          ${c.compendium_race_id ? `<button class="btn btn-outline-secondary" onclick="unlinkCharIdentity('race')" title="Unlink race from compendium"><i class="fa-solid fa-link-slash"></i></button>` : ''}
        </div>
      </div>
      <div class="col-md-4">
        <label class="form-label">Class ${c.compendium_class_id ? '<span class="badge badge-compendium" title="Linked from compendium"><i class="fa-solid fa-book me-1"></i></span>' : ''}</label>
        <div class="input-group input-group-sm">
          <input class="form-control form-control-sm" value="${esc(c.class)}" oninput="autoSaveField('class',this)">
          <button class="btn btn-outline-gold" onclick="linkCharIdentity('class')" title="Link class from compendium"><i class="fa-solid fa-book"></i></button>
          ${c.compendium_class_id ? `<button class="btn btn-outline-secondary" onclick="unlinkCharIdentity('class')" title="Unlink class from compendium"><i class="fa-solid fa-link-slash"></i></button>` : ''}
        </div>
      </div>
      <div class="col-md-4"><label class="form-label">Subclass</label><input class="form-control form-control-sm" value="${esc(c.subclass)}" oninput="autoSaveField('subclass',this)"></div>
    </div>
    <div class="row g-3 mt-1">
      <div class="col-md-4"><label class="form-label">Level</label><input class="form-control form-control-sm" type="number" value="${c.level}" oninput="autoSaveField('level',this)"></div>
      <div class="col-md-4">
        <label class="form-label">Background ${c.compendium_background_id ? '<span class="badge badge-compendium" title="Linked from compendium"><i class="fa-solid fa-book me-1"></i></span>' : ''}</label>
        <div class="input-group input-group-sm">
          <input class="form-control form-control-sm" value="${esc(c.background)}" oninput="autoSaveField('background',this)">
          <button class="btn btn-outline-gold" onclick="linkCharIdentity('background')" title="Link background from compendium"><i class="fa-solid fa-book"></i></button>
          ${c.compendium_background_id ? `<button class="btn btn-outline-secondary" onclick="unlinkCharIdentity('background')" title="Unlink background from compendium"><i class="fa-solid fa-link-slash"></i></button>` : ''}
        </div>
      </div>
      <div class="col-md-4"><label class="form-label">Alignment</label><input class="form-control form-control-sm" value="${esc(c.alignment)}" oninput="autoSaveField('alignment',this)"></div>
    </div>
    <div class="mt-2 form-check">
      <input type="checkbox" class="form-check-input" id="hpAutoCalcCb" ${c.hp_auto_calc ? 'checked' : ''} onchange="autoSaveField('hp_auto_calc',this.checked)">
      <label class="form-check-label small" for="hpAutoCalcCb">Auto-calculate HP from classes</label>
      <button class="btn btn-sm btn-outline-gold ms-2" onclick="calcHP()"><i class="fa-solid fa-calculator me-1"></i>Recalculate HP</button>
    </div>
    <h5 class="mt-3">Multi-Class</h5>
    <div id="multiClassArea">
      ${(c.classes && c.classes.length ? c.classes.map((cc: any) => `
        <div class="inv-item">
          <span class="fw-bold">${esc(cc.class)}</span>
          ${cc.subclass ? `<span class="text-muted small">(${esc(cc.subclass)})</span>` : ''}
          <span class="badge badge-blood ms-1">Lv ${cc.level}</span>
          <span class="badge badge-muted ms-1">${esc(cc.hit_dice)}</span>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-outline-primary" onclick="editClass(${cc.id})"><i class="fa-solid fa-pen"></i></button>
            <button class="btn btn-sm btn-outline-danger" onclick="deleteClass(${cc.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </div>`).join('')
        : '<div class="text-muted small">Single class. Add multi-class entries below.</div>')}
    </div>
    <button class="btn btn-sm btn-outline-primary mt-1" onclick="addClass()"><i class="fa-solid fa-plus me-1"></i>Add Class</button>
    <hr class="my-3">
    ${['personality_traits','ideals','bonds','flaws','appearance'].map(f => `
      <div class="mb-3"><label class="form-label">${capitalize(f.replace(/_/g,' '))}</label>
      <textarea class="form-control form-control-sm" rows="2" oninput="autoSaveField('${f}',this)">${esc((c as any)[f])}</textarea></div>
    `).join('')}
    <div class="mb-3"><label class="form-label">Backstory</label>
    <textarea class="form-control form-control-sm" rows="4" oninput="autoSaveField('backstory',this)">${esc(c.backstory)}</textarea></div>
    <div class="mt-3">
      <button class="btn btn-outline-primary btn-sm" onclick="shareCharacter()"><i class="fa-solid fa-share-nodes me-1"></i>Share Character</button>
    </div>`;
}
expose('renderDetails', renderDetails);

// Race Colors
expose('showRaceColors', async function () {
  try {
    const data = await api('GET', '/api/race-colors');
    const colors = data.colors || {};
    showModal('Race Colors', `
      <p class="small text-muted">Set colors for character races. These appear as colored badges on the campaign overview and character lists.</p>
      <div id="raceColorsList">
        ${Object.entries(colors).map(([race, color]) => `
          <div class="row g-2 mb-2 align-items-center">
            <div class="col-4"><label class="form-label mb-0 small">${esc(race)}</label></div>
            <div class="col-2"><input type="color" class="form-control form-control-color" value="${esc(color as string)}" data-race="${esc(race)}"></div>
            <div class="col-6"><input class="form-control form-control-sm" value="${esc(color as string)}" data-race="${esc(race)}" oninput="this.previousElementSibling.value=this.value"><span class="badge ms-1" style="background:${esc(color as string)};color:#fff">Preview</span></div>
          </div>
        `).join('')}
      </div>
      <div class="mt-3">
        <div class="row g-2 align-items-center">
          <div class="col-4"><input class="form-control form-control-sm" id="newRaceColorName" placeholder="New race"></div>
          <div class="col-2"><input type="color" class="form-control form-control-color" id="newRaceColorPicker" value="#6c757d"></div>
          <div class="col-3"><input class="form-control form-control-sm" id="newRaceColorValue" value="#6c757d" oninput="document.getElementById('newRaceColorPicker').value=this.value"></div>
          <div class="col-3"><button class="btn btn-sm btn-outline-primary w-100" onclick="addRaceColor()"><i class="fa-solid fa-plus me-1"></i>Add</button></div>
        </div>
      </div>
      <button class="btn btn-primary w-100 mt-3" onclick="saveRaceColors()"><i class="fa-solid fa-save me-1"></i>Save Changes</button>
    `);
  } catch (e: any) { renderError(e); }
});
expose('addRaceColor', function () {
  const name = (document.getElementById('newRaceColorName') as HTMLInputElement).value.trim();
  const color = (document.getElementById('newRaceColorPicker') as HTMLInputElement).value;
  if (!name) { toast('Enter a race name', true); return; }
  const list = document.getElementById('raceColorsList')!;
  list.insertAdjacentHTML('beforeend', `
    <div class="row g-2 mb-2 align-items-center">
      <div class="col-4"><label class="form-label mb-0 small">${esc(name)}</label></div>
      <div class="col-2"><input type="color" class="form-control form-control-color" value="${color}" data-race="${esc(name)}"></div>
      <div class="col-6"><input class="form-control form-control-sm" value="${color}" data-race="${esc(name)}" oninput="this.previousElementSibling.value=this.value"></div>
    </div>
  `);
  (document.getElementById('newRaceColorName') as HTMLInputElement).value = '';
});
expose('saveRaceColors', async function () {
  const colors: Record<string, string> = {};
  document.querySelectorAll('#raceColorsList input[type="color"]').forEach((el) => {
    const input = el as HTMLInputElement;
    const race = input.getAttribute('data-race');
    if (race) colors[race] = input.value;
  });
  await api('PUT', '/api/race-colors', { colors });
  hideModal();
  toast('Race colors saved');
});

// Multi-Class
expose('addClass', function () {
  showModal('Add Class', `
    <div class="mb-3"><label class="form-label">Class</label><input class="form-control" id="mcClass"></div>
    <div class="mb-3"><label class="form-label">Subclass</label><input class="form-control" id="mcSubclass"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="mcLevel" type="number" value="1" min="1"></div>
      <div class="col-6"><label class="form-label">Hit Dice</label>
        <select class="form-select" id="mcHD"><option value="d6">d6</option><option value="d8">d8</option><option value="d10" selected>d10</option><option value="d12">d12</option></select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveClass()"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
});
expose('saveClass', async function () {
  await api('POST', `/api/characters/${currentChar.id}/classes`, {
    class: (document.getElementById('mcClass') as HTMLInputElement).value,
    subclass: (document.getElementById('mcSubclass') as HTMLInputElement).value,
    level: +(document.getElementById('mcLevel') as HTMLInputElement).value || 1,
    hit_dice: (document.getElementById('mcHD') as HTMLSelectElement).value,
  });
  hideModal();
  await refreshChar();
  renderSheet();
  toast('Class added');
});
expose('editClass', function (id: number) {
  const cc = currentChar.classes.find((c: any) => c.id === id);
  if (!cc) return;
  showModal('Edit Class', `
    <div class="mb-3"><label class="form-label">Class</label><input class="form-control" id="mcClass" value="${esc(cc.class)}"></div>
    <div class="mb-3"><label class="form-label">Subclass</label><input class="form-control" id="mcSubclass" value="${esc(cc.subclass)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="mcLevel" type="number" value="${cc.level}" min="1"></div>
      <div class="col-6"><label class="form-label">Hit Dice</label>
        <select class="form-select" id="mcHD">${['d6','d8','d10','d12'].map(d => `<option value="${d}"${d===cc.hit_dice?' selected':''}>${d}</option>`).join('')}</select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEditClass(${id})"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
});
expose('saveEditClass', async function (id: number) {
  await api('PUT', `/api/classes/${id}`, {
    class: (document.getElementById('mcClass') as HTMLInputElement).value,
    subclass: (document.getElementById('mcSubclass') as HTMLInputElement).value,
    level: +(document.getElementById('mcLevel') as HTMLInputElement).value || 1,
    hit_dice: (document.getElementById('mcHD') as HTMLSelectElement).value,
  });
  hideModal();
  await refreshChar();
  renderSheet();
  toast('Class updated');
});
expose('deleteClass', async function (id: number) {
  if (!confirm('Remove this class?')) return;
  await api('DELETE', `/api/classes/${id}`);
  await refreshChar();
  renderSheet();
  toast('Class removed');
});

export { renderFeatures, renderDetails };
