import { describe, it, expect, vi, beforeEach } from 'vitest';

const apiMock = vi.fn();
vi.mock('./lib/api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));
vi.mock('./lib/expose', () => ({ expose: () => {} }));
const showViewMock = vi.fn();
vi.mock('./navigation', () => ({ showView: (...args: unknown[]) => showViewMock(...args) }));
const toastMock = vi.fn();
vi.mock('./lib/dom', async (importOriginal) => {
  const orig: any = await importOriginal();
  return { ...orig, toast: (...a: unknown[]) => toastMock(...a) };
});

import { enterDefaultView, continueAsDm, selectCampaign } from './selection';
import { setCurrentCampaign, setCurrentChar, currentChar } from './lib/state';

const storage: Record<string, string> = {};
vi.stubGlobal('localStorage', {
  getItem: (k: string) => storage[k] ?? null,
  setItem: (k: string, v: string) => { storage[k] = v; },
  removeItem: (k: string) => { delete storage[k]; },
  clear: () => { Object.keys(storage).forEach((k) => delete storage[k]); },
});

beforeEach(() => {
  apiMock.mockReset();
  showViewMock.mockReset();
  toastMock.mockReset();
  localStorage.clear();
  setCurrentCampaign(null);
  setCurrentChar(null);
  delete (window as any).showParty;
  document.body.innerHTML = '<div id="campaignPickerList"></div><div id="characterPickerList"></div><button id="continueAsDmBtn"></button>';
});

describe('role-aware campaign selection', () => {
  it('sends DMs to the party view on the default view', () => {
    setCurrentCampaign({ id: 1, name: 'DM Camp', my_role: 'dm' });
    (window as any).showParty = vi.fn();
    enterDefaultView();
    expect((window as any).showParty).toHaveBeenCalled();
    expect(showViewMock).not.toHaveBeenCalled();
  });

  it('sends players to the characters view on the default view', () => {
    setCurrentCampaign({ id: 2, name: 'Player Camp', my_role: 'player' });
    enterDefaultView();
    expect(showViewMock).toHaveBeenCalledWith('characters');
  });

  it('skips the character picker when selecting a campaign you run', async () => {
    apiMock.mockResolvedValue([{ id: 7, name: 'Curse of Strahd', my_role: 'dm' }]);
    (window as any).showParty = vi.fn();
    await selectCampaign(7);
    expect((window as any).showParty).toHaveBeenCalled();
    expect(showViewMock).not.toHaveBeenCalledWith('characterPicker');
  });

  it('offers the character picker for campaigns you play in', async () => {
    apiMock.mockImplementation((_method: string, path: string) => {
      if (path === '/api/campaigns/mine') return Promise.resolve([{ id: 3, name: 'Tomb', my_role: 'player' }]);
      return Promise.resolve([]);
    });
    await selectCampaign(3);
    expect(showViewMock).toHaveBeenCalledWith('characterPicker');
    expect(document.getElementById('characterPickerList')!.innerHTML).toContain('No characters in this campaign yet.');
    expect((document.getElementById('continueAsDmBtn') as HTMLElement).style.display).toBe('none');
  });

  it('continueAsDm clears the character and opens the party view', () => {
    setCurrentCampaign({ id: 4, name: 'One-shot', my_role: 'dm' });
    setCurrentChar({ id: 99, name: 'Aria' });
    (window as any).showParty = vi.fn();
    continueAsDm(4);
    expect(currentChar).toBeNull();
    expect((window as any).showParty).toHaveBeenCalled();
  });
});
