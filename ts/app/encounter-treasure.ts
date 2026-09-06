// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';

// ─── Encounter Difficulty Calculator ───

const CR_XP: Record<string, number> = {
  '0': 10, '1/8': 25, '1/4': 50, '1/2': 100, '1': 200, '2': 450, '3': 700,
  '4': 1100, '5': 1800, '6': 2300, '7': 2900, '8': 3900, '9': 5000, '10': 5900,
  '11': 7200, '12': 8400, '13': 10000, '14': 11500, '15': 13000, '16': 15000,
  '17': 18000, '18': 20000, '19': 22000, '20': 25000, '21': 33000, '22': 41000,
  '23': 50000, '24': 62000, '25': 75000, '30': 155000,
};

expose('showEncounterDifficulty', function () {
  showModal('Encounter Difficulty Calculator', `
    <div class="diff-calc-section">
      <h6>Party</h6>
      <div class="row g-2 mb-2">
        <div class="col-6"><label class="form-label"># Characters</label><input class="form-control" id="ecPartySize" type="number" value="4" min="1" max="10" oninput="calcEncounterDifficulty()"></div>
        <div class="col-6"><label class="form-label">Average Level</label><input class="form-control" id="ecAvgLevel" type="number" value="5" min="1" max="20" oninput="calcEncounterDifficulty()"></div>
      </div>
      <h6 class="mt-2">Monsters</h6>
      <div id="ecMonsterList"></div>
      <button class="btn btn-sm btn-outline-primary mt-1" onclick="addMonsterRow()"><i class="fa-solid fa-plus me-1"></i>Add Monster</button>
      <div id="ecResult" class="mt-3"></div>
    </div>
    <div class="text-center mt-2">
      <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
    </div>
  `);
  // Add first monster row
  addMonsterRow();
});

expose('addMonsterRow', function () {
  const list = document.getElementById('ecMonsterList');
  if (!list) return;
  const idx = list.children.length;
  const crOptions = Object.keys(CR_XP).map(cr => `<option value="${cr}">${cr}</option>`).join('');
  const row = document.createElement('div');
  row.className = 'row g-2 mb-1 align-items-center';
  row.innerHTML = `
    <div class="col-4"><input class="form-control form-control-sm" id="ecMonsterName${idx}" placeholder="Name"></div>
    <div class="col-3"><select class="form-select form-select-sm" id="ecMonsterCR${idx}" onchange="calcEncounterDifficulty()">${crOptions}</select></div>
    <div class="col-2"><input class="form-control form-control-sm" id="ecMonsterQty${idx}" type="number" value="1" min="1" oninput="calcEncounterDifficulty()"></div>
    <div class="col-3"><button class="btn btn-sm btn-outline-danger" onclick="this.closest('.row').remove();calcEncounterDifficulty()"><i class="fa-solid fa-xmark"></i></button></div>
  `;
  list.appendChild(row);
  calcEncounterDifficulty();
});

