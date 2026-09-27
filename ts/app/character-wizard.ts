/**
 * Guided character builder — 4-step wizard that reuses the existing
 * POST /api/characters + /api/characters/:id/proficiencies endpoints.
 * Complements, rather than replaces, the quick 3-field newChar() modal.
 */
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { expose } from '../lib/expose';

const ABILITIES = ['str', 'dex', 'con', 'int', 'wis', 'cha'] as const;
type Ability = (typeof ABILITIES)[number];

const STANDARD_ARRAY = [15, 14, 13, 12, 10, 8];

const SKILLS = [
  'Acrobatics', 'Animal Handling', 'Arcana', 'Athletics', 'Deception', 'History',
  'Insight', 'Intimidation', 'Investigation', 'Medicine', 'Nature', 'Perception',
  'Performance', 'Persuasion', 'Religion', 'Sleight of Hand', 'Stealth', 'Survival',
];

const SAVES: Ability[] = ['str', 'dex', 'con', 'int', 'wis', 'cha'];

interface WizardState {
  step: number;
  name: string;
  race: string;
  class: string;
  background: string;
  level: number;
  method: 'array' | 'manual';
  abilities: Record<Ability, number>;
  skills: string[];
  saves: Ability[];
}

const STEP_TITLES = ['Identity', 'Abilities', 'Proficiencies', 'Review'];

function defaultState(): WizardState {
  const abilities = { str: 10, dex: 10, con: 10, int: 10, wis: 10, cha: 10 } as Record<Ability, number>;
  STANDARD_ARRAY.forEach((v, i) => { abilities[ABILITIES[i]] = v; });
  return {
    step: 0, name: '', race: '', class: '', background: '', level: 1,
    method: 'array', abilities, skills: [], saves: [],
  };
}

let state: WizardState = defaultState();

function mod(v: number): string {
  const m = Math.floor((v - 10) / 2);
  return (m >= 0 ? '+' : '') + m;
}

function slug(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, '-');
}

function val(id: string): string {
  return (document.getElementById(id) as HTMLInputElement | null)?.value || '';
}

/** Read the currently-visible step's DOM values back into state. */
function collect(): void {
  if (state.step === 0) {
    state.name = val('wizName');
    state.race = val('wizRace');
    state.class = val('wizClass');
    state.background = val('wizBackground');
    state.level = Math.min(20, Math.max(1, parseInt(val('wizLevel'), 10) || 1));
  } else if (state.step === 1) {
    for (const a of ABILITIES) {
      const v = parseInt(val('wizAbil-' + a), 10);
      if (!isNaN(v)) state.abilities[a] = v;
    }
  }
}

function nav(create: boolean): string {
  return `
    <div class="d-flex gap-2 mt-3">
      <button class="btn btn-outline-secondary" data-testid="wizard-back" onclick="wizardBack()"${state.step === 0 ? ' disabled' : ''}>Back</button>
      ${create
        ? '<button class="btn btn-primary flex-fill" data-testid="wizard-create" onclick="wizardCreate()"><i class="fa-solid fa-plus me-1"></i>Create Character</button>'
        : '<button class="btn btn-primary flex-fill" data-testid="wizard-next" onclick="wizardNext()">Next</button>'}
    </div>`;
}

function bodyFor(step: number): string {
  if (step === 0) {
    return `
      <div data-testid="character-wizard">
        <div class="mb-3"><label class="form-label">Name *</label><input class="form-control" id="wizName" value="${esc(state.name)}" placeholder="Character name"></div>
        <div class="row g-3 mb-3">
          <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="wizRace" list="wizRaceList" value="${esc(state.race)}"><datalist id="wizRaceList"></datalist></div>
          <div class="col-6"><label class="form-label">Class</label><input class="form-control" id="wizClass" list="wizClassList" value="${esc(state.class)}"><datalist id="wizClassList"></datalist></div>
        </div>
        <div class="row g-3 mb-3">
          <div class="col-6"><label class="form-label">Background</label><input class="form-control" id="wizBackground" list="wizBgList" value="${esc(state.background)}"><datalist id="wizBgList"></datalist></div>
          <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="wizLevel" type="number" min="1" max="20" value="${state.level}"></div>
        </div>
      </div>
      ${nav(false)}`;
  }
  if (step === 1) {
    const grid = ABILITIES.map(a => `
      <div class="col-4">
        <label class="form-label small text-uppercase">${a}</label>
        ${state.method === 'array'
          ? `<select class="form-select" id="wizAbil-${a}" onchange="wizardUpdateMods()">${STANDARD_ARRAY.map(v => `<option value="${v}"${state.abilities[a] === v ? ' selected' : ''}>${v}</option>`).join('')}</select>`
          : `<input class="form-control" id="wizAbil-${a}" type="number" min="1" max="30" value="${state.abilities[a]}" oninput="wizardUpdateMods()">`}
        <div class="small text-muted mt-1" id="wizMod-${a}">${mod(state.abilities[a])}</div>
      </div>`).join('');
    return `
      <div data-testid="character-wizard">
        <div class="btn-group mb-3" role="group" aria-label="Ability score method">
          <button type="button" class="btn btn-sm ${state.method === 'array' ? 'btn-primary' : 'btn-outline-primary'}" onclick="wizardSetMethod('array')">Standard Array</button>
          <button type="button" class="btn btn-sm ${state.method === 'manual' ? 'btn-primary' : 'btn-outline-primary'}" onclick="wizardSetMethod('manual')">Manual</button>
        </div>
        <div class="row g-2">${grid}</div>
      </div>
      ${nav(false)}`;
  }
  if (step === 2) {
    return `
      <div data-testid="character-wizard">
        <h6>Skills</h6>
        <div class="row g-1">${SKILLS.map(sk => `
          <div class="col-6 col-md-4"><div class="form-check">
            <input class="form-check-input" type="checkbox" id="wizSkill-${slug(sk)}"${state.skills.includes(sk) ? ' checked' : ''} onchange="wizardToggleSkill('${sk}')">
            <label class="form-check-label" for="wizSkill-${slug(sk)}">${sk}</label>
          </div></div>`).join('')}</div>
        <h6 class="mt-3">Saving Throws</h6>
        <div class="row g-1">${SAVES.map(a => `
          <div class="col-6 col-md-4"><div class="form-check">
            <input class="form-check-input" type="checkbox" id="wizSave-${a}"${state.saves.includes(a) ? ' checked' : ''} onchange="wizardToggleSave('${a}')">
            <label class="form-check-label" for="wizSave-${a}">${a.toUpperCase()}</label>
          </div></div>`).join('')}</div>
      </div>
      ${nav(false)}`;
  }
  return `
    <div data-testid="character-wizard">
      <p class="mb-1"><strong>${esc(state.name || 'Unnamed')}</strong> · Level ${state.level} ${esc(state.race)} ${esc(state.class)}${state.background ? ' · ' + esc(state.background) : ''}</p>
      <div class="small text-muted mb-2">${ABILITIES.map(a => `${a.toUpperCase()} ${state.abilities[a]} (${mod(state.abilities[a])})`).join(' · ')}</div>
      <div class="small text-muted mb-2">Skills: ${state.skills.length ? esc(state.skills.join(', ')) : 'none'}</div>
      <div class="small text-muted mb-2">Saves: ${state.saves.length ? esc(state.saves.map(s => s.toUpperCase()).join(', ')) : 'none'}</div>
    </div>
    ${nav(true)}`;
}

