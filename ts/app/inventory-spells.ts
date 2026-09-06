// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, capitalize, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { currentChar, setCurrentChar } from '../lib/state';
import { renderInventory } from '../characters/inventory';
import { renderSpells } from '../characters/spells';

expose('addInventory', function () {
  showModal('Add Item', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="invName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Quantity</label><input class="form-control" id="invQty" type="number" value="1"></div>
      <div class="col-6"><label class="form-label">Weight (lbs)</label><input class="form-control" id="invWeight" type="number" value="0" step="0.1"></div>
    </div>
    <div class="mb-3"><label class="form-label">Category</label>
      <select class="form-select" id="invCat">
        <option value="gear">Gear</option><option value="weapon">Weapon</option><option value="armor">Armor</option>
        <option value="potion">Potion</option><option value="scroll">Scroll</option><option value="tool">Tool</option>
        <option value="wondrous">Wondrous Item</option><option value="other">Other</option>
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveInventory(this)"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
});

expose('editInventory', function (id:number,name:string,qty:number,cat:string,weight:number,equipped:boolean) {
  showModal('Edit Item', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="invName" value="${esc(name)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Quantity</label><input class="form-control" id="invQty" type="number" value="${qty}"></div>
      <div class="col-6"><label class="form-label">Weight (lbs)</label><input class="form-control" id="invWeight" type="number" value="${weight}" step="0.1"></div>
    </div>
    <div class="mb-3"><label class="form-label">Category</label>
      <select class="form-select" id="invCat">${['gear','weapon','armor','potion','scroll','tool','wondrous','other'].map(c=>`<option value="${c}"${c===cat?' selected':''}>${capitalize(c)}</option>`).join('')}</select></div>
    <div class="mb-3"><div class="form-check"><input type="checkbox" class="form-check-input" id="invEquip"${equipped?' checked':''}><label class="form-check-label">Equipped</label></div></div>
    <button class="btn btn-primary w-100" onclick="saveEditInventory(${id},this)">Save</button>
  `);
});

expose('saveInventory', async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/inventory`, {
    name: (document.getElementById('invName') as HTMLInputElement).value,
    quantity: +(document.getElementById('invQty') as HTMLInputElement).value || 1,
    weight: +(document.getElementById('invWeight') as HTMLInputElement).value || 0,
    category: (document.getElementById('invCat') as HTMLSelectElement).value,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast('Item added');
});

expose('saveEditInventory', async function (id:number, btn:HTMLElement) {
  await api('PUT', `/api/inventory/${id}`, {
    name: (document.getElementById('invName') as HTMLInputElement).value,
    quantity: +(document.getElementById('invQty') as HTMLInputElement).value || 1,
    weight: +(document.getElementById('invWeight') as HTMLInputElement).value || 0,
    category: (document.getElementById('invCat') as HTMLSelectElement).value,
    equipped: (document.getElementById('invEquip') as HTMLInputElement).checked,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast('Item updated');
});

expose('deleteInventory', async function (id:number) {
  if (!confirm('Remove this item?')) return;
  await api('DELETE', `/api/inventory/${id}`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast('Item removed');
});

expose('toggleEquip', async function (id:number) {
  const item = currentChar.inventory.find((i:any) => i.id === id);
  if (!item) return;
  item.equipped = !item.equipped;
  await api('PUT', `/api/inventory/${id}`, item);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast(item.equipped ? 'Equipped' : 'Unequipped');
});

expose('enableSpellcasting', async function () {
  currentChar.spellcasting = {
    ability: 'int', save_dc: 10, attack_bonus: 0,
    slots_1_max: 2, slots_1_used: 0,
  };
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
});

expose('addSpell', function () {
  showModal('Add Spell', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="spellName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="spellLevel" type="number" value="0" min="0" max="9"></div>
      <div class="col-6"><label class="form-label">School</label>
        <select class="form-select" id="spellSchool">${['Abjuration','Conjuration','Divination','Enchantment','Evocation','Illusion','Necromancy','Transmutation'].map(s=>`<option value="${s}">${s}</option>`).join('')}</select></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Casting Time</label><input class="form-control" id="spellCast" value="1 action"></div>
      <div class="col-6"><label class="form-label">Range</label><input class="form-control" id="spellRange" value="Self"></div>
    </div>
    <div class="mb-3"><label class="form-label">Components</label><input class="form-control" id="spellComp" value="V,S"></div>
    <div class="mb-3"><label class="form-label">Duration</label><input class="form-control" id="spellDur" value="Instantaneous"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="spellDesc" rows="3"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveSpell(this)">Add Spell</button>
  `);
});

expose('editSpell', function (id:number,name:string,level:number,school:string,prepared:boolean,comp:string,range:string,cast:string,dur:string,desc:string) {
  showModal('Edit Spell', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="spellName" value="${esc(name)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="spellLevel" type="number" value="${level}" min="0" max="9"></div>
      <div class="col-6"><label class="form-label">School</label>
        <select class="form-select" id="spellSchool">${['Abjuration','Conjuration','Divination','Enchantment','Evocation','Illusion','Necromancy','Transmutation'].map(s=>`<option value="${s}"${s===school?' selected':''}>${s}</option>`).join('')}</select></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Casting Time</label><input class="form-control" id="spellCast" value="${esc(cast)}"></div>
      <div class="col-6"><label class="form-label">Range</label><input class="form-control" id="spellRange" value="${esc(range)}"></div>
    </div>
    <div class="mb-3"><label class="form-label">Components</label><input class="form-control" id="spellComp" value="${esc(comp)}"></div>
    <div class="mb-3"><label class="form-label">Duration</label><input class="form-control" id="spellDur" value="${esc(dur)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="spellDesc" rows="3">${esc(desc)}</textarea></div>
    <div class="mb-3"><div class="form-check"><input type="checkbox" class="form-check-input" id="spellPrep"${prepared?' checked':''}><label class="form-check-label">Prepared</label></div></div>
    <button class="btn btn-primary w-100" onclick="saveEditSpell(${id},this)">Save Spell</button>
  `);
});

expose('saveSpell', async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/spells`, {
    name: (document.getElementById('spellName') as HTMLInputElement).value,
    level: +(document.getElementById('spellLevel') as HTMLInputElement).value || 0,
    school: (document.getElementById('spellSchool') as HTMLSelectElement).value,
    casting_time: (document.getElementById('spellCast') as HTMLInputElement).value,
    range: (document.getElementById('spellRange') as HTMLInputElement).value,
    components: (document.getElementById('spellComp') as HTMLInputElement).value,
    duration: (document.getElementById('spellDur') as HTMLInputElement).value,
    description: (document.getElementById('spellDesc') as HTMLTextAreaElement).value,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell added');
});

expose('saveEditSpell', async function (id:number, btn:HTMLElement) {
  await api('PUT', `/api/spells/${id}`, {
    name: (document.getElementById('spellName') as HTMLInputElement).value,
    level: +(document.getElementById('spellLevel') as HTMLInputElement).value || 0,
    school: (document.getElementById('spellSchool') as HTMLSelectElement).value,
    casting_time: (document.getElementById('spellCast') as HTMLInputElement).value,
    range: (document.getElementById('spellRange') as HTMLInputElement).value,
    components: (document.getElementById('spellComp') as HTMLInputElement).value,
    duration: (document.getElementById('spellDur') as HTMLInputElement).value,
    description: (document.getElementById('spellDesc') as HTMLTextAreaElement).value,
    prepared: (document.getElementById('spellPrep') as HTMLInputElement).checked,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell updated');
});

expose('deleteSpell', async function (id:number) {
  if (!confirm('Remove this spell?')) return;
  await api('DELETE', `/api/spells/${id}`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell removed');
});

expose('saveSpellPrep', async function () {
  const spellIds: number[] = [];
  (currentChar.spells || []).forEach((s:any) => {
    const cb = document.getElementById(`prep-${s.id}`) as HTMLInputElement;
    if (cb && cb.checked) spellIds.push(s.id);
  });
  await api('PUT', `/api/characters/${currentChar.id}/spells/prepare`, { spell_ids: spellIds });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell preparation saved');
});
