import { describe, it, expect, beforeEach, vi } from 'vitest';

const { exposed } = vi.hoisted(() => ({ exposed: {} as Record<string, any> }));

vi.mock('../lib/expose', () => ({
  expose: (name: string, value: any) => { exposed[name] = value; return value; },
}));
vi.mock('../lib/dom', () => ({
  esc: (s: any) => String(s ?? ''),
  showModal: (_title: string, body: string) => { document.body.innerHTML = body; },
}));
vi.mock('../lib/api', () => ({ api: vi.fn() }));

import { api } from '../lib/api';
import './dm-tools';

const mockApi = api as unknown as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockApi.mockReset();
  document.body.innerHTML = '';
});

describe('dm-tools', () => {
  it('renders the DM tools modal with all generator controls', () => {
    exposed.showDmTools();
    const html = document.body.innerHTML;
    for (const id of ['dm-treasure-tier', 'dm-tool-treasure', 'dm-weather-biome', 'dm-weather-season', 'dm-tool-weather', 'dm-name-race', 'dm-tool-name', 'dm-budget-levels', 'dm-tool-budget']) {
      expect(html).toContain(`data-testid="${id}"`);
    }
  });

  it('generates a treasure hoard', async () => {
    exposed.showDmTools();
    mockApi.mockResolvedValue({ tier: 2, coins: { gp: 100 }, gems: ['Ruby'], art_objects: [], magic_items: [], total_gp_value: 150 });
    await exposed.dmGenerateTreasure();
    const html = document.getElementById('dmToolsResult')!.innerHTML;
    expect(html).toContain('Treasure Hoard');
    expect(html).toContain('Ruby');
    expect(html).toContain('150');
  });

  it('calculates an adventuring-day budget', async () => {
    exposed.showDmTools();
    mockApi.mockResolvedValue({ party_budget: 1200, suggested_encounters: '6-8 medium or hard encounters' });
    await exposed.dmDailyBudget();
    const html = document.getElementById('dmToolsResult')!.innerHTML;
    expect(html).toContain('1200 XP');
    expect(html).toContain('6-8 medium');
  });

  it('renders the campaign DM screen aggregate', async () => {
    vi.useFakeTimers();
    mockApi.mockResolvedValue({
      campaign: { name: 'DMScreenCamp', party_name: 'The Bold', dm_notes: 'notes' },
      session: null,
      combat: { active: true, current_turn: 'Goblin', entries: [{ id: 1 }] },
      party: [{ name: 'Hero', hp_current: 5, hp_max: 10, ac: 15, conditions: ['Poisoned'] }],
      ambience: { track: 'rain', action: 'play' },
    });
    await exposed.showDmScreen(1);
    const html = document.getElementById('dmScreenContent')!.innerHTML;
    expect(html).toContain('DMScreenCamp');
    expect(html).toContain('Goblin');
    expect(html).toContain('Hero');
    expect(html).toContain('Poisoned');
    expect(html).toContain('data-testid="dm-screen-refresh"');
    vi.clearAllTimers();
    vi.useRealTimers();
  });
});
