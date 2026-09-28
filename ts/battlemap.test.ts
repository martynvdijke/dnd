import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setCurrentCampaign } from './lib/state';

const apiMock = vi.fn();
const exposed: Record<string, (...args: any[]) => any> = vi.hoisted(() => ({}));
vi.mock('./lib/api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));
vi.mock('./navigation', () => ({ showView: vi.fn() }));
vi.mock('./lib/expose', () => ({ expose: (name: string, fn: any) => { exposed[name] = fn; } }));
// keep real esc
vi.mock('./lib/dom', async (importOriginal) => {
  const orig: any = await importOriginal();
  return { ...orig, toast: vi.fn(), showModal: vi.fn(), hideModal: vi.fn() };
});

import { showBattlemap } from './battlemap';

const map = { id: 1, name: 'Tavern', image_url: '', width: 1000, height: 800, grid_size: 50, grid_units: 'ft' };

describe('battlemap', () => {
  beforeEach(() => {
    apiMock.mockReset();
    document.body.innerHTML = '<div id="battlemapContent"></div>';
    setCurrentCampaign({ id: 42, name: 'Test Campaign' } as any);
  });

  it('renders board and tokens from api response', async () => {
    apiMock.mockResolvedValueOnce({
      map,
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
    apiMock.mockResolvedValueOnce({ map: { ...map, grid_units: '' }, tokens: [] });
    await showBattlemap();
    const html = document.getElementById('battlemapContent')!.innerHTML;
    expect(html).toContain('linear-gradient');
    expect(html).toContain('5%');
    expect(html).toContain('6.25%');
  });

  it('token HTML contains name and HP bar/colour', async () => {
    apiMock.mockResolvedValueOnce({
      map: { ...map, width: 500, height: 500, grid_units: '' },
      tokens: [{ id: 5, name: 'Orc', x: 0.1, y: 0.1, color: '#123456', size: 1, hp_current: 8, hp_max: 10, ac: 13 }],
    });
    await showBattlemap();
    const html = document.getElementById('battlemapContent')!.innerHTML;
    expect(html).toContain('Orc');
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

  it('renders an aura circle for a token with an aura radius', async () => {
    apiMock.mockResolvedValueOnce({
      map,
      tokens: [{ id: 1, name: 'Paladin', x: 0.5, y: 0.5, color: '#b8963e', size: 1, aura_radius: 2 }],
    });
    await showBattlemap();
    const aura = document.querySelector('.bm-aura');
    expect(aura).not.toBeNull();
    expect(aura!.getAttribute('r')).toBe('100'); // 2 cells * 50px
  });

  it('renders condition badges from the token conditions JSON', async () => {
    apiMock.mockResolvedValueOnce({
      map,
      tokens: [{ id: 1, name: 'Goblin', x: 0.5, y: 0.5, color: '#ff0000', size: 1, conditions: '["Poisoned","Prone"]' }],
    });
    await showBattlemap();
    const badges = document.querySelectorAll('.bm-condition');
    expect(badges.length).toBe(2);
    expect(document.getElementById('battlemapContent')!.innerHTML).toContain('Poisoned');
  });

  it('draws a vision mask when vision is enabled and a token is selected', async () => {
    apiMock.mockResolvedValueOnce({
      map: { ...map, walls: [{ x1: 0.5, y1: 0, x2: 0.5, y2: 1 }] },
      tokens: [{ id: 1, name: 'Scout', x: 0.25, y: 0.5, color: '#00ff00', size: 1, vision_radius: 6 }],
    });
    await showBattlemap();
    exposed.battlemapToggleVision();
    exposed.battlemapVisionToken('1');
    const mask = document.querySelector('.bm-vision-mask');
    expect(mask).not.toBeNull();
    expect(mask!.getAttribute('fill-rule')).toBe('evenodd');
  });

  it('saves pending walls via the walls endpoint', async () => {
    apiMock.mockResolvedValueOnce({ map, tokens: [] });
    await showBattlemap();
    exposed.battlemapToggleWalls();
    // Simulate two board clicks forming a wall.
    const board = document.getElementById('battlemapBoard')!;
    board.getBoundingClientRect = () => ({ left: 0, top: 0, width: 1000, height: 800 } as any);
    board.dispatchEvent(new PointerEvent('pointerdown', { clientX: 100, clientY: 100 }));
    board.dispatchEvent(new PointerEvent('pointerdown', { clientX: 200, clientY: 200 }));
    apiMock.mockResolvedValueOnce({ ok: true });
    apiMock.mockResolvedValueOnce({ map, tokens: [] });
    await exposed.battlemapSaveWalls();
    expect(apiMock).toHaveBeenCalledWith('PUT', '/api/maps/1/walls', {
      walls: [{ x1: 0.1, y1: 0.125, x2: 0.2, y2: 0.25 }],
    });
  });
});
