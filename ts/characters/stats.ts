/**
 * Character stats rendering — ability scores, skills, XP bar.
 * Extracted from app.ts (address-tech-debt-and-ux). Replaces the earlier
 * partial port; app.ts versions are authoritative.
 */
import { expose } from '../lib/expose';
import { currentChar, currentUser } from '../lib/state';
import { api } from '../lib/api';
import { esc } from '../lib/dom';
import { rulesHelpButton } from '../lib/rules-help';

export const XP_TABLE = [0, 300, 900, 2700, 6500, 14000, 23000, 34000, 48000, 64000, 85000, 100000, 120000, 140000, 165000, 195000, 225000, 265000, 305000, 355000];

type AbilityKey = 'str' | 'dex' | 'con' | 'int' | 'wis' | 'cha';

interface Proficiency { id: number; name: string; type: string }

interface CharStats extends Record<string, unknown> {
  str?: number; dex?: number; con?: number; int?: number; wis?: number; cha?: number;
  str_mod?: number; dex_mod?: number; con_mod?: number; int_mod?: number; wis_mod?: number; cha_mod?: number;
  proficiency_bonus: number;
  inspiration: number;
  passive_perception: number;
  xp: number;
  level: number;
  exhaustion_level?: number;
  proficiencies?: Proficiency[];
}

// ─── Stats ───

export function renderStats(): void {
  const c = currentChar as CharStats | null;
  if (!c) return;
  const el = document.getElementById('statsSection');
  if (!el) return;
  const abils = (['str','dex','con','int','wis','cha'] as AbilityKey[]).map((k, i) => ({ key: k, label: k.toUpperCase(), desc: ['Strength','Dexterity','Constitution','Intelligence','Wisdom','Charisma'][i], skills: ['Athletics','Acrobatics, Sleight of Hand, Stealth','','Arcana, History, Investigation, Nature, Religion','Animal Handling, Insight, Medicine, Perception, Survival','Deception, Intimidation, Performance, Persuasion'][i] }));
  el.innerHTML = `
    <div class="stats-grid row g-3">
      ${abils.map(a => {
        const val = c[a.key] as number | undefined ?? 0;
        const mod = c[`${a.key}_mod`] as number | undefined ?? 0;
        const cls = mod > 0 ? 'text-success' : mod < 0 ? 'text-danger' : 'text-muted';
        return `<div class="ability-column col-6 col-md-4 col-lg-2">
          <div class="ability-box" title="${a.desc} (${a.label})\nModifier: ${mod >= 0 ? '+' : ''}${mod}\nSkills: ${a.skills || 'None'}">
            <div class="abil-label" onclick="rollCheck('check','${a.key}','normal')" style="cursor:pointer">${a.label}</div>
            ${window.renderStepper(a.key, val, 1, 1, 30, a.label)}
            <div class="abil-mod ${cls}">${mod >= 0 ? '+' : ''}${mod}</div>
          </div>
        </div>`;
      }).join('')}
    </div>
    <div class="d-flex gap-2 mt-3">
      <button class="btn btn-sm btn-outline-primary" onclick="rollCheck('check','str','advantage')"><i class="fa-solid fa-chevron-up me-1"></i>Advantage</button>
      <button class="btn btn-sm btn-outline-primary" onclick="rollCheck('check','str','disadvantage')"><i class="fa-solid fa-chevron-down me-1"></i>Disadvantage</button>
    </div>
    <div class="ornament my-2">✧</div>
    <div class="row g-3 mt-2">
      <div class="col-6 col-md-3"><label class="form-label">Proficiency</label>${window.renderStepper('proficiency_bonus', c.proficiency_bonus, 1, 1, 10, 'Proficiency Bonus')}</div>
      <div class="col-6 col-md-3"><label class="form-label">Inspiration</label>${window.renderStepper('inspiration', c.inspiration, 1, 0, undefined, 'Inspiration')}</div>
      <div class="col-6 col-md-3"><label class="form-label">Passive Percep.</label><input type="number" class="form-control form-control-sm" value="${c.passive_perception}" oninput="autoSaveField('passive_perception',this)" style="min-height:44px;font-size:1.1rem"></div>
      <div class="col-6 col-md-3"><label class="form-label">XP</label>${window.renderStepper('xp', c.xp, 1, 0, undefined, 'XP')}</div>
    </div>
    <div class="mt-2" id="xpBarContainer">${renderXPBar(c)}</div>
    <!-- Passive Investigation & Insight -->
    <div class="row g-2 mt-2">
      <div class="col-4 col-md-2">
        <div class="passive-score-box" title="10 + WIS modifier + proficiency if proficient">
          <div class="score-value">${(c.wis_mod||0) + ((c.proficiencies||[]).some((p) => p.name==='Insight')?c.proficiency_bonus:0) + 10}</div>
          <div class="score-label">Passive Insight</div>
          <div class="score-breakdown">10 + ${c.wis_mod||0} WIS${(c.proficiencies||[]).some((p) => p.name==='Insight')?' + '+c.proficiency_bonus+' Prof':''}</div>
        </div>
      </div>
      <div class="col-4 col-md-2">
        <div class="passive-score-box" title="10 + INT modifier + proficiency if proficient">
          <div class="score-value">${(c.int_mod||0) + ((c.proficiencies||[]).some((p) => p.name==='Investigation')?c.proficiency_bonus:0) + 10}</div>
          <div class="score-label">Passive Investigation</div>
          <div class="score-breakdown">10 + ${c.int_mod||0} INT${(c.proficiencies||[]).some((p) => p.name==='Investigation')?' + '+c.proficiency_bonus+' Prof':''}</div>
        </div>
      </div>
      <div class="col-4 col-md-3">
        <div class="exhaustion-display">
          <span class="exhaustion-level ex-${c.exhaustion_level||0}">${c.exhaustion_level||0}</span>
          <div>
            <div class="exhaustion-label">Exhaustion ${rulesHelpButton('exhaustion')}</div>
            <div class="exhaustion-effect">${(['-','Disadvantage on ability checks','Speed halved','Disadvantage on attacks & saves','HP max halved','Speed reduced to 0','Death'] as string[])[c.exhaustion_level||0]||''}</div>
          </div>
          <div class="ms-auto d-flex gap-1">
            <button class="exhaustion-btn" onclick="adjustExhaustion(-1)" title="Reduce exhaustion" style="min-width:44px;min-height:44px">−</button>
            <button class="exhaustion-btn" onclick="adjustExhaustion(1)" title="Increase exhaustion" style="min-width:44px;min-height:44px">+</button>
          </div>
        </div>
      </div>
    </div>
    <div class="row g-2 mt-2">
      <div class="col-12 col-md-4">
        <label class="form-label small">Damage Resistances</label>
        <input class="form-control form-control-sm" data-testid="defense-resistances" value="${esc((c.damage_resistances as string)||'')}" oninput="autoSaveField('damage_resistances',this)" placeholder="fire, cold">
      </div>
      <div class="col-12 col-md-4">
        <label class="form-label small">Damage Vulnerabilities</label>
        <input class="form-control form-control-sm" data-testid="defense-vulnerabilities" value="${esc((c.damage_vulnerabilities as string)||'')}" oninput="autoSaveField('damage_vulnerabilities',this)" placeholder="radiant">
      </div>
      <div class="col-12 col-md-4">
        <label class="form-label small">Damage Immunities</label>
        <input class="form-control form-control-sm" data-testid="defense-immunities" value="${esc((c.damage_immunities as string)||'')}" oninput="autoSaveField('damage_immunities',this)" placeholder="poison">
      </div>
    </div>
    <h5 class="mt-3">Skills <small class="text-muted fw-normal">(click to roll)</small></h5>
    <div id="skillsArea"><div class="skills-grid">${renderSkills(c)}</div></div>
    <h5 class="mt-3">Proficiencies</h5>
    <div id="profsArea">${(c.proficiencies||[]).map((p) =>
      `<span class="badge badge-blood me-1 mb-1">${esc(p.name)} (${p.type}) <a href="#" onclick="deleteProf(${p.id});return false" class="text-white text-decoration-none">×</a></span>`
    ).join('')}</div>
    <button class="btn btn-sm btn-outline-primary mt-2" onclick="addProf()"><i class="fa-solid fa-plus me-1"></i>Add Proficiency</button>
  `;
  void renderMemberships(el);
}

