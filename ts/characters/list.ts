/**
 * Character list view — browse, search, and select characters.
 *
 * Manages the character grid display and client-side filtering.
 * Functions are exposed globally for inline HTML onclick handlers.
 */

import { esc, toast } from '../lib/dom';
import { api } from '../lib/api';

export async function loadCharacters() {
  try {
    const chars = await api('GET', '/api/characters');
    const grid = document.getElementById('charGrid')!;
    grid.innerHTML = chars.map((c: any) => `
      <div class="col-md-6 col-lg-4">
        <div class="character-card" onclick="openChar(${c.id})">
          <div class="char-name">${esc(c.name)}</div>
          <div class="char-detail">${esc(c.race)} ${esc(c.class)} · Level ${c.level}</div>
          <div class="char-hp mt-1">HP: ${c.hp_current}/${c.hp_max}</div>
        </div>
      </div>
    `).join('');
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).loadCharacters = loadCharacters;

export function filterCharacters() {
  const q = (document.getElementById('charSearch') as HTMLInputElement)?.value?.toLowerCase() || '';
  document.querySelectorAll('#charGrid .character-card').forEach(card => {
    const parent = card.closest('.col-md-6') as HTMLElement;
    if (parent) {
      parent.style.display = !q || card.textContent?.toLowerCase().includes(q) ? '' : 'none';
    }
  });
}
(window as any).filterCharacters = filterCharacters;
