import { describe, expect, it, vi } from 'vitest';

vi.mock('./navigation', () => ({ showView: vi.fn() }));
vi.mock('./lib/api', () => ({ api: vi.fn() }));
vi.mock('./lib/state', () => ({ currentUser: null, currentChar: null, currentCampaign: null }));

import { partyPassive } from './party';

describe('party passive scores', () => {
  it('is 10 + ability modifier when not proficient', () => {
    expect(partyPassive({ wis_mod: 3, proficiency_bonus: 2, proficiencies: [] }, 'Perception', 'wis')).toBe(13);
  });

  it('adds the proficiency bonus for proficient skills', () => {
    expect(partyPassive({ wis_mod: 3, proficiency_bonus: 2, proficiencies: [{ name: 'Perception', type: 'skill' }] }, 'Perception', 'wis')).toBe(15);
  });

  it('uses the requested ability for investigation', () => {
    expect(partyPassive({ int_mod: 1, proficiency_bonus: 2, proficiencies: [{ name: 'Investigation', type: 'skill' }] }, 'Investigation', 'int')).toBe(13);
  });

  it('falls back to 10 without character data', () => {
    expect(partyPassive(null, 'Insight', 'wis')).toBe(10);
  });
});
