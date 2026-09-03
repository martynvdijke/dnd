// @ts-nocheck — extracted from app.ts monolith (address-tech-debt-and-ux)
import { expose } from '../lib/expose';
import { currentChar, setCurrentChar } from '../lib/state';
import { esc, attrEscape, showModal } from '../lib/dom';
import { api } from '../lib/api';

export function renderSpells() {
  const spells = currentChar.spells || [];
  const sc = currentChar.spellcasting || {};
  document.getElementById('spellsSection')!.innerHTML = sc.ability ? `
    <h5>Spellcasting</h5>
    <div class="row g-3 mb-3">
      <div class="col-md-4"><label class="form-label">Ability</label><input class="form-control form-control-sm" value="${esc(sc.ability)}" onchange="updateSpellcasting('ability',this.value)"></div>
      <div class="col-md-4"><label class="form-label">Save DC</label><input class="form-control form-control-sm" type="number" value="${sc.save_dc||0}" onchange="updateSpellcasting('save_dc',+this.value)"></div>
      <div class="col-md-4"><label class="form-label">Atk Bonus</label><input class="form-control form-control-sm" type="number" value="${sc.attack_bonus||0}" onchange="updateSpellcasting('attack_bonus',+this.value)"></div>
    </div>
    <h6>Spell Slots</h6>
    <div class="d-flex gap-3 flex-wrap mb-3">
      ${[1,2,3,4,5,6,7,8,9].map(lv => {
        const mx = sc[`slots_${lv}_max`] || 0;
        if (!mx) return '';
        return `<div class="text-center">
          <div class="text-muted small">Lv ${lv}</div>
          <input type="number" class="form-control form-control-sm text-center" style="width:55px" id="slotUse${lv}" value="${sc[`slots_${lv}_used`]||0}" onchange="updateSpellSlot(${lv})" min="0" max="${mx}">
          <div class="text-muted small">/ ${mx}</div>
        </div>`;
      }).join('')}
    </div>
    <div class="d-flex justify-content-between align-items-center mt-3">
      <h6>Known Spells <span class="text-muted small fw-normal">${spells.filter((s:any)=>s.prepared).length}/${spells.filter((s:any)=>s.level>0 && spells.filter((ss:any)=>ss.level===s.level).length > 0).length + 3} prepared</span></h6>
      <div class="d-flex gap-2">
        <button class="btn btn-sm btn-outline-gold" onclick="showPrepareSpells()"><i class="fa-solid fa-book-open me-1"></i>Prepare Spells</button>
        <button class="btn btn-primary btn-sm" onclick="addSpell()"><i class="fa-solid fa-plus me-1"></i>Add Spell</button>
      </div>
    </div>
    <div class="row g-2 mt-2">
      ${spells.map((s:any) => `
        <div class="col-md-6">
          <div class="card spell-card ${s.prepared ? 'border-gold' : ''}">
            <div class="card-body py-2 px-3">
              <div class="d-flex justify-content-between align-items-start">
                <div>
                  <span class="fw-bold">${esc(s.name)}</span>
                  <span class="badge ${s.level === 0 ? 'badge-muted' : 'badge-blood'} ms-1">${s.level > 0 ? 'Lv' + s.level : 'Cantrip'}</span>
                  <span class="badge badge-gold ms-1">${esc(s.school)}</span>
                </div>
                <div class="d-flex gap-1">
                  <button class="btn btn-sm btn-outline-primary" onclick="editSpell(${s.id},'${attrEscape(s.name)}',${s.level},'${attrEscape(s.school)}',${s.prepared},'${attrEscape(s.components||'')}','${attrEscape(s.range||'')}','${attrEscape(s.casting_time||'')}','${attrEscape(s.duration||'')}','${attrEscape(s.description||'')}')"><i class="fa-solid fa-pen"></i></button>
                  <button class="btn btn-sm btn-outline-danger" onclick="deleteSpell(${s.id})"><i class="fa-solid fa-trash"></i></button>
                </div>
              </div>
              <div class="small text-muted mt-1">
                ${s.casting_time ? `<span class="me-2"><i class="fa-regular fa-clock me-1"></i>${esc(s.casting_time)}</span>` : ''}
                ${s.range ? `<span class="me-2"><i class="fa-solid fa-bullseye me-1"></i>${esc(s.range)}</span>` : ''}
                ${s.duration ? `<span><i class="fa-regular fa-hourglass me-1"></i>${esc(s.duration)}</span>` : ''}
              </div>
              ${s.description ? `<p class="mb-0 mt-1 small text-muted">${esc(s.description).substring(0, 150)}${s.description.length > 150 ? '...' : ''}</p>` : ''}
            </div>
          </div>
        </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-wand-sparkles fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Spells Known</p><p class="small text-muted">Add spells to your spellbook using the button above.</p></div>'}
    </div>` : `
    <div class="empty-state"><i class="fa-solid fa-wand-sparkles fa-2x mb-2 d-block text-muted"></i>
    <p class="text-muted fst-italic">No spellcasting.</p>
    <button class="btn btn-outline-primary btn-sm" onclick="enableSpellcasting()"><i class="fa-solid fa-magic me-1"></i>Set Up Spellcasting</button></div>`;
}

export async function updateSpellcasting(field:string, value:any) {
  if (!currentChar) return;
  const sc = currentChar.spellcasting || {};
  sc[field] = value;
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
}

export async function updateSpellSlot(level:number) {
  if (!currentChar) return;
  const sc = currentChar.spellcasting || {};
  sc[`slots_${level}_used`] = +(document.getElementById(`slotUse${level}`) as HTMLInputElement).value || 0;
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
}

export function showPrepareSpells() {
  const spells = currentChar.spells || [];
  const sc = currentChar.spellcasting || {};
  const maxPrepared = currentChar.level > 0 ? (currentChar.class_mod || 0) + currentChar.level : 0; // WIS/INT/CHA mod + level
  const currentPrepared = spells.filter((s:any) => s.prepared).length;

  // Group by level
  const byLevel: Record<number, any[]> = {};
  spells.forEach((s:any) => {
    const lv = s.level || 0;
    if (!byLevel[lv]) byLevel[lv] = [];
    byLevel[lv].push(s);
  });

  const bodyHtml = `
    <div class="mb-2 d-flex justify-content-between">
      <span class="fw-bold">Prepared: ${currentPrepared} / ${maxPrepared}</span>
      <span class="text-muted small">Max = spellcasting mod + level</span>
    </div>
    <div class="spell-prep-list">
      ${Object.keys(byLevel).sort((a,b)=>+a - +b).map(lv => `
        <div class="spell-prep-group">
          <h6>${lv === '0' ? 'Cantrips' : 'Level ' + lv}</h6>
          ${byLevel[+lv].map((s:any) => `
            <div class="spell-prep-item">
              <input type="checkbox" class="form-check-input" id="prep-${s.id}" ${s.prepared ? 'checked' : ''}>
              <label for="prep-${s.id}">${esc(s.name)} <span class="text-muted">(${esc(s.school)})</span></label>
            </div>
          `).join('')}
        </div>
      `).join('')}
    </div>
    <button class="btn btn-gold w-100 mt-3" onclick="saveSpellPrep()"><i class="fa-solid fa-book-open me-1"></i>Save Preparation</button>
  `;
  showModal('Prepare Spells', bodyHtml);
}

expose('renderSpells', renderSpells);
expose('updateSpellcasting', updateSpellcasting);
expose('updateSpellSlot', updateSpellSlot);
expose('showPrepareSpells', showPrepareSpells);