function render(): void {
  showModal(`Guided Builder — ${STEP_TITLES[state.step]} (Step ${state.step + 1} of 4)`, bodyFor(state.step));
  if (state.step === 0) void loadDatalists();
}

async function loadDatalists(): Promise<void> {
  const fill = (id: string, url: string) => {
    fetch(url, { credentials: 'include' })
      .then(r => r.json())
      .then((items: Array<{ name: string }>) => {
        const el = document.getElementById(id);
        if (el) el.innerHTML = items.map(i => `<option value="${esc(i.name)}">`).join('');
      })
      .catch(() => {});
  };
  fill('wizRaceList', '/api/compendium/races');
  fill('wizClassList', '/api/compendium/classes');
  fill('wizBgList', '/api/compendium/backgrounds');
}

export function newCharWizard(): void {
  state = defaultState();
  render();
}

export function wizardNext(): void {
  collect();
  if (state.step === 0 && !state.name.trim()) {
    toast('Name is required', true);
    return;
  }
  if (state.step < 3) state.step++;
  render();
}

export function wizardBack(): void {
  collect();
  if (state.step > 0) state.step--;
  render();
}

export function wizardSetMethod(method: 'array' | 'manual'): void {
  collect();
  state.method = method;
  render();
}

export function wizardUpdateMods(): void {
  collect();
  for (const a of ABILITIES) {
    const el = document.getElementById('wizMod-' + a);
    if (el) el.textContent = mod(state.abilities[a]);
  }
}

export function wizardToggleSkill(name: string): void {
  const i = state.skills.indexOf(name);
  if (i >= 0) state.skills.splice(i, 1);
  else state.skills.push(name);
}

export function wizardToggleSave(ability: string): void {
  const a = ability as Ability;
  const i = state.saves.indexOf(a);
  if (i >= 0) state.saves.splice(i, 1);
  else state.saves.push(a);
}

export async function wizardCreate(): Promise<void> {
  collect();
  if (!state.name.trim()) {
    toast('Name is required', true);
    return;
  }
  try {
    const char = await api<{ id: number }>('POST', '/api/characters', {
      name: state.name.trim(),
      race: state.race,
      class: state.class,
      background: state.background,
      level: state.level,
      ...state.abilities,
    });
    if (!char.id) throw new Error('Character creation failed');
    const profs: Array<{ type: string; name: string }> = [
      ...state.skills.map(name => ({ type: 'skill', name })),
      ...state.saves.map(a => ({ type: 'save', name: a })),
    ];
    for (const p of profs) {
      try {
        await api('POST', `/api/characters/${char.id}/proficiencies`, { character_id: char.id, ...p });
      } catch {
        toast(`Could not add ${p.type} ${p.name}`, true);
      }
    }
    hideModal();
    toast('Character created');
    await (window as unknown as Record<string, ((id: number) => Promise<void>) | undefined>)['openChar']?.(char.id);
  } catch (e) {
    toast((e as Error).message, true);
  }
}

expose('newCharWizard', newCharWizard);
expose('wizardNext', wizardNext);
expose('wizardBack', wizardBack);
expose('wizardSetMethod', wizardSetMethod);
expose('wizardUpdateMods', wizardUpdateMods);
expose('wizardToggleSkill', wizardToggleSkill);
expose('wizardToggleSave', wizardToggleSave);
expose('wizardCreate', wizardCreate);