interface CampaignOption { id: number; name: string; my_role?: string }

/**
 * Campaign memberships panel. Owners (and admins) can change memberships;
 * everyone else sees a read-only list. Shared characters never expose controls.
 */
export async function renderMemberships(el: HTMLElement): Promise<void> {
  const c = currentChar as any;
  if (!c) return;
  const cu = currentUser as any;
  const canManage = (!!cu && c.user_id === cu.id) || cu?.role === 'admin';
  const editable = canManage && (window as any).canEditCharacter !== false;
  const section = document.createElement('div');
  section.className = 'mt-4';
  section.setAttribute('data-testid', 'character-campaigns-section');
  el.appendChild(section);

  const refs: { id: number; name: string }[] = c.campaigns || [];
  if (!editable) {
    section.innerHTML = `<h5>Campaigns</h5>` + (refs.length
      ? `<div class="d-flex flex-wrap gap-1">${refs.map(r => `<span class="badge bg-dark border border-secondary" data-testid="character-campaign-badge">${esc(r.name)}</span>`).join('')}</div>`
      : '<div class="small text-muted">No campaigns</div>');
    return;
  }

  section.innerHTML = `<h5>Campaigns</h5>
    <div class="small text-muted mb-1">Membership changes are saved with the character.</div>
    <div id="membershipOptions"><div class="small text-muted">Loading campaigns...</div></div>`;
  const optionsEl = section.querySelector('#membershipOptions') as HTMLElement | null;
  if (!optionsEl) return;
  try {
    const options: CampaignOption[] = await api('GET', '/api/campaigns/mine');
    const selected = new Set<number>(refs.map(r => r.id));
    optionsEl.innerHTML = options.length
      ? options.map(o => `
          <div class="form-check">
            <input class="form-check-input" type="checkbox" id="membershipCampaign-${o.id}"${selected.has(o.id) ? ' checked' : ''} data-testid="membership-campaign-option" onchange="toggleCharacterMembership(${o.id})">
            <label class="form-check-label" for="membershipCampaign-${o.id}">${esc(o.name)}${o.my_role === 'dm' ? ' <span class="badge bg-secondary">DM</span>' : ''}</label>
          </div>`).join('')
      : '<div class="small text-muted">You are not a member of any campaign yet.</div>';
  } catch {
    optionsEl.innerHTML = '<div class="small text-muted">Could not load campaigns.</div>';
  }
}

