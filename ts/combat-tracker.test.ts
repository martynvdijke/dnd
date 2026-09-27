import { beforeEach, describe, expect, it, vi } from 'vitest';

const { api } = vi.hoisted(() => ({ api: vi.fn() }));
vi.mock('./lib/api', () => ({ api }));
vi.mock('./navigation', () => ({ showView: vi.fn() }));
vi.mock('./lib/state', () => ({ currentCampaign: null }));

import './combat-tracker';

describe('combat tracker action economy', () => {
  beforeEach(() => {
    api.mockReset();
    api.mockResolvedValue({});
    document.body.innerHTML = '<div id="combatTrackerView"></div><div id="combatTrackerContent"></div>';
    (window as any).showCombatTracker = vi.fn();
  });

  it('spends an action', async () => {
    await (window as any).useCombatEconomy(5, 'action');
    expect(api).toHaveBeenCalledWith('POST', '/api/combat/5/economy', { action: true });
  });

  it('spends 5 ft of movement', async () => {
    await (window as any).useCombatEconomy(7, 'move');
    expect(api).toHaveBeenCalledWith('POST', '/api/combat/7/economy', { movement: 5 });
  });
});
