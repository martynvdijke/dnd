import { describe, it, expect, beforeEach, vi } from 'vitest';

const { exposed } = vi.hoisted(() => ({ exposed: {} as Record<string, any> }));

vi.mock('../lib/expose', () => ({
  expose: (name: string, value: any) => { exposed[name] = value; return value; },
}));
vi.mock('../lib/dom', () => ({
  esc: (s: any) => String(s ?? ''),
  showModal: (_title: string, body: string) => { document.body.innerHTML = body; },
  hideModal: vi.fn(),
  toast: vi.fn(),
}));
vi.mock('../lib/api', () => ({ api: vi.fn() }));
vi.mock('../navigation', () => ({ showView: vi.fn() }));
vi.mock('../lib/ambience', () => ({
  playAmbience: vi.fn(), stopAmbience: vi.fn(), listAmbienceTracks: () => [],
}));

import { api } from '../lib/api';
import './campaign-dashboard';

const mockApi = api as unknown as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockApi.mockReset();
  document.body.innerHTML = '';
});

describe('campaign-dashboard combat analytics', () => {
  it('renders the combat analytics card', async () => {
    mockApi.mockImplementation((_method: string, path: string) => {
      if (path.includes('/combat-log/stats')) {
        return Promise.resolve({
          total_damage: 42, total_healing: 12, crit_count: 1,
          damage_by_type: [{ type: 'fire', damage: 20 }],
          top_healers: [{ name: 'Cleric', healing: 12 }],
        });
      }
      return Promise.resolve({
        characters: [], active_quests: 0, upcoming_sessions: 0, active_conditions: 0,
        downtime_count: 0, recent_journal: 0, total_members: 0,
        upcoming_events: [], recent_timeline: [], recent_recaps: [], recent_combats: [], recent_dice_rolls: [],
      });
    });

    await exposed.showCampaignDashboard(1, 'Test Camp');
    const html = document.getElementById('campaignDashContent')!.innerHTML;
    expect(html).toContain('data-testid="dash-combat-analytics"');
    expect(html).toContain('fire');
    expect(html).toContain('Cleric');
  });
});