expose('calcEncounterDifficulty', function () {
  const partySize = parseInt((document.getElementById('ecPartySize') as HTMLInputElement)?.value) || 4;
  const avgLevel = parseInt((document.getElementById('ecAvgLevel') as HTMLInputElement)?.value) || 5;
  const resultEl = document.getElementById('ecResult');
  if (!resultEl) return;

  // Party XP thresholds (DMG)
  const thresholds = {
    easy: avgLevel * 25 * partySize,
    medium: avgLevel * 50 * partySize,
    hard: avgLevel * 75 * partySize,
    deadly: avgLevel * 100 * partySize,
  };

  // Sum monster XP
  const monsterList = document.getElementById('ecMonsterList');
  if (!monsterList) return;
  let totalXp = 0;
  let monsterCount = 0;
  const monsters: Array<{ name: string; cr: string; qty: number; xp: number }> = [];
  for (let i = 0; i < monsterList.children.length; i++) {
    const nameInput = document.getElementById(`ecMonsterName${i}`) as HTMLInputElement;
    const crSelect = document.getElementById(`ecMonsterCR${i}`) as HTMLSelectElement;
    const qtyInput = document.getElementById(`ecMonsterQty${i}`) as HTMLInputElement;
    if (nameInput && crSelect && qtyInput) {
      const cr = crSelect.value;
      const qty = parseInt(qtyInput.value) || 1;
      const xp = (CR_XP[cr] || 0) * qty;
      totalXp += xp;
      monsterCount += qty;
      monsters.push({ name: nameInput.value || `CR ${cr}`, cr, qty, xp });
    }
  }

  // Encounter multiplier
  let multiplier = 1;
  if (monsterCount >= 2) multiplier = 1.5;
  if (monsterCount >= 3) multiplier = 2;
  if (monsterCount >= 7) multiplier = 2.5;
  if (monsterCount >= 11) multiplier = 3;
  if (monsterCount >= 15) multiplier = 4;

  const adjustedXp = Math.round(totalXp * multiplier);

  // Determine difficulty
  let difficulty = 'easy';
  let badgeClass = 'diff-badge-easy';
  let pct = (adjustedXp / thresholds.deadly) * 100;
  if (adjustedXp >= thresholds.deadly) { difficulty = 'deadly'; badgeClass = 'diff-badge-deadly'; }
  else if (adjustedXp >= thresholds.hard) { difficulty = 'hard'; badgeClass = 'diff-badge-hard'; }
  else if (adjustedXp >= thresholds.medium) { difficulty = 'medium'; badgeClass = 'diff-badge-medium'; }
  pct = Math.min(100, pct);

  resultEl.innerHTML = `
    <div class="diff-meter position-relative" style="height:20px">
      <div class="diff-marker" style="left:${pct}%"></div>
    </div>
    <div class="d-flex justify-content-between small text-muted">
      <span>Easy (${thresholds.easy})</span>
      <span>Medium (${thresholds.medium})</span>
      <span>Hard (${thresholds.hard})</span>
      <span>Deadly (${thresholds.deadly})</span>
    </div>
    <div class="text-center mt-2">
      <span class="${badgeClass}">${difficulty.toUpperCase()}</span>
      <span class="ms-2 fw-bold">${adjustedXp.toLocaleString()} adjusted XP</span>
    </div>
    <div class="small text-muted mt-1">
      Total XP: ${totalXp.toLocaleString()} × ${multiplier} modifier
      ${monsterCount > 1 ? `(${monsterCount} monsters)` : ''}
      &middot; Per character: ${Math.round(adjustedXp / partySize).toLocaleString()} XP
    </div>
    ${monsters.filter(m => m.name).length ? `<div class="mt-2 small">${monsters.filter(m => m.name).map(m => `<div>${esc(m.name)} ×${m.qty} (${m.xp.toLocaleString()} XP)</div>`).join('')}</div>` : ''}
  `;
});

// ─── Treasure Generator ───

const TREASURE_TABLES: Record<string, Array<{ dice: string; coin: string; multiplier: number }>> = {
  easy: [
    { dice: '2d6', coin: 'CP', multiplier: 10 },
    { dice: '1d6', coin: 'SP', multiplier: 5 },
  ],
  medium: [
    { dice: '4d6', coin: 'CP', multiplier: 10 },
    { dice: '2d6', coin: 'SP', multiplier: 10 },
    { dice: '1d4', coin: 'GP', multiplier: 10 },
  ],
  hard: [
    { dice: '2d6', coin: 'CP', multiplier: 100 },
    { dice: '4d6', coin: 'SP', multiplier: 50 },
    { dice: '2d6', coin: 'GP', multiplier: 20 },
    { dice: '1d4', coin: 'PP', multiplier: 10 },
  ],
  deadly: [
    { dice: '4d6', coin: 'CP', multiplier: 100 },
    { dice: '6d6', coin: 'SP', multiplier: 100 },
    { dice: '4d6', coin: 'GP', multiplier: 100 },
    { dice: '2d6', coin: 'PP', multiplier: 20 },
  ],
};