/** Toggle a campaign membership checkbox and queue a save. */
export function toggleCharacterMembership(campaignId: number): void {
  const c = currentChar as any;
  if (!c) return;
  const ids: number[] = Array.isArray(c.campaign_ids)
    ? c.campaign_ids
    : (c.campaigns || []).map((r: { id: number }) => r.id);
  const i = ids.indexOf(campaignId);
  if (i >= 0) ids.splice(i, 1);
  else ids.push(campaignId);
  c.campaign_ids = ids;
  (window as any).markDirty?.();
}

export function renderSkills(c: CharStats | Record<string, unknown>): string {
  const skls = [
    {name:'Athletics',abil:'str'},{name:'Acrobatics',abil:'dex'},{name:'Sleight of Hand',abil:'dex'},{name:'Stealth',abil:'dex'},
    {name:'Arcana',abil:'int'},{name:'History',abil:'int'},{name:'Investigation',abil:'int'},{name:'Nature',abil:'int'},{name:'Religion',abil:'int'},
    {name:'Animal Handling',abil:'wis'},{name:'Insight',abil:'wis'},{name:'Medicine',abil:'wis'},{name:'Perception',abil:'wis'},{name:'Survival',abil:'wis'},
    {name:'Deception',abil:'cha'},{name:'Intimidation',abil:'cha'},{name:'Performance',abil:'cha'},{name:'Persuasion',abil:'cha'},
  ];
  const rawProfs = (c['proficiencies'] as Proficiency[] | undefined) || [];
  const profs = rawProfs.filter((p) => p.type === 'skill').map((p) => p.name.toLowerCase());
  return skls.map(s => {
    const isProf = profs.includes(s.name.toLowerCase());
    const mod = (c[`${s.abil}_mod`] as number | undefined) ?? 0;
    const proficiencyBonus = c['proficiency_bonus'] as number | undefined ?? 0;
    const total = isProf ? mod + proficiencyBonus : mod;
    const sign = total >= 0 ? '+' : '';
    const breakdown = isProf ? `${s.abil.toUpperCase()} ${mod >= 0 ? '+' : ''}${mod} + Prof ${proficiencyBonus} = ${sign}${total}` : `${s.abil.toUpperCase()} ${mod >= 0 ? '+' : ''}${mod} = ${sign}${total}`;
    return `<div class="skill-row d-flex justify-content-between" onclick="rollCheck('skill','${s.name}','normal')" title="${breakdown}">
      <span class="skill-name">${s.name}${isProf ? ' <span class="text-primary">★</span>' : ''}</span>
      <span class="fw-bold">${sign}${total}</span>
    </div>`;
  }).join('');
}

// ─── XP Progress Bar ───

export function renderXPBar(c: CharStats | Record<string, unknown>): string {
  const level = (c['level'] as number | undefined) || 1;
  const xp = (c['xp'] as number | undefined) || 0;
  const idx = Math.min(level - 1, XP_TABLE.length - 2);
  const currentMilestone = XP_TABLE[idx];
  const nextMilestone = XP_TABLE[idx + 1] || currentMilestone + 10000;
  if (level >= 20) {
    return `<div class="small text-muted fst-italic">Maximum level reached</div>`;
  }
  const progress = nextMilestone > currentMilestone ? Math.min(100, Math.max(0, ((xp - currentMilestone) / (nextMilestone - currentMilestone)) * 100)) : 0;
  return `
    <div class="d-flex justify-content-between small mb-1">
      <span class="text-muted">Level ${level}</span>
      <span class="text-muted">${xp.toLocaleString()} / ${nextMilestone.toLocaleString()} XP</span>
      <span class="text-muted">Level ${level + 1}</span>
    </div>
    <div class="hp-bar" style="height:8px" title="${Math.round(progress)}% to next level">
      <div class="hp-bar-fill" style="width:${progress}%;height:100%;background:linear-gradient(90deg,var(--gold),var(--gold-light))"></div>
    </div>`;
}

// Window registrations (centralized)
expose('renderStats', renderStats);
expose('renderSkills', renderSkills);
expose('renderXPBar', renderXPBar);
expose('toggleCharacterMembership', toggleCharacterMembership);
