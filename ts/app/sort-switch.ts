// @ts-nocheck — split from monolith, pre-existing type errors
import { expose } from '../lib/expose';
import { currentChar, currentCampaign, setCurrentChar } from '../lib/state';
import { refreshChar } from '../lib/refresh';
import { renderError } from '../lib/errors';
import { renderSheet } from '../characters/sheet';
import { setCurrentTab } from '../lib/state';
import { showView } from '../navigation';
import { loadCharacters } from '../characters/list';

function sortList(key: string, order: 'asc' | 'desc' = 'asc') {
  const container = document.getElementById(key + 'List');
  if (!container) return;
  const items = Array.from(container.querySelectorAll('.inv-item'));
  const sorted = items.sort((a, b) => {
    const va = a.getAttribute('data-sort') || a.textContent?.trim() || '';
    const vb = b.getAttribute('data-sort') || b.textContent?.trim() || '';
    const na = parseFloat(va), nb = parseFloat(vb);
    if (!isNaN(na) && !isNaN(nb)) return order === 'asc' ? na - nb : nb - na;
    return order === 'asc' ? va.localeCompare(vb) : vb.localeCompare(va);
  });
  sorted.forEach(item => container.appendChild(item));
}
expose('sortList', sortList);

async function openChar(id: number) {
  try {
    await refreshChar(id);
    expose('currentChar', currentChar);
    expose('canEditCharacter', !!(currentChar as any).can_edit);
    setCurrentTab('stats');
    showView('sheet');
    renderSheet();
  } catch (e: any) {
    renderError(e);
  }
}
expose('openChar', openChar);

expose('switchCampaign', function () {
  setCurrentChar(null);
  (window as any).loadCampaignPicker();
});

expose('switchCharacter', function () {
  if (!currentCampaign) {
    (window as any).loadCampaignPicker();
    return;
  }
  (window as any).loadCharacterPicker(currentCampaign.id);
});

expose('getCurrentCampaign', () => currentCampaign);

export { sortList, openChar };
