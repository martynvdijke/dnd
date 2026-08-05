// @ts-nocheck — extracted from app.ts monolith (address-tech-debt-and-ux)
import { expose } from '../lib/expose';
import { currentChar, setCurrentChar } from '../lib/state';
import { esc, capitalize, toast, openCompendiumPicker } from '../lib/dom';
import { api, getCsrfToken } from '../lib/api';

export async function updateCurrency() {
  if (!currentChar) return;
  const coins = ['cp','sp','ep','gp','pp'];
  const updates: Record<string,number> = {};
  coins.forEach(c => { updates[c] = +(document.getElementById('coin' + c) as HTMLInputElement)?.value || 0; });
  await api('PUT', `/api/characters/${currentChar.id}/currency`, updates);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  toast('Currency updated');
}

export function renderInventory() {
  const inv = currentChar.inventory || [];
  const c = currentChar;
  const categories: Record<string, any[]> = { weapon: [], armor: [], gear: [], potion: [], scroll: [], tool: [], wondrous: [], other: [] };
  inv.forEach((i:any) => { if (categories[i.category]) categories[i.category].push(i); else categories.other.push(i); });
  const total = inv.reduce((s:number,i:any)=>s+(i.weight||0)*(i.quantity||1),0);

  // Encumbrance: STR x5 (light), x10 (encumbered), x15 (heavy)
  const str = c.str || 10;
  const lightMax = str * 5;
  const encumberedMax = str * 10;
  const heavyMax = str * 15;
  const encPct = heavyMax > 0 ? Math.min(100, (total / heavyMax) * 100) : 0;
  let encState = 'light';
  let encLabel = 'Light Load';
  if (total > heavyMax) { encState = 'over'; encLabel = 'Over Capacity'; }
  else if (total > encumberedMax) { encState = 'heavy'; encLabel = 'Heavily Encumbered'; }
  else if (total > lightMax) { encState = 'encumbered'; encLabel = 'Encumbered'; }

  // Attunement count (equipped items with attunement flag)
  const attuneItems = inv.filter((i:any) => i.equipped && i.attunement);
  const attuneCount = attuneItems.length;
  let attuneState = 'attune-ok';
  if (attuneCount >= 3) attuneState = 'attune-full';
  else if (attuneCount >= 2) attuneState = 'attune-warn';

  // Equipped items grouped for loadout
  const equipped = inv.filter((i:any) => i.equipped);
  const loadoutGroups: Record<string, any[]> = { weapon: [], armor: [], shield: [], ring: [], wondrous: [], other: [] };
  equipped.forEach((i:any) => {
    if (i.category === 'weapon') loadoutGroups.weapon.push(i);
    else if (i.category === 'armor' && i.ac_bonus > 0 && i.name.toLowerCase().includes('shield')) loadoutGroups.shield.push(i);
    else if (i.category === 'armor') loadoutGroups.armor.push(i);
    else if (i.category === 'wondrous') loadoutGroups.wondrous.push(i);
    else loadoutGroups.other.push(i);
  });

  document.getElementById('inventorySection')!.innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5>Inventory <span class="text-muted small">(Total: ${total} / ${heavyMax} lbs)</span></h5>
      <div class="d-flex gap-2 align-items-center">
        <span class="attune-counter ${attuneState}" title="Attuned items">🔗 ${attuneCount}/3</span>
        <button class="btn btn-outline-primary btn-sm" onclick="openInventoryPicker()"><i class="fa-solid fa-book me-1"></i>Link from Compendium</button>
        <button class="btn btn-primary btn-sm" onclick="addInventory()"><i class="fa-solid fa-plus me-1"></i>Add Item</button>
      </div>
    </div>
    <div class="encumbrance-bar" title="${total} lbs / ${heavyMax} lbs max">
      <div class="encumbrance-bar-fill enc-${encState}" style="width:${encPct}%"></div>
    </div>
    <div class="d-flex justify-content-between">
      <span class="encumbrance-state enc-${encState}">${esc(encLabel)}</span>
      <span class="text-muted small">${encLabel === 'Light Load' ? 'Speed normal' : encLabel === 'Encumbered' ? 'Speed -10' : encLabel === 'Heavily Encumbered' ? 'Speed -20, Disadvantage on checks' : 'Speed 0'}</span>
    </div>
    <!-- Loadout Panel -->
    <div class="loadout-panel mt-2">
      <div class="loadout-header" onclick="this.nextElementSibling.classList.toggle('d-none')">
        <h6><i class="fa-solid fa-shield-halved me-1"></i>Loadout (${equipped.length} equipped)</h6>
        <span class="text-muted small"><i class="fa-solid fa-chevron-down"></i></span>
      </div>
      <div class="loadout-body">
        ${Object.entries(loadoutGroups).filter(([,items]) => items.length).map(([cat, items]) => `
          <div class="loadout-category">
            <div class="loadout-category-label">${capitalize(cat)}</div>
            ${(items as any[]).map((i:any) => `
              <div class="loadout-item">
                <span class="item-name">${esc(i.name)}</span>
                <span class="item-detail">${i.damage_dice ? esc(i.damage_dice) + (i.damage_type ? ' ' + esc(i.damage_type) : '') : i.ac_bonus > 0 ? 'AC +' + i.ac_bonus : ''}</span>
              </div>
            `).join('')}
          </div>
        `).join('') || '<div class="text-muted small fst-italic">No items equipped.</div>'}
      </div>
    </div>
    <div class="mt-2" id="invList">
      ${Object.entries(categories).filter(([,items]) => items.length).map(([cat, items]) => `
        <h6 class="mt-3 text-muted">${capitalize(cat)}</h6>
        ${(items as any[]).map((i:any) => `
          <div class="inv-item${i.equipped ? ' equipped' : ''}${i.is_identified === false ? ' unidentified' : ''}">
            <div>
              <span class="fw-bold">${esc(i.name)}</span>
              ${i.quantity > 1 ? `<span class="badge badge-muted">x${i.quantity}</span>` : ''}
              ${i.equipped ? '<span class="badge badge-gold">Equipped</span>' : ''}
              ${i.attunement ? '<span class="badge-attunement" title="Requires Attunement">Attune</span>' : ''}
              ${i.is_identified === false ? '<span class="badge-unidentified">Unidentified</span>' : ''}
              ${i.damage_dice && (i.is_identified !== false) ? `<span class="badge badge-blood ms-1">${esc(i.damage_dice)} ${esc(i.damage_type)}</span>` : ''}
              ${i.ac_bonus > 0 && (i.is_identified !== false) ? `<span class="badge badge-gold ms-1">AC+${i.ac_bonus}</span>` : ''}
              ${i.is_identified === false && i.damage_dice ? `<span class="badge badge-muted ms-1">???</span>` : ''}
              ${i.compendium_equipment_id ? '<span class="badge badge-compendium ms-1" title="Linked from compendium"><i class="fa-solid fa-book me-1"></i></span>' : ''}
            </div>
            <div class="d-flex gap-1">
              ${i.is_identified === false ? `<button class="btn-identify" onclick="toggleIdentify(${i.id})" title="Identify item">🔍 ID</button>` : ''}
              ${i.magic && i.is_identified !== false ? `<button class="btn-identify" onclick="toggleIdentify(${i.id})" title="Mark unidentified">🔮</button>` : ''}
              <button class="btn btn-sm btn-outline-primary" onclick="editInventory(${i.id},'${esc(i.name)}',${i.quantity},'${esc(i.category)}',${i.weight},${i.equipped})" title="Edit"><i class="fa-solid fa-pen"></i></button>
              ${i.compendium_equipment_id ? `<button class="btn btn-sm btn-outline-secondary" onclick="unlinkCompendiumItem(${i.id})" title="Unlink from compendium"><i class="fa-solid fa-link-slash"></i></button>` : ''}
              <button class="btn btn-sm btn-outline-secondary" onclick="toggleEquip(${i.id})" title="${i.equipped ? 'Unequip' : 'Equip'}"><i class="fa-solid fa-shield-halved"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteInventory(${i.id})" title="Remove"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>`).join('')}
      `).join('') || '<div class="empty-state"><i class="fa-solid fa-backpack fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Pockets</p><p class="small text-muted">No items yet. Add gear to your inventory.</p></div>'}
    </div>`;
}

expose('updateCurrency', updateCurrency);
expose('renderInventory', renderInventory);

expose('openInventoryPicker', function () {
  openCompendiumPicker({
    title: 'Link from Compendium',
    placeholder: 'Search equipment...',
    search: (q) => api('GET', `/api/compendium/equipment?q=${encodeURIComponent(q)}`),
    render: (e: any) => `<div><span class="fw-bold">${esc(e.name)}</span>
      ${e.category ? `<span class="text-muted small ms-1">${esc(e.category)}</span>` : ''}
      ${e.item_rarity ? `<span class="text-muted small"> · ${esc(e.item_rarity)}</span>` : ''}
      ${e.weight ? `<span class="text-muted small"> · ${e.weight} lbs</span>` : ''}</div>`,
    onPick: (e: any) => { linkCompendiumItem(e).catch((err: Error) => toast(err.message, true)); },
  });
});

async function linkCompendiumItem(item: any) {
  const fd = new FormData();
  fd.append('compendium_equipment_id', String(item.id));
  fd.append('quantity', '1');
  const headers: Record<string, string> = {};
  const csrf = getCsrfToken();
  if (csrf) headers['X-CSRF-Token'] = csrf;
  const res = await fetch(`/api/characters/${currentChar.id}/inventory/link`, { method: 'POST', body: fd, headers, credentials: 'include' });
  if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'Link failed');
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast(`Linked ${item.name} to inventory`);
}

expose('unlinkCompendiumItem', async function (itemId: number) {
  await api('DELETE', `/api/characters/${currentChar.id}/inventory/${itemId}/link`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast('Unlinked from compendium (item kept)');
});
