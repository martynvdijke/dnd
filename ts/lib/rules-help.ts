/**
 * Inline rules help — short SRD-plain-language explanations for sheet stats.
 * Static map, no backend; unknown keys fall back to a generic message.
 */
import { esc, showModal } from './dom';
import { expose } from './expose';

export interface RuleHelp {
  title: string;
  body: string;
}

export const RULES: Record<string, RuleHelp> = {
  ac: {
    title: 'Armor Class',
    body: 'How hard you are to hit. Unarmored: 10 + DEX modifier. Light armor: armor base + full DEX. Medium armor: armor base + DEX (max +2). Heavy armor: armor base only. A shield or any equipped item with an AC bonus adds on top. When no body armor is equipped, villum keeps your manual AC.',
  },
  initiative: {
    title: 'Initiative',
    body: 'Rolled once at the start of combat: 1d20 + DEX modifier. Higher results act first. The combat tracker rolls this for you.',
  },
  speed: {
    title: 'Speed',
    body: 'How far you can move on your turn, in feet. Some conditions (grappled, restrained, paralyzed, stunned, unconscious) and exhaustion can reduce it to 0.',
  },
  exhaustion: {
    title: 'Exhaustion',
    body: 'Six levels, each worse: 1 disadvantage on ability checks; 2 speed halved; 3 disadvantage on attacks and saves; 4 HP max halved; 5 speed 0; 6 death. Villum applies these to rolls and HP automatically. A long rest removes one level.',
  },
  conditions: {
    title: 'Conditions',
    body: 'Active conditions apply their mechanical effects automatically: blinded, frightened, poisoned and restrained impose disadvantage; prone and blinded grant attackers advantage; paralyzed, petrified, stunned and unconscious also zero your speed and auto-fail STR/DEX saves.',
  },
  advantage: {
    title: 'Advantage & Disadvantage',
    body: 'Advantage rolls two d20 and keeps the higher; disadvantage keeps the lower. One source of each cancels out and you roll a single d20.',
  },
  cover: {
    title: 'Cover & Range',
    body: 'Half cover adds +2 to the target AC, three-quarters cover adds +5, total cover means you cannot be targeted. Attacking beyond a weapon’s long range is impossible; beyond normal range imposes disadvantage. Enter these in the attack dialog.',
  },
  'death-saves': {
    title: 'Death Saves',
    body: 'At 0 HP, roll a d20 at the start of each turn. 10 or higher is a success, lower is a failure. A natural 20 restores 1 HP; a natural 1 counts as two failures. Three successes stabilise you, three failures kill you.',
  },
  'passive-perception': {
    title: 'Passive Perception',
    body: '10 + WIS modifier + proficiency bonus if you are proficient in Perception. Used to notice hidden creatures and details without actively searching.',
  },
  proficiency: {
    title: 'Proficiency Bonus',
    body: 'A bonus added to rolls you are proficient with — skills, saving throws, and attacks. It grows with your total level (from +2 up to +6).',
  },
};

export function getRulesHelp(key: string): RuleHelp {
  const known = RULES[key];
  if (known) return known;
  return {
    title: 'Rules Help',
    body: 'No rules explanation is available for this topic yet.',
  };
}

export function showRulesHelp(key: string): void {
  const help = getRulesHelp(key);
  showModal(help.title, `<p class="mb-0">${esc(help.body)}</p>`);
}

export function rulesHelpButton(key: string, label?: string): string {
  const aria = label || getRulesHelp(key).title;
  return `<button type="button" class="rules-help-btn" data-rule="${esc(key)}" onclick="showRulesHelp('${esc(key)}')" aria-label="Help: ${esc(aria)}" title="Help: ${esc(aria)}"><i class="fa-solid fa-circle-question" aria-hidden="true"></i></button>`;
}

expose('showRulesHelp', showRulesHelp);
