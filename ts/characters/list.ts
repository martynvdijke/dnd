/**
 * Character list view — browse, search, and select characters.
 *
 * Manages the character grid display and client-side filtering.
 * Functions are exposed globally for inline HTML onclick handlers.
 */

import { esc, toast } from '../lib/dom';
import { api } from '../lib/api';
import { expose } from '../lib/expose';

export async function loadCharacters() {
  try {
    const chars = await api('GET', '/api/characters');
    const grid = document.getElementById('charGrid')!;
    grid.innerHTML = chars.map((c: any) => `
      <div class="col-md-6 col-lg-4">
        <div class="character-card" onclick="openChar(${c.id})">
          <div class="d-flex align-items-center gap-2 mb-1">
            ${c.portrait_url ? `<img src="${esc(c.portrait_url)}" class="character-portrait" style="width:32px;height:32px;object-fit:cover;border-radius:50%" alt="">` : ''}
            <div class="char-name">${esc(c.name)}</div>
          </div>
          <div class="char-detail">
            ${c.race_color ? `<span class="badge" style="background:${c.race_color};color:#fff">${esc(c.race)}</span>` : esc(c.race)}
            ${esc(c.class)} · Level ${c.level}
          </div>
          <div class="char-hp mt-1">HP: ${c.hp_current}/${c.hp_max}</div>
        </div>
      </div>
    `).join('');
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('loadCharacters', loadCharacters);

export function filterCharacters() {
  const q = (document.getElementById('charSearch') as HTMLInputElement)?.value?.toLowerCase() || '';
  document.querySelectorAll('#charGrid .character-card').forEach(card => {
    const parent = card.closest('.col-md-6') as HTMLElement;
    if (parent) {
      parent.style.display = !q || card.textContent?.toLowerCase().includes(q) ? '' : 'none';
    }
  });
}
expose('filterCharacters', filterCharacters);
