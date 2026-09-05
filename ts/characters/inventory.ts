import { expose } from '../lib/expose';
import { currentChar, setCurrentChar } from '../lib/state';
import { esc, attrEscape, capitalize, toast, openCompendiumPicker } from '../lib/dom';
import { api, getCsrfToken, getApiToken } from '../lib/api';
import type { Character, InventoryItem } from '../lib/api-types';

type InvItem = InventoryItem & {
  ac_bonus?: number;
  damage_dice?: string;
  damage_type?: string;
  attunement?: boolean;
  is_identified?: boolean;
  magic?: boolean;
  compendium_equipment_id?: number;
  compendium_entry_id?: number;
};

export async function updateCurrency(): Promise<void> {
  if (!currentChar) return;
  const coins = ['cp','sp','ep','gp','pp'];
  const updates: Record<string,number> = {};
  coins.forEach(c => { updates[c] = +((document.getElementById('coin' + c) as HTMLInputElement | null)?.value || '0') || 0; });
  await api<void>('PUT', `/api/characters/${(currentChar as Character).id}/currency`, updates);
  setCurrentChar(await api<Character>('GET', `/api/characters/${(currentChar as Character).id}`));
  toast('Currency updated');
}

export function renderInventory(): void {
  if (!currentChar) return;
  const inv = ((currentChar as Character)['inventory'] as InvItem[] | undefined) || [];
  const c = currentChar as Character & Record<string, unknown>;
  const categories: Record<string, InvItem[]> = { weapon: [], armor: [], gear: [], potion: [], scroll: [], tool: [], wondrous: [], other: [] };
  inv.forEach((i) => { if (categories[i.category || 'other']) categories[i.category || 'other']!.push(i); else categories['other']!.push(i); });
  const total = inv.reduce((s, i)=>s+(i.weight||0)*(i.quantity||1),0);

  const str = (c['str'] as number) || 10;
  const lightMax = str * 5;
  const encumberedMax = str * 10;
  const heavyMax = str * 15;
  const encPct = heavyMax > 0 ? Math.min(100, (total / heavyMax) * 100) : 0;
  let encState = 'light';
  let encLabel = 'Light Load';
  if (total > heavyMax) { encState = 'over'; encLabel = 'Over Capacity'; }
  else if (total > encumberedMax) { encState = 'heavy'; encLabel = 'Heavily Encumbered'; }
  else if (total > lightMax) { encState = 'encumbered'; encLabel = 'Encumbered'; }

  const attuneItems = inv.filter((i) => i.equipped && i.attunement);
  const attuneCount = attuneItems.length;
  let attuneState = 'attune-ok';
  if (attuneCount >= 3) attuneState = 'attune-full';
  else if (attuneCount >= 2) attuneState = 'attune-warn';

  const equipped = inv.filter((i) => i.equipped);
  const loadoutGroups: Record<string, InvItem[]> = { weapon: [], armor: [], shield: [], ring: [], wondrous: [], other: [] };
  equipped.forEach((i) => {
    if (i.category === 'weapon') loadoutGroups['weapon']!.push(i);
    else if (i.category === 'armor' && (i.ac_bonus||0) > 0 && i.name.toLowerCase().includes('shield')) loadoutGroups['shield']!.push(i);
    else if (i.category === 'armor') loadoutGroups['armor']!.push(i);
    else if (i.category === 'wondrous') loadoutGroups['wondrous']!.push(i);
    else loadoutGroups['other']!.push(i);
  });

  const invSection = document.getElementById('inventorySection');
  if (!invSection) return;
  invSection.innerHTML = `
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
            ${(items as InvItem[]).map((i) => `
              <div class="loadout-item">
                <span class="item-name">${esc(i.name)}</span>
                <span class="item-detail">${i.damage_dice ? esc(i.damage_dice) + (i.damage_type ? ' ' + esc(i.damage_type) : '') : (i.ac_bonus||0) > 0 ? 'AC +' + i.ac_bonus : ''}</span>
              </div>
            `).join('')}
          </div>
        `).join('') || '<div class="text-muted small fst-italic">No items equipped.</div>'}
      </div>
    </div>
    <div class="mt-2" id="invList">
      ${Object.entries(categories).filter(([,items]) => items.length).map(([cat, items]) => `
        <h6 class="mt-3 text-muted">${capitalize(cat)}</h6>
        ${(items as InvItem[]).map((i) => `
          <div class="inv-item${i.equipped ? ' equipped' : ''}${i.is_identified === false ? ' unidentified' : ''}">
            <div>
              <span class="fw-bold">${esc(i.name)}</span>
              ${(i.quantity||0) > 1 ? `<span class="badge badge-muted">x${i.quantity}</span>` : ''}
              ${i.equipped ? '<span class="badge badge-gold">Equipped</span>' : ''}
              ${i.attunement ? '<span class="badge-attunement" title="Requires Attunement">Attune</span>' : ''}
              ${i.is_identified === false ? '<span class="badge-unidentified">Unidentified</span>' : ''}
              ${i.damage_dice && (i.is_identified !== false) ? `<span class="badge badge-blood ms-1">${esc(i.damage_dice)} ${esc(i.damage_type)}</span>` : ''}
              ${(i.ac_bonus||0) > 0 && (i.is_identified !== false) ? `<span class="badge badge-gold ms-1">AC+${i.ac_bonus}</span>` : ''}
              ${i.is_identified === false && i.damage_dice ? `<span class="badge badge-muted ms-1">???</span>` : ''}
              ${(i.compendium_equipment_id || i.compendium_entry_id) ? '<span class="badge badge-compendium ms-1" title="Linked from compendium"><i class="fa-solid fa-book me-1"></i></span>' : ''}
            </div>
            <div class="d-flex gap-1">
              ${i.is_identified === false ? `<button class="btn-identify" onclick="toggleIdentify(${i.id})" title="Identify item">🔍 ID</button>` : ''}
              ${i.magic && i.is_identified !== false ? `<button class="btn-identify" onclick="toggleIdentify(${i.id})" title="Mark unidentified">🔮</button>` : ''}
              <button class="btn btn-sm btn-outline-primary" onclick="editInventory(${i.id},'${attrEscape(i.name)}',${i.quantity},'${attrEscape(i.category||'')}',${i.weight},${i.equipped})" title="Edit"><i class="fa-solid fa-pen"></i></button>
              ${(i.compendium_equipment_id || i.compendium_entry_id) ? `<button class="btn btn-sm btn-outline-secondary" onclick="unlinkCompendiumItem(${i.id})" title="Unlink from compendium"><i class="fa-solid fa-link-slash"></i></button>` : ''}
              <button class="btn btn-sm btn-outline-secondary" onclick="toggleEquip(${i.id})" title="${i.equipped ? 'Unequip' : 'Equip'}"><i class="fa-solid fa-shield-halved"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteInventory(${i.id})" title="Remove"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>`).join('')}
      `).join('') || '<div class="empty-state"><i class="fa-solid fa-backpack fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Pockets</p><p class="small text-muted">No items yet. Add gear to your inventory.</p></div>'}
    </div>`;
}

expose('updateCurrency', updateCurrency);
expose('renderInventory', renderInventory);

interface CompendiumItem { id: number; name: string; source?: string; category?: string; item_rarity?: string; weight?: number; schema_name?: string }

expose('openInventoryPicker', function () {
  openCompendiumPicker({
    title: 'Link from Compendium',
    placeholder: 'Search equipment...',
    search: (q) => api<CompendiumItem[]>('GET', `/api/compendium/equipment?q=${encodeURIComponent(q)}`),
    render: (e: CompendiumItem) => `<div><span class="fw-bold">${esc(e.name)}</span>
      ${e.category ? `<span class="text-muted small ms-1">${esc(e.category)}</span>` : ''}
      ${e.item_rarity ? `<span class="text-muted small"> · ${esc(e.item_rarity)}</span>` : ''}
      ${e.weight ? `<span class="text-muted small"> · ${e.weight} lbs</span>` : ''}
      ${e.source === 'entry' && e.schema_name ? `<span class="badge bg-info ms-1">${esc(e.schema_name)}</span>` : ''}</div>`,
    onPick: (e: CompendiumItem) => { linkCompendiumItem(e).catch((err: Error) => toast(err.message, true)); },
  });
});

async function linkCompendiumItem(item: CompendiumItem): Promise<void> {
  const fd = new FormData();
  fd.append(item.source === 'entry' ? 'compendium_entry_id' : 'compendium_equipment_id', String(item.id));
  fd.append('quantity', '1');
  const headers: Record<string, string> = {};
  const csrf = getCsrfToken();
  if (csrf) headers['X-CSRF-Token'] = csrf;
  const apiToken = getApiToken();
  if (apiToken) headers['Authorization'] = `Bearer ${apiToken}`;
  const res = await fetch(`/api/characters/${(currentChar as Character).id}/inventory/link`, { method: 'POST', body: fd, headers, credentials: 'include' });
  if (!res.ok) throw new Error(((await res.json().catch(() => ({}))) as { error?: string }).error || 'Link failed');
  setCurrentChar(await api<Character>('GET', `/api/characters/${(currentChar as Character).id}`));
  renderInventory();
  toast(`Linked ${item.name} to inventory`);
}

expose('unlinkCompendiumItem', async function (itemId: number): Promise<void> {
  await api<void>('DELETE', `/api/characters/${(currentChar as Character).id}/inventory/${itemId}/link`);
  setCurrentChar(await api<Character>('GET', `/api/characters/${(currentChar as Character).id}`));
  renderInventory();
  toast('Unlinked from compendium (item kept)');
});
