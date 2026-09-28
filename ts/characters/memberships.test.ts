import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setCurrentChar, setCurrentUser, currentChar } from '../lib/state';

const apiMock = vi.fn();
vi.mock('../lib/api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));
vi.mock('../lib/expose', () => ({ expose: () => {} }));

import { toggleCharacterMembership, renderMemberships } from './stats';

beforeEach(() => {
  apiMock.mockReset();
  setCurrentUser(null);
  setCurrentChar(null);
  delete (window as any).canEditCharacter;
  delete (window as any).markDirty;
  document.body.innerHTML = '';
});

describe('character campaign memberships', () => {
  it('toggles memberships and queues a save', () => {
    (window as any).markDirty = vi.fn();
    setCurrentChar({ id: 1, user_id: 5, campaigns: [{ id: 2, name: 'A' }] });

    toggleCharacterMembership(2);
    expect(currentChar.campaign_ids).toEqual([]);
    expect((window as any).markDirty).toHaveBeenCalledTimes(1);

    toggleCharacterMembership(3);
    expect(currentChar.campaign_ids).toEqual([3]);
    expect((window as any).markDirty).toHaveBeenCalledTimes(2);
  });

  it('renders a read-only list for characters the user cannot manage', async () => {
    setCurrentUser({ id: 1, username: 'player', role: 'user' });
    setCurrentChar({ id: 2, user_id: 99, campaigns: [{ id: 2, name: 'Vault' }] });
    const el = document.createElement('div');
    await renderMemberships(el);
    expect(el.textContent).toContain('Vault');
    expect(el.querySelector('input')).toBeNull();
  });

  it('renders checkboxes for owners, marking current memberships', async () => {
    (window as any).canEditCharacter = true;
    setCurrentUser({ id: 9, username: 'owner', role: 'user' });
    setCurrentChar({ id: 2, user_id: 9, campaigns: [{ id: 2, name: 'Vault' }] });
    apiMock.mockResolvedValue([{ id: 2, name: 'Vault' }, { id: 3, name: 'Strahd', my_role: 'dm' }]);
    const el = document.createElement('div');
    await renderMemberships(el);
    const checked = el.querySelectorAll<HTMLInputElement>('input[data-testid="membership-campaign-option"]');
    expect(checked).toHaveLength(2);
    expect((el.querySelector('#membershipCampaign-2') as HTMLInputElement).checked).toBe(true);
    expect((el.querySelector('#membershipCampaign-3') as HTMLInputElement).checked).toBe(false);
  });
});
