import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setCurrentCampaign } from './lib/state';

const apiMock = vi.fn();
vi.mock('./lib/api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));
vi.mock('./navigation', () => ({ showView: vi.fn() }));
vi.mock('./lib/expose', () => ({ expose: () => {} }));
// keep real esc
vi.mock('./lib/dom', async (importOriginal) => {
  const orig: any = await importOriginal();
  return { ...orig, toast: vi.fn(), showModal: vi.fn(), hideModal: vi.fn() };
});

import { showBattlemap } from './battlemap';

describe('battlemap', () => {
  beforeEach(() => {
    apiMock.mockReset();
    document.body.innerHTML = '<div id="battlemapContent"></div>';
    setCurrentCampaign({ id: 42, name: 'Test Campaign' } as any);
  });

  it('renders board and tokens from api response', async () => {
    apiMock.mockResolvedValueOnce({
      map: { id: 1, name: 'Tavern', image_url: '', width: 1000, height: 800, grid_size: 50, grid_units: 'ft' },
      tokens: [
        { id: 1, name: 'Goblin', x: 0.5, y: 0.5, color: '#ff0000', size: 1, hp_current: 7, hp_max: 10, ac: 15 },
        { id: 2, name: 'Hero', x: 0.2, y: 0.3, color: '#00ff00', size: 1 },
      ],
    });
    await showBattlemap();
    const board = document.querySelector('[data-testid="battlemap-board"]');
    expect(board).not.toBeNull();
    const tokens = document.querySelectorAll('.bm-token');
    expect(tokens.length).toBe(2);
  });

  it('grid overlay contains linear-gradient with expected row/col counts', async () => {
    // 1000/50=20 cols, 800/50=16 rows => background-size 5% 6.25%
    apiMock.mockResolvedValueOnce({
      map: { id: 1, name: 'Map', image_url: '', width: 1000, height: 800, grid_size: 50, grid_units: '' },
      tokens: [],
    });
    await showBattlemap();
    const html = document.getElementById('battlemapContent')!.innerHTML;
    expect(html).toContain('linear-gradient');
    expect(html).toContain('5%');
    expect(html).toContain('6.25%');
  });

  it('token HTML contains name and HP bar/colour', async () => {
    apiMock.mockResolvedValueOnce({
      map: { id: 1, name: 'Map', image_url: '', width: 500, height: 500, grid_size: 50, grid_units: '' },
      tokens: [
        { id: 5, name: 'Orc', x: 0.1, y: 0.1, color: '#123456', size: 1, hp_current: 8, hp_max: 10, ac: 13 },
      ],
    });
    await showBattlemap();
    const html = document.getElementById('battlemapContent')!.innerHTML;
    expect(html).toContain('Orc');
    // HP bar colour green when >50%
    expect(html).toContain('#2f9e44');
    expect(html).toContain('8/10');
    expect(html).toContain('bm-hp');
  });

  it('uses campaignId override when provided', async () => {
    setCurrentCampaign(null);
    apiMock.mockResolvedValueOnce({ map: null, tokens: [] });
    await showBattlemap(99);
    expect(apiMock).toHaveBeenCalledWith('GET', '/api/campaigns/99/battlemap');
  });
});
