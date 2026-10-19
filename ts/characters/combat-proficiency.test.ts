import { describe, expect, it, vi } from 'vitest';

vi.mock('../lib/api', () => ({ api: vi.fn() }));
vi.mock('../lib/state', () => ({ currentChar: null, setCurrentChar: vi.fn(), rollMode: 'normal' }));
vi.mock('../lib/animations', () => ({ animateHpChange: vi.fn() }));

import { equippedWeapons, hitDiceRemaining, isSaveProficient, weaponAttackBonus } from './combat';

describe('saving throw proficiencies', () => {
  it('applies class proficiencies for a single-class fighter', () => {
    const fighter = { class: 'Fighter' };
    expect(isSaveProficient(fighter, 'str')).toBe(true);
    expect(isSaveProficient(fighter, 'con')).toBe(true);
    expect(isSaveProficient(fighter, 'dex')).toBe(false);
    expect(isSaveProficient(fighter, 'wis')).toBe(false);
  });

  it('handles multi-class characters and subclasses', () => {
    expect(isSaveProficient({ classes: [{ class: 'Rogue' }, { class: 'Wizard' }] }, 'int')).toBe(true);
    expect(isSaveProficient({ class: 'Cleric (Life Domain)' }, 'wis')).toBe(true);
  });

  it('prefers an explicit saving_throw_proficiencies override', () => {
    expect(isSaveProficient({ class: 'Fighter', saving_throw_proficiencies: ['wis'] }, 'wis')).toBe(true);
    expect(isSaveProficient({ class: 'Fighter', saving_throw_proficiencies: ['wis'] }, 'str')).toBe(false);
  });

  it('honours custom save proficiencies', () => {
    expect(isSaveProficient({ class: 'Commoner', proficiencies: [{ name: 'Cha', type: 'save' }] }, 'cha')).toBe(true);
  });
});

describe('equipped weapon shortcuts', () => {
  const inventory = [
    { name: 'Longsword', category: 'weapon', equipped: true, damage_dice: '1d8' },
    { name: 'Shield', category: 'armor', equipped: true },
    { name: 'Dagger', category: 'weapon', equipped: false, damage_dice: '1d4' },
    { name: 'Rapier', category: 'weapon', is_equipped: true, damage_dice: '1d8' },
  ];

  it('lists only equipped, identified weapons', () => {
    expect(equippedWeapons({ inventory }).map((w) => w.name)).toEqual(['Longsword', 'Rapier']);
  });

  it('derives attack bonus from the best ability mod plus proficiency', () => {
    expect(weaponAttackBonus({ str_mod: 3, dex_mod: 1, proficiency_bonus: 2 }, { name: 'Longsword' })).toBe(5);
  });

  it('uses an explicit attack bonus when present', () => {
    expect(weaponAttackBonus({ str_mod: 3, proficiency_bonus: 2 }, { name: 'Longsword', attack_bonus: 7 })).toBe(7);
  });
});

describe('hit dice tracking', () => {
  it('uses the persisted remaining hit dice, falling back to total minus used', () => {
    expect(hitDiceRemaining({ hit_dice_current: 2 })).toBe(2);
    expect(hitDiceRemaining({ hit_dice_total: 3, hit_dice_used: 1 })).toBe(2);
    expect(hitDiceRemaining({ hit_dice_total: 2, hit_dice_used: 5 })).toBe(0);
    expect(hitDiceRemaining({})).toBe(0);
  });
});
