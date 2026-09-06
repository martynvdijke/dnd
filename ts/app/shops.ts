// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, attrEscape, capitalize, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { currentUser, currentChar, allLocations } from '../lib/state';
import { showView } from '../navigation';

// ─── Shops State ───

let selectedShopShops: any[] = [];
let allCampaigns: any[] = [];
let compendiumEquipment: any[] = [];

async function ensureCompendiumEquipment() {
  if (compendiumEquipment.length === 0) {
    try { compendiumEquipment = await api('GET', '/api/compendium/equipment'); } catch {}
  }
}

// ─── Main Shops View ───

expose('showShops', async function () {
  showView('shops');
  const el = document.getElementById('shopsContent')!;
  try {
    const [shops, campaigns] = await Promise.all([
      api('GET', '/api/shops'),
      api('GET', '/api/campaigns').catch(() => []),
    ]);
    allCampaigns = campaigns;
    selectedShopShops = shops;

    const isDM = currentUser?.role === 'dm' || currentUser?.role === 'admin';

    let html = `<div class="d-flex justify-content-between align-items-center flex-wrap gap-2 mb-3">
      <h1 class="h4 mb-0"><i class="fa-solid fa-store me-2"></i>Shops & Trading</h1>
      <div class="d-flex gap-2">
        ${isDM ? `<button class="btn btn-gold btn-sm" onclick="showNewShop()"><i class="fa-solid fa-plus me-1"></i>New Shop</button>` : ''}
        <button class="btn btn-outline-primary btn-sm" onclick="showTransactions()"><i class="fa-solid fa-receipt me-1"></i>Transactions</button>
      </div>
    </div>`;

    html += `<div class="mb-3"><input class="form-control" id="shopSearch" placeholder="Search shops by name, location..." oninput="filterShops()"></div>`;
    html += `<div class="row g-3" id="shopsGrid">`;

    if (!shops.length) {
      html += `<div class="col-12"><div class="empty-state"><i class="fa-solid fa-store fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Shops Found</p><p class="small text-muted">${isDM ? 'Create your first shop using the button above.' : 'Ask your DM to add shops to the world!'}</p></div></div>`;
    } else {
      for (const s of shops) {
        html += renderShopCard(s, isDM);
      }
    }

    html += `</div>`;

    el.innerHTML = html;
  } catch (e: any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load shops: ${esc(e.message)}</p></div>`;
  }
});

function renderShopCard(s: any, isDM: boolean): string {
  const locName = s.location_name || 'Anywhere';
  return `<div class="col-md-6 col-lg-4">
    <div class="character-card" onclick="showShopDetail(${s.id})">
      <div class="d-flex justify-content-between align-items-start">
        <div class="char-name mb-1">${esc(s.name)}</div>
        ${isDM ? `<div class="d-flex gap-1" onclick="event.stopPropagation()">
          <button class="btn btn-sm btn-outline-primary py-0 px-1" onclick="showEditShop(${s.id})" title="Edit"><i class="fa-solid fa-pen" style="font-size:0.65rem"></i></button>
          <button class="btn btn-sm btn-outline-danger py-0 px-1" onclick="deleteShop(${s.id})" title="Delete"><i class="fa-solid fa-trash" style="font-size:0.65rem"></i></button>
        </div>` : ''}
      </div>
      ${s.description ? `<div class="char-detail small">${esc(s.description)}</div>` : ''}
      <div class="mt-1 d-flex gap-2 flex-wrap">
        <span class="badge badge-gold" style="font-size:0.6rem"><i class="fa-solid fa-location-dot me-1"></i>${esc(locName)}</span>
        <span class="badge badge-blood" style="font-size:0.6rem">Buy +${s.markup_percent}%</span>
        <span class="badge badge-muted" style="font-size:0.6rem">Sell ${s.markup_buy_percent}%</span>
      </div>
    </div>
  </div>`;
}

expose('filterShops', function () {
  const q = (document.getElementById('shopSearch') as HTMLInputElement)?.value?.toLowerCase() || '';
  document.querySelectorAll('#shopsGrid .character-card').forEach(card => {
    const parent = card.closest('.col-md-6, .col-lg-4') as HTMLElement;
    if (parent) parent.style.display = !q || card.textContent?.toLowerCase().includes(q) ? '' : 'none';
  });
});

// ─── Shop CRUD (DM) ───

expose('showNewShop', function () {
  const campaignOpts = allCampaigns.map((c: any) =>
    `<option value="${c.id}">${esc(c.name)}${c.party_name ? ' (' + esc(c.party_name) + ')' : ''}</option>`
  ).join('');

  showModal('New Shop', `
    <div class="mb-3"><label class="form-label">Shop Name</label><input class="form-control" id="shopName" placeholder="e.g. The Rusty Sword"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopDesc" rows="2" placeholder="What does this shop sell?"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Buy Markup %</label><input class="form-control" id="shopMarkup" type="number" value="0" step="1"><small class="text-muted">Added to item prices</small></div>
      <div class="col-6"><label class="form-label">Sell % of Value</label><input class="form-control" id="shopSellMarkup" type="number" value="50" step="1"><small class="text-muted">% of price paid to players</small></div>
    </div>
    <div class="mb-3"><label class="form-label">Campaign</label>
      <select class="form-select" id="shopCampaign">
        ${campaignOpts || '<option value="">No campaigns available</option>'}
      </select></div>
    <div class="mb-3"><label class="form-label">Location (optional)</label>
      <select class="form-select" id="shopLocation">
        <option value="">No specific location</option>
        ${allLocations.map((l: any) => `<option value="${l.id}">${esc(l.name)} (${esc(l.type)})</option>`).join('')}
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveShop()"><i class="fa-solid fa-plus me-1"></i>Create Shop</button>
  `);
});

expose('saveShop', async function (shopId?: number) {
  const name = (document.getElementById('shopName') as HTMLInputElement).value;
  if (!name) { toast('Shop name is required', true); return; }

  const data: any = {
    name,
    description: (document.getElementById('shopDesc') as HTMLTextAreaElement).value,
    markup_percent: +(document.getElementById('shopMarkup') as HTMLInputElement).value || 0,
    markup_buy_percent: +(document.getElementById('shopSellMarkup') as HTMLInputElement).value || 50,
  };
  const locVal = (document.getElementById('shopLocation') as HTMLSelectElement).value;
  if (locVal) data.location_id = parseInt(locVal);

  try {
    if (shopId) {
      await api('PUT', `/api/shops/${shopId}`, data);
      toast('Shop updated');
    } else {
      const campaignId = (document.getElementById('shopCampaign') as HTMLSelectElement).value;
      if (!campaignId) { toast('Campaign is required', true); return; }
      await api('POST', `/api/campaigns/${campaignId}/shops`, data);
      toast('Shop created');
    }
    hideModal();
    (window as any).showShops();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('showEditShop', async function (shopId: number) {
  const shop = selectedShopShops.find((s: any) => s.id === shopId);
  if (!shop) { toast('Shop not found', true); return; }

  showModal(`Edit: ${esc(shop.name)}`, `
    <div class="mb-3"><label class="form-label">Shop Name</label><input class="form-control" id="shopName" value="${esc(shop.name)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopDesc" rows="2">${esc(shop.description)}</textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Buy Markup %</label><input class="form-control" id="shopMarkup" type="number" value="${shop.markup_percent}" step="1"></div>
      <div class="col-6"><label class="form-label">Sell % of Value</label><input class="form-control" id="shopSellMarkup" type="number" value="${shop.markup_buy_percent}" step="1"></div>
    </div>
    <div class="mb-3"><label class="form-label">Location (optional)</label>
      <select class="form-select" id="shopLocation">
        <option value="">No specific location</option>
        ${allLocations.map((l: any) => `<option value="${l.id}" ${shop.location_id === l.id ? 'selected' : ''}>${esc(l.name)} (${esc(l.type)})</option>`).join('')}
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveShop(${shopId})"><i class="fa-solid fa-save me-1"></i>Save Changes</button>
  `);
});

expose('deleteShop', async function (shopId: number) {
  if (!confirm('Delete this shop? All items will be removed.')) return;
  try {
    await api('DELETE', `/api/shops/${shopId}`);
    toast('Shop deleted');
    (window as any).showShops();
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Shop Items Detail ───

expose('showShopDetail', async function (shopId: number) {
  try {
    const [items, shops] = await Promise.all([
      api('GET', `/api/shops/${shopId}/items`),
      selectedShopShops.length ? Promise.resolve(selectedShopShops) : api('GET', '/api/shops'),
    ]);
    selectedShopShops = shops;
    const shop = shops.find((s: any) => s.id === shopId);
    if (!shop) { toast('Shop not found', true); return; }

    const isDM = currentUser?.role === 'dm' || currentUser?.role === 'admin';
    const hasChar = !!currentChar;

    showModal(esc(shop.name), `
      <div class="shop-detail-modal">
        ${shop.description ? `<p class="text-muted small">${esc(shop.description)}</p>` : ''}
        <div class="d-flex gap-2 mb-3 flex-wrap">
          <span class="badge badge-gold"><i class="fa-solid fa-location-dot me-1"></i>${esc(shop.location_name || 'Anywhere')}</span>
          ${items.length > 0 ? `<span class="badge badge-muted"><i class="fa-solid fa-box me-1"></i>${items.length} items</span>` : ''}
          <span class="badge badge-blood">Buy: +${shop.markup_percent}%</span>
          <span class="badge badge-blood">Sell: ${shop.markup_buy_percent}%</span>
        </div>

        <div class="d-flex justify-content-between align-items-center mb-2">
          <h6 class="mb-0">Stock</h6>
          ${isDM ? `<button class="btn btn-sm btn-gold" onclick="showAddShopItem(${shopId})"><i class="fa-solid fa-plus me-1"></i>Add Item</button>` : ''}
        </div>

        <div id="shopItemsList">
          ${items.length === 0 ? '<div class="text-muted small text-center py-3 fst-italic">No items in stock.</div>' :
            items.map((i: any) => {
              const effPrice = Math.round(i.price * (1 + shop.markup_percent / 100) * 100) / 100;
              return `<div class="inv-item">
                <div>
                  <span class="fw-bold">${esc(i.name)}</span>
                  ${i.category ? `<span class="badge badge-muted ms-1" style="font-size:0.6rem">${esc(i.category)}</span>` : ''}
                  ${i.compendium_equipment_id ? `<span class="badge badge-compendium ms-1" title="Linked from compendium" style="font-size:0.6rem"><i class="fa-solid fa-book me-1"></i></span>` : ''}
                  ${i.quantity > 0
                    ? `<span class="badge badge-gold ms-1" style="font-size:0.6rem">x${i.quantity}</span>`
                    : `<span class="badge bg-danger ms-1" style="font-size:0.6rem">Out of stock</span>`}
                  ${i.description ? `<div class="small text-muted">${esc(i.description)}</div>` : ''}
                </div>
                <div class="d-flex align-items-center gap-1 flex-nowrap">
                  <span class="fw-bold text-nowrap" style="color:var(--gold);min-width:50px;text-align:right">${effPrice} gp</span>
                  ${i.quantity > 0 && hasChar ? `<button class="btn btn-sm btn-gold py-0 px-1" onclick="buyItem(${shopId}, ${i.id})" style="font-size:0.65rem">Buy</button>` : ''}
                  ${isDM ? `
                    <button class="btn btn-sm btn-outline-primary py-0 px-1" onclick="editShopItem(${i.id})" title="Edit" style="font-size:0.65rem"><i class="fa-solid fa-pen"></i></button>
                    ${i.compendium_equipment_id ? `<button class="btn btn-sm btn-outline-secondary py-0 px-1" onclick="unlinkShopItem(${shopId}, ${i.id})" title="Unlink from compendium" style="font-size:0.65rem"><i class="fa-solid fa-link-slash"></i></button>` : ''}
                    <button class="btn btn-sm btn-outline-danger py-0 px-1" onclick="deleteShopItem(${i.id})" title="Remove" style="font-size:0.65rem"><i class="fa-solid fa-trash"></i></button>
                  ` : ''}
                </div>
              </div>`;
            }).join('')}
        </div>

        ${hasChar ? `<hr><button class="btn btn-outline-primary w-100" onclick="hideModal();sellItem(${shopId})"><i class="fa-solid fa-sack-dollar me-1"></i>Sell Items to Shop</button>` : ''}
        ${!hasChar ? '<p class="small text-muted mt-2 text-center fst-italic">Open a character sheet to buy or sell items.</p>' : ''}
      </div>
    `);
    // Dynamically add modal-lg class for wider shop view
    const dialog = document.querySelector('#genericModal .modal-dialog') as HTMLElement;
    if (dialog) dialog.classList.add('modal-lg');
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Shop Item CRUD (DM) ───

expose('showAddShopItem', async function (shopId: number) {
  await ensureCompendiumEquipment();
  showModal('Add Shop Item', `
    <input type="hidden" id="shopItemCompId" value="">
    <div class="mb-3"><label class="form-label">Item Name</label>
      <input class="form-control" id="shopItemName" list="compEquipSuggestions" placeholder="Start typing for compendium suggestions...">
      <datalist id="compEquipSuggestions">
        ${compendiumEquipment.map((i: any) => `<option value="${esc(i.name)}">`).join('')}
      </datalist>
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopItemDesc" rows="2"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">Price (gp)</label><input class="form-control" id="shopItemPrice" type="number" value="1" step="0.01"></div>
      <div class="col-4"><label class="form-label">Quantity</label><input class="form-control" id="shopItemQty" type="number" value="1" min="0"></div>
      <div class="col-4"><label class="form-label">Category</label>
        <select class="form-select" id="shopItemCat">
          <option value="">General</option><option value="weapon">Weapon</option><option value="armor">Armor</option>
          <option value="potion">Potion</option><option value="scroll">Scroll</option><option value="gear">Gear</option>
          <option value="ammunition">Ammunition</option><option value="food">Food/Drink</option><option value="other">Other</option>
        </select></div>
    </div>
    <div class="d-flex gap-2">
      <button class="btn btn-outline-gold btn-sm flex-grow-1" onclick="pickCompItem('shopItemName')"><i class="fa-solid fa-book me-1"></i>From Compendium</button>
    </div>
    <div id="shopItemLinkStatus" class="small text-muted mt-1"></div>
    <div class="mt-3">
      <button class="btn btn-primary w-100" onclick="saveShopItem(${shopId})"><i class="fa-solid fa-plus me-1"></i>Add Item</button>
    </div>
  `);
});

expose('pickCompItem', function (inputId: string) {
  showModal('Pick from Compendium', `
    <div class="mb-3"><input class="form-control" id="compPickSearch" placeholder="Search equipment..." oninput="filterCompPick()"></div>
    <div id="compPickList" style="max-height:300px;overflow-y:auto">
      ${compendiumEquipment.map((i: any) => `
        <div class="inv-item comp-pick-item" data-id="${i.id}" data-name="${esc(i.name)}" data-cost="${i.cost || i.price || 1}" data-cat="${i.category || ''}"
             onclick="selectCompPick('${inputId}')" style="cursor:pointer">
          <div><span class="fw-bold small">${esc(i.name)}</span>
            <span class="text-muted small">${i.cost ? i.cost + ' gp' : ''} ${i.category ? '· ' + esc(i.category) : ''}</span></div>
        </div>
      `).join('')}
    </div>
  `);
});

// Records the picked compendium item id on the shop item form and updates the status line.
function shopItemSetLink(id: number | null, name: string) {
  const hidden = document.getElementById('shopItemCompId') as HTMLInputElement;
  if (hidden) hidden.value = id ? String(id) : '';
  const status = document.getElementById('shopItemLinkStatus');
  if (status) {
    status.innerHTML = id
      ? `<span class="badge badge-compendium"><i class="fa-solid fa-book me-1"></i>Linked from compendium: ${esc(name)}</span>`
      : '';
  }
}

expose('filterCompPick', function () {
  const q = (document.getElementById('compPickSearch') as HTMLInputElement)?.value?.toLowerCase() || '';
  document.querySelectorAll('.comp-pick-item').forEach(el => {
    const name = el.getAttribute('data-name') || '';
    (el as HTMLElement).style.display = !q || name.toLowerCase().includes(q) ? '' : 'none';
  });
});

expose('selectCompPick', function (inputId: string) {
  const active = document.activeElement as HTMLElement;
  if (active && active.classList.contains('comp-pick-item')) {
    const id = parseInt(active.getAttribute('data-id') || '0', 10);
    const name = active.getAttribute('data-name') || '';
    const cost = active.getAttribute('data-cost') || '1';
    const cat = active.getAttribute('data-cat') || '';
    const input = document.getElementById(inputId) as HTMLInputElement;
    if (input) input.value = name;
    shopItemSetLink(id, name);
    const priceInput = document.getElementById('shopItemPrice') as HTMLInputElement;
    if (priceInput && priceInput.value === '1') priceInput.value = cost;
    const catSelect = document.getElementById('shopItemCat') as HTMLSelectElement;
    if (catSelect && cat) { const opt = Array.from(catSelect.options).find(o => o.value === cat); if (opt) opt.selected = true; }
    hideModal();
  }
});

// This is handled by onclick directly on the comp-pick-item elements
document.addEventListener('click', function compPickHandler(e) {
  const target = e.target as HTMLElement;
  const item = target.closest('.comp-pick-item') as HTMLElement;
  if (item) {
    const inputId = item.closest('[onclick*="selectCompPick"]')?.getAttribute('onclick')?.match(/'([^']+)'/)?.pop() || 'shopItemName';
    const id = parseInt(item.getAttribute('data-id') || '0', 10);
    const name = item.getAttribute('data-name') || '';
    const cost = item.getAttribute('data-cost') || '1';
    const cat = item.getAttribute('data-cat') || '';
    const input = document.getElementById(inputId) as HTMLInputElement;
    if (input) input.value = name;
    shopItemSetLink(id, name);
    const priceInput = document.getElementById('shopItemPrice') as HTMLInputElement;
    if (priceInput && priceInput.value === '1') priceInput.value = cost;
    const catSelect = document.getElementById('shopItemCat') as HTMLSelectElement;
    if (catSelect && cat) { const opt = Array.from(catSelect.options).find(o => o.value === cat); if (opt) opt.selected = true; }
    hideModal();
  }
});

expose('saveShopItem', async function (shopId: number, itemId?: number) {
  const name = (document.getElementById('shopItemName') as HTMLInputElement).value;
  if (!name) { toast('Item name is required', true); return; }

  const compIdEl = document.getElementById('shopItemCompId') as HTMLInputElement;
  const compId = compIdEl ? parseInt(compIdEl.value || '0', 10) : 0;

  const data: any = {
    name,
    description: (document.getElementById('shopItemDesc') as HTMLTextAreaElement).value,
    price: +(document.getElementById('shopItemPrice') as HTMLInputElement).value || 0,
    quantity: +(document.getElementById('shopItemQty') as HTMLInputElement).value || 0,
    category: (document.getElementById('shopItemCat') as HTMLSelectElement).value,
    properties: '{}',
  };
  if (compId > 0) data.compendium_equipment_id = compId;

  try {
    if (itemId) {
      await api('PUT', `/api/shop-items/${itemId}`, data);
      toast('Item updated');
    } else {
      await api('POST', `/api/shops/${shopId}/items`, data);
      toast('Item added to shop');
    }
    hideModal();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('unlinkShopItem', async function (shopId: number, itemId: number) {
  try {
    await api('DELETE', `/api/shop-items/${itemId}/link`);
    toast('Unlinked from compendium (item kept)');
    (window as any).showShopDetail(shopId);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('editShopItem', async function (itemId: number) {
  // Find the item from any shop detail modal
  const itemsEl = document.getElementById('shopItemsList');
  if (!itemsEl) return;
  const itemRow = itemsEl.querySelector(`[onclick*="deleteShopItem(${itemId})"]`)?.closest('.inv-item');
  // We need to fetch the item data
  try {
    // Scan through shops to find this item
    for (const shop of selectedShopShops) {
      const items = await api('GET', `/api/shops/${shop.id}/items`);
      const item = items.find((i: any) => i.id === itemId);
      if (item) {
        showModal('Edit Shop Item', `
          <input type="hidden" id="shopItemCompId" value="${item.compendium_equipment_id || ''}">
          <div class="mb-3"><label class="form-label">Item Name</label><input class="form-control" id="shopItemName" value="${esc(item.name)}"></div>
          <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopItemDesc" rows="2">${esc(item.description)}</textarea></div>
          <div class="row g-3 mb-3">
            <div class="col-4"><label class="form-label">Price (gp)</label><input class="form-control" id="shopItemPrice" type="number" value="${item.price}" step="0.01"></div>
            <div class="col-4"><label class="form-label">Quantity</label><input class="form-control" id="shopItemQty" type="number" value="${item.quantity}" min="0"></div>
            <div class="col-4"><label class="form-label">Category</label>
              <select class="form-select" id="shopItemCat">
                <option value="">General</option>
                ${['weapon','armor','potion','scroll','gear','ammunition','food','other'].map(c =>
                  `<option value="${c}" ${item.category === c ? 'selected' : ''}>${capitalize(c)}</option>`
                ).join('')}
              </select></div>
          </div>
          <div class="d-flex gap-2 mb-2">
            <button class="btn btn-outline-gold btn-sm flex-grow-1" onclick="pickCompItem('shopItemName')"><i class="fa-solid fa-book me-1"></i>${item.compendium_equipment_id ? 'Relink from Compendium' : 'From Compendium'}</button>
          </div>
          <div id="shopItemLinkStatus" class="small text-muted mt-1">
            ${item.compendium_equipment_id ? '<span class="badge badge-compendium"><i class="fa-solid fa-book me-1"></i>Linked from compendium</span>' : ''}
          </div>
          <button class="btn btn-primary w-100 mt-2" onclick="saveShopItem(${shop.id}, ${itemId})"><i class="fa-solid fa-save me-1"></i>Save Changes</button>
        `);
        return;
      }
    }
    toast('Item not found', true);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('deleteShopItem', async function (itemId: number) {
  if (!confirm('Remove this item from the shop?')) return;
  try {
    await api('DELETE', `/api/shop-items/${itemId}`);
    toast('Item removed from shop');
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Buy / Sell ───

expose('buyItem', async function (shopId: number, itemId: number) {
  if (!currentChar) { toast('Select a character first', true); return; }

  showModal('Buy Item', `
    <div class="text-center mb-3">
      <p class="mb-1">How many would you like to buy?</p>
      <div class="input-group" style="max-width:200px;margin:0 auto">
        <button class="btn btn-outline-secondary" onclick="qtyAdjust(-1)">−</button>
        <input class="form-control text-center" id="buyQty" type="number" value="1" min="1" max="99">
        <button class="btn btn-outline-secondary" onclick="qtyAdjust(1)">+</button>
      </div>
    </div>
    <button class="btn btn-gold w-100" onclick="confirmBuy(${shopId}, ${itemId})"><i class="fa-solid fa-cart-shopping me-1"></i>Purchase</button>
  `);
});

expose('qtyAdjust', function (delta: number) {
  const input = document.getElementById('buyQty') as HTMLInputElement;
  if (input) {
    const val = Math.max(1, Math.min(99, (parseInt(input.value) || 1) + delta));
    input.value = String(val);
  }
});

expose('confirmBuy', async function (shopId: number, itemId: number) {
  const qty = parseInt((document.getElementById('buyQty') as HTMLInputElement).value) || 1;
  try {
    const result = await api('POST', `/api/shops/${shopId}/buy`, {
      item_id: itemId,
      quantity: qty,
      character_id: currentChar!.id,
    });
    hideModal();
    toast(`Purchased ${esc(result.item_name || 'item')} x${qty} for ${result.total_cost} gp`);
    // Refresh shop items to update stock
    if (typeof (window as any).showShopDetail === 'function') {
      // Re-open shop detail with updated data
      (window as any).showShopDetail(shopId);
    }
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('sellItem', async function (shopId: number) {
  if (!currentChar) { toast('Select a character first', true); return; }

  try {
    const char = await api('GET', `/api/characters/${currentChar.id}`);
    const inv = char.inventory || [];
    if (!inv.length) {
      toast('No items to sell', true);
      return;
    }

    // Get shop details for sell markup
    const shops = selectedShopShops.length ? selectedShopShops : await api('GET', '/api/shops');
    const shop = shops.find((s: any) => s.id === shopId);
    const sellPct = shop ? shop.markup_buy_percent : 50;

    showModal('Sell Items', `
      <p class="text-muted small mb-2">Shop pays ${sellPct}% of base price. Select items to sell.</p>
      <div id="sellItemsList" style="max-height:350px;overflow-y:auto">
        ${inv.map((i: any) => `
          <div class="inv-item sell-inv-item" data-id="${i.id}" data-price="${i.cost || i.price || 0}" style="cursor:pointer" onclick="toggleSellItem(this)">
            <div>
              <span class="fw-bold small">${esc(i.name)}</span>
              ${i.quantity > 1 ? `<span class="badge badge-muted ms-1">x${i.quantity}</span>` : ''}
              <div class="small text-muted">Est. ${Math.round((i.cost || i.price || 0) * sellPct / 100)} gp each</div>
            </div>
            <div class="d-flex align-items-center gap-2">
              <input type="number" class="form-control form-control-sm" style="width:60px" value="1" min="1" max="${i.quantity}" onclick="event.stopPropagation()">
              <span class="sell-check"><i class="fa-regular fa-square"></i></span>
            </div>
          </div>
        `).join('')}
      </div>
      <button class="btn btn-gold w-100 mt-3" onclick="confirmSell(${shopId})"><i class="fa-solid fa-sack-dollar me-1"></i>Sell Selected Items</button>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('toggleSellItem', function (el: HTMLElement) {
  el.classList.toggle('selected');
  const check = el.querySelector('.sell-check i')!;
  if (el.classList.contains('selected')) {
    check.className = 'fa-solid fa-square-check text-success';
  } else {
    check.className = 'fa-regular fa-square';
  }
});

expose('confirmSell', async function (shopId: number) {
  const selected: Array<{ inventory_item_id: number; quantity: number }> = [];
  document.querySelectorAll('.sell-inv-item.selected').forEach((el: any) => {
    const id = parseInt(el.getAttribute('data-id'));
    const qtyInput = el.querySelector('input[type="number"]') as HTMLInputElement;
    const qty = parseInt(qtyInput?.value || '1');
    selected.push({ inventory_item_id: id, quantity: qty });
  });

  if (!selected.length) { toast('Select items to sell', true); return; }

  try {
    // Sell each item sequentially (the API sells one item at a time)
    let totalGold = 0;
    for (const s of selected) {
      const result = await api('POST', `/api/shops/${shopId}/sell`, s);
      totalGold += result.gold_earned || 0;
    }
    hideModal();
    toast(`Sold ${selected.length} item(s) for ${totalGold} gp`);
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Transactions ───

expose('showTransactions', async function () {
  try {
    const txns = await api('GET', '/api/shop-transactions');
    showModal('Transaction History', `
      <div style="max-height:400px;overflow-y:auto">
        ${txns.length === 0 ? '<div class="text-muted text-center py-4 fst-italic">No transactions yet.</div>' :
          txns.map((t: any) => {
            const date = t.created_at ? new Date(t.created_at).toLocaleDateString() : '';
            const typeIcon = t.transaction_type === 'buy' ? 'fa-cart-shopping text-success' : 'fa-sack-dollar text-warning';
            return `<div class="inv-item">
              <div>
                <span class="fw-bold small">${esc(t.item_name || 'Unknown')}</span>
                <span class="badge ${t.transaction_type === 'buy' ? 'badge-blood' : 'badge-gold'} ms-1" style="font-size:0.6rem">${esc(t.transaction_type)}</span>
                <div class="small text-muted">${esc(t.shop_name || '')} ${t.character_name ? '· ' + esc(t.character_name) : ''}</div>
              </div>
              <div class="text-nowrap">
                <span class="fw-bold" style="color:var(--gold)">${t.total_cost || t.gold_earned || 0} gp</span>
                <span class="text-muted small ms-2">x${t.quantity || 1}</span>
                <div class="text-muted small">${date}</div>
              </div>
            </div>`;
          }).join('')}
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
});
