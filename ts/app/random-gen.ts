// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { currentChar, setCurrentChar } from '../lib/state';
import { renderSheet } from '../characters/sheet';

// ─── HP Auto-Calc in details ───

expose('calcHP', async function () {
  if (!currentChar) return;
  try {
    const result = await api('POST', `/api/characters/${currentChar.id}/calc-hp`);
    setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
    renderSheet();
    toast(`HP calculated: ${result.hp_max} HP`);
  } catch (e: any) { toast(e.message, true); }
});

// ─── Random Character Generator ───

expose('generateRandomChar', async function () {
  try {
    const rc = await api('GET', '/api/generate/character');
    showModal('Random Character', `
      <div class="text-center mb-3">
        <span class="fw-bold fs-5">${esc(rc.name)}</span>
      </div>
      <div class="row g-2 mb-3">
        <div class="col-6"><span class="text-muted">Race:</span> ${esc(rc.race)}</div>
        <div class="col-6"><span class="text-muted">Class:</span> ${esc(rc.class)}</div>
        <div class="col-6"><span class="text-muted">Level:</span> ${rc.level}</div>
        <div class="col-6"><span class="text-muted">Background:</span> ${esc(rc.background)}</div>
        <div class="col-6"><span class="text-muted">Alignment:</span> ${esc(rc.alignment)}</div>
        <div class="col-6"><span class="text-muted">Personality:</span> ${esc(rc.personality)}</div>
      </div>
      <div class="row g-2 mb-3">
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">STR</div><div class="stat-value">${rc.str}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">DEX</div><div class="stat-value">${rc.dex}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">CON</div><div class="stat-value">${rc.con}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">INT</div><div class="stat-value">${rc.int}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">WIS</div><div class="stat-value">${rc.wis}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">CHA</div><div class="stat-value">${rc.cha}</div></div></div>
      </div>
      <div class="mb-2"><span class="text-muted small">Quirk:</span> <span class="small">${esc(rc.quirk)}</span></div>
      <div><span class="text-muted small">Backstory Hook:</span> <span class="small fst-italic">${esc(rc.backstory_hook)}</span></div>
      <hr>
      <p class="small text-muted">Use this as inspiration for your next character!</p>
    `);
  } catch (e: any) { toast(e.message, true); }
});

// ─── Character Comparison ───

expose('showComparison', async function () {
  const sel = document.getElementById('charCompareSelect') as HTMLSelectElement;
  if (!sel) return;
  const selected = Array.from(sel.selectedOptions).map(o => o.value).filter(v => v);
  if (selected.length < 2) { toast('Select at least 2 characters', true); return; }
  try {
    const chars = await api('GET', `/api/characters/compare?ids=${selected.join(',')}`);
    showModal('Character Comparison', `
      <div class="table-responsive"><table class="table table-sm table-bordered">
        <thead><tr><th></th>${chars.map((c: any) => `<th class="text-center">${esc(c.name)}</th>`).join('')}</tr></thead>
        <tbody>
          ${[['Race','race'],['Class','class'],['Level','level'],['Background','background'],['Alignment','alignment'],
             ['HP','hp_current + "/" + hp_max'],['AC','ac'],['Speed','speed'],['Initiative','initiative'],
             ['STR','str'],['DEX','dex'],['CON','con'],['INT','int'],['WIS','wis'],['CHA','cha'],['XP','xp']].map(([label, field]) => `
            <tr><td class="fw-bold">${label}</td>
              ${chars.map((c: any) => {
                if (field === 'hp_current + "/" + hp_max') {
                  return `<td class="text-center">${c.hp_current}/${c.hp_max}</td>`;
                }
                return `<td class="text-center">${c[field] ?? '-'}</td>`;
              }).join('')}
            </tr>`).join('')}
        </tbody>
      </table></div>
    `);
  } catch (e: any) { toast(e.message, true); }
});

// ─── Add character comparison to character list view ───

let compareMode = false;

expose('toggleCompareMode', function () {
  compareMode = !compareMode;
  const el = document.getElementById('charGrid')!;
  const btn = document.getElementById('compareBtn') as HTMLButtonElement;
  if (compareMode) {
    el.querySelectorAll('.character-card').forEach(card => card.classList.add('compare-selectable'));
    // Add compare bar
    let bar = document.getElementById('compareBar');
    if (!bar) {
      bar = document.createElement('div');
      bar.id = 'compareBar';
      bar.className = 'd-flex align-items-center gap-2 p-2 mb-2 border rounded';
      bar.style.background = 'var(--parchment)';
      bar.innerHTML = `
        <span class="small fw-bold me-2">Compare:</span>
        <select multiple class="form-select form-select-sm" id="charCompareSelect" style="height:2rem;width:auto;min-width:200px"></select>
        <button class="btn btn-sm btn-gold" onclick="showComparison()"><i class="fa-solid fa-arrow-right me-1"></i>Compare</button>
        <button class="btn btn-sm btn-outline-secondary" onclick="toggleCompareMode()">Done</button>`;
      document.getElementById('charactersView')?.insertBefore(bar, document.getElementById('charGrid'));
    }
    // Populate select
    const select = document.getElementById('charCompareSelect') as HTMLSelectElement;
    select.innerHTML = '';
    document.querySelectorAll('#charGrid .character-card').forEach(card => {
      const id = card.getAttribute('onclick')?.match(/\d+/)?.[0];
      const name = card.querySelector('.char-name')?.textContent;
      if (id && name) {
        select.innerHTML += `<option value="${id}">${esc(name)}</option>`;
      }
    });
    if (btn) btn.textContent = 'Cancel Compare';
  } else {
    el.querySelectorAll('.character-card').forEach(card => card.classList.remove('compare-selectable'));
    const bar = document.getElementById('compareBar');
    if (bar) bar.remove();
    if (btn) btn.textContent = 'Compare';
  }
});