const MAGIC_ITEMS: Record<string, string[]> = {
  common: ['Potion of Healing', 'Spell Scroll (Cantrip)', 'Cloak of Billowing', 'Candle of the Deep', 'Bag of Tricks (Grey)'],
  uncommon: ['Bag of Holding', 'Cloak of Protection', 'Boots of Striding', 'Wand of Magic Detection', 'Potion of Invisibility', '+1 Weapon'],
  rare: ['Flame Tongue', 'Cloak of Displacement', 'Ring of Protection', 'Belt of Hill Giant Strength', 'Potion of Greater Healing'],
  'very rare': ['Belt of Fire Giant Strength', 'Ring of Spell Turning', 'Cloak of Invisibility', 'Staff of the Magi', 'Potion of Supreme Healing'],
};

async function rollDice(dice: string): Promise<number> {
  try {
    const result = await api('POST', '/api/roll', { expression: dice });
    return result.total || 0;
  } catch {
    return 0;
  }
}

expose('showTreasureGenerator', function () {
  showModal('Treasure Generator', `
    <div class="diff-calc-section">
      <div class="row g-2 mb-2">
        <div class="col-6">
          <label class="form-label">Party Level</label>
          <select class="form-select" id="tgLevel">
            ${Array.from({length: 20}, (_, i) => `<option value="${i+1}" ${i+1 === 5 ? 'selected' : ''}>Level ${i+1}</option>`).join('')}
          </select>
        </div>
        <div class="col-6">
          <label class="form-label">Difficulty</label>
          <select class="form-select" id="tgDifficulty">
            <option value="easy">Easy</option>
            <option value="medium" selected>Medium</option>
            <option value="hard">Hard</option>
            <option value="deadly">Deadly</option>
          </select>
        </div>
      </div>
      <button class="btn btn-gold w-100" onclick="generateTreasure()"><i class="fa-solid fa-wand-sparkles me-1"></i>Generate Treasure</button>
      <div id="tgResult"></div>
    </div>
    <div class="text-center mt-2">
      <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
    </div>
  `);
});

expose('generateTreasure', async function () {
  const lvl = parseInt((document.getElementById('tgLevel') as HTMLSelectElement).value) || 5;
  const diff = (document.getElementById('tgDifficulty') as HTMLSelectElement).value;
  const resultEl = document.getElementById('tgResult');
  if (!resultEl) return;

  const table = TREASURE_TABLES[diff];
  const lines: string[] = [];
  let totalGp = 0;

  for (const entry of table) {
    const rolled = await rollDice(entry.dice);
    const amount = rolled * entry.multiplier;
    const line = `${rolled} × ${entry.multiplier} = ${amount.toLocaleString()} ${entry.coin}`;
    lines.push(line);

    // Convert to GP estimate
    const gpMultiplier: Record<string, number> = { CP: 0.01, SP: 0.1, EP: 0.5, GP: 1, PP: 10 };
    totalGp += amount * (gpMultiplier[entry.coin] || 0);
  }

  // Magic item tier based on level
  let magicTier = 'common';
  if (lvl >= 5) magicTier = 'uncommon';
  if (lvl >= 11) magicTier = 'rare';
  if (lvl >= 17) magicTier = 'very rare';

  const magicPool = MAGIC_ITEMS[magicTier] || [];
  const magicItem = magicPool[Math.floor(Math.random() * magicPool.length)];

  resultEl.innerHTML = `
    <div class="treasure-result">
      <div class="treasure-total">≈ ${totalGp.toLocaleString()} GP</div>
      ${lines.map(l => `<div class="treasure-line">${l}</div>`).join('')}
      <div class="treasure-line fw-bold mt-2">Magic Item: ${magicItem} (${magicTier})</div>
    </div>
    <button class="btn btn-sm btn-outline-primary mt-2 w-100" onclick="generateTreasure()"><i class="fa-solid fa-rotate me-1"></i>Generate Again</button>
  `;
});
