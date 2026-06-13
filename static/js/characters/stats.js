/**
 * Character stats rendering — ability scores, skills, saves, XP bar.
 */
import { esc } from '../lib/dom';
import { currentChar } from '../app';
export const XP_TABLE = [0, 300, 900, 2700, 6500, 14000, 23000, 34000, 48000, 64000, 85000, 100000, 120000, 140000, 165000, 195000, 225000, 265000, 305000, 355000];
export function renderStats() {
    const el = document.getElementById('statsSection');
    if (!el || !currentChar)
        return;
    const c = currentChar;
    const abi = ['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma'];
    const abiLabels = { strength: 'STR', dexterity: 'DEX', constitution: 'CON', intelligence: 'INT', wisdom: 'WIS', charisma: 'CHA' };
    el.innerHTML = `
    <div class="row g-2 mb-3">
      ${abi.map(a => {
        const val = c[a] || 10;
        const mod = Math.floor((val - 10) / 2);
        const modStr = mod >= 0 ? `+${mod}` : `${mod}`;
        return `<div class="col-4 col-md-2 stat-block text-center">
          <div class="stat-label">${abiLabels[a]}</div>
          <div class="stat-value">${val}</div>
          <div class="stat-mod">${modStr}</div>
          <button class="btn btn-sm btn-outline-gold mt-1 py-0 px-1" onclick="rollCheck('ability','${a}','normal')"><i class="fa-solid fa-dice"></i></button>
        </div>`;
    }).join('')}
    </div>
    <div id="skillsContainer" class="mb-3"></div>
    <div id="xpBarContainer"></div>`;
    renderSkills(c);
    renderXPBar(c);
}
export function renderSkills(c) {
    const container = document.getElementById('skillsContainer');
    if (!container)
        return;
    const skillNames = ['acrobatics', 'animal_handling', 'arcana', 'athletics', 'deception', 'history', 'insight', 'intimidation', 'investigation', 'medicine', 'nature', 'perception', 'performance', 'persuasion', 'religion', 'sleight_of_hand', 'stealth', 'survival'];
    const skillLabels = { acrobatics: 'Acrobatics', animal_handling: 'Animal Handling', arcana: 'Arcana', athletics: 'Athletics', deception: 'Deception', history: 'History', insight: 'Insight', intimidation: 'Intimidation', investigation: 'Investigation', medicine: 'Medicine', nature: 'Nature', perception: 'Perception', performance: 'Performance', persuasion: 'Persuasion', religion: 'Religion', sleight_of_hand: 'Sleight of Hand', stealth: 'Stealth', survival: 'Survival' };
    const skillAbi = { acrobatics: 'dexterity', animal_handling: 'wisdom', arcana: 'intelligence', athletics: 'strength', deception: 'charisma', history: 'intelligence', insight: 'wisdom', intimidation: 'charisma', investigation: 'intelligence', medicine: 'wisdom', nature: 'intelligence', perception: 'wisdom', performance: 'charisma', persuasion: 'charisma', religion: 'intelligence', sleight_of_hand: 'dexterity', stealth: 'dexterity', survival: 'wisdom' };
    let proficiencyBonus = 2;
    if (c.level >= 17)
        proficiencyBonus = 6;
    else if (c.level >= 13)
        proficiencyBonus = 5;
    else if (c.level >= 9)
        proficiencyBonus = 4;
    else if (c.level >= 5)
        proficiencyBonus = 3;
    container.innerHTML = `<table class="table table-sm skill-table">
    <thead><tr><th>Skill</th><th>Ability</th><th>Bonus</th><th class="text-center">Prof?</th></tr></thead>
    <tbody>${skillNames.map(s => {
        const abi = skillAbi[s];
        const abiVal = c[abi] || 10;
        const abiMod = Math.floor((abiVal - 10) / 2);
        const prof = c.proficiencies?.includes(s) || c.proficiencies?.includes(skillLabels[s]) || false;
        const expertise = c.expertise?.includes(s) || false;
        const bonus = abiMod + (expertise ? proficiencyBonus * 2 : prof ? proficiencyBonus : 0);
        const bonusStr = bonus >= 0 ? `+${bonus}` : `${bonus}`;
        return `<tr><td>${esc(skillLabels[s])}</td><td class="text-muted">${skillAbi[s].substring(0, 3).toUpperCase()}</td><td class="fw-bold">${bonusStr}</td><td class="text-center">${expertise ? '<i class="fa-solid fa-star text-gold"></i>' : prof ? '<i class="fa-solid fa-check text-success"></i>' : ''}</td></tr>`;
    }).join('')}</tbody>
  </table>`;
}
export function renderXPBar(c) {
    const container = document.getElementById('xpBarContainer');
    if (!container)
        return;
    const lv = c.level || 1;
    const xp = c.xp || 0;
    const nextLv = lv >= 20 ? '—' : XP_TABLE[lv];
    const prevLv = lv > 1 ? XP_TABLE[lv - 1] : 0;
    const pct = nextLv === '—' ? 100 : ((xp - prevLv) / (nextLv - prevLv)) * 100;
    container.innerHTML = `
    <div class="d-flex justify-content-between small">
      <span>Level ${lv}</span><span>${xp} XP</span><span>Next: ${nextLv}</span>
    </div>
    <div class="xp-bar"><div class="xp-bar-fill" style="width:${Math.min(100, Math.max(0, pct))}%"></div></div>`;
}
