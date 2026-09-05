import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setCurrentChar } from '../lib/state';
import { renderCombat } from './combat';

vi.mock('../lib/api', () => ({ api: vi.fn().mockResolvedValue([]) }));

function mockChar() {
  return {
    id: 1, ac: 15, initiative: 2, speed: 30,
    hp_max: 20, hp_current: 12, temp_hp: 0,
    proficiency_bonus: 2,
    str_mod: 3, dex_mod: 2, con_mod: 1, int_mod: 0, wis_mod: 1, cha_mod: -1,
    death_saves_successes: 1, death_saves_failures: 0,
    concentrating: false, concentrating_on: '',
    hit_dice_total: 3, hit_dice_used: 1,
  };
}

beforeEach(() => {
  document.body.innerHTML = '<div id="combatSection"></div><div id="conditionBadges"></div>';
  // renderCombat calls renderStepper via window
  (window as any).renderStepper = (field: string, value: number) => `<span class="stepper" data-field="${field}">${value}</span>`;
  setCurrentChar(mockChar());
});

describe('renderCombat', () => {
  it('renders HP bar and title', () => {
    renderCombat();
    const el = document.getElementById('combatSection')!;
    expect(el.innerHTML).toContain('Hit Points');
    expect(el.innerHTML).toContain('12 / 20');
    expect(el.innerHTML).toContain('charHpBarFill');
  });

  it('renders AC/initiative/speed steppers', () => {
    renderCombat();
    const html = document.getElementById('combatSection')!.innerHTML;
    expect(html).toContain('AC');
    expect(html).toContain('Initiative');
    expect(html).toContain('Speed');
  });

  it('handles temp HP', () => {
    const c = mockChar(); c.temp_hp = 5;
    setCurrentChar(c);
    renderCombat();
    expect(document.getElementById('combatSection')!.innerHTML).toContain('temp');
  });

  it('shows death saves and concentration', () => {
    renderCombat();
    const html = document.getElementById('combatSection')!.innerHTML;
    expect(html).toContain('Death Saves');
    expect(html).toContain('Concentration');
  });

  it('renders saving throws', () => {
    renderCombat();
    const html = document.getElementById('combatSection')!.innerHTML;
    expect(html).toContain('STR');
    expect(html).toContain('DEX');
  });
});
