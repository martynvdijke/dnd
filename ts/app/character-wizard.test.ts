import { describe, it, expect, vi, beforeEach } from 'vitest';

const apiMock = vi.fn();
vi.mock('../lib/api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));

const toastMock = vi.fn();
const showModalMock = vi.fn();
const hideModalMock = vi.fn();
vi.mock('../lib/dom', async (importOriginal) => {
  const orig: any = await importOriginal();
  return {
    ...orig,
    esc: (s: string | null | undefined) => s ?? '',
    toast: (...a: unknown[]) => toastMock(...a),
    hideModal: (...a: unknown[]) => hideModalMock(...a),
    // Mirror real showModal by injecting the body so DOM-driven steps work.
    showModal: (title: string, body: string) => {
      showModalMock(title, body);
      document.getElementById('genericModalBody')!.innerHTML = body;
    },
  };
});
vi.mock('../lib/expose', () => ({ expose: () => {} }));

import {
  newCharWizard, wizardNext, wizardSetMethod, wizardCreate,
  wizardToggleSkill, wizardToggleSave,
} from './character-wizard';

const body = () => document.getElementById('genericModalBody')!.innerHTML;
const lastTitle = () => {
  const calls = showModalMock.mock.calls;
  return (calls[calls.length - 1]?.[0] ?? '') as string;
};

beforeEach(() => {
  apiMock.mockReset();
  apiMock.mockResolvedValue({});
  toastMock.mockReset();
  showModalMock.mockReset();
  hideModalMock.mockReset();
  document.body.innerHTML = '<div id="genericModalBody"></div><div id="toastContainer"></div>';
  (window as any).openChar = vi.fn();
});

describe('character wizard', () => {
  it('opens on the identity step', () => {
    newCharWizard();
    expect(lastTitle()).toContain('Step 1 of 4');
    expect(body()).toContain('data-testid="character-wizard"');
    expect(body()).toContain('data-testid="wizard-next"');
    expect(body()).toContain('id="wizName"');
  });

  it('blocks an empty name and advances once filled', () => {
    newCharWizard();
    wizardNext();
    expect(toastMock).toHaveBeenCalledWith('Name is required', true);
    expect(lastTitle()).toContain('Step 1 of 4');

    (document.getElementById('wizName') as HTMLInputElement).value = 'Aria';
    wizardNext();
    expect(lastTitle()).toContain('Step 2 of 4');
    expect(body()).toContain('id="wizAbil-str"');
  });

  it('switches to manual ability entry', () => {
    newCharWizard();
    wizardNext();
    (document.getElementById('wizName') as HTMLInputElement).value = 'Aria';
    wizardSetMethod('manual');
    expect(body()).toContain('type="number"');
  });

  it('creates the character and its proficiencies', async () => {
    apiMock.mockImplementation((method: string, path: string) => {
      if (method === 'POST' && path === '/api/characters') return Promise.resolve({ id: 7 });
      return Promise.resolve({});
    });
    newCharWizard();
    (document.getElementById('wizName') as HTMLInputElement).value = 'Aria';
    wizardNext(); // -> abilities
    wizardNext(); // -> proficiencies
    wizardToggleSkill('Stealth');
    wizardToggleSave('dex');
    await wizardCreate();

    expect(apiMock).toHaveBeenCalledWith('POST', '/api/characters', expect.objectContaining({ name: 'Aria', str: 15, dex: 14 }));
    expect(apiMock).toHaveBeenCalledWith('POST', '/api/characters/7/proficiencies', { character_id: 7, type: 'skill', name: 'Stealth' });
    expect(apiMock).toHaveBeenCalledWith('POST', '/api/characters/7/proficiencies', { character_id: 7, type: 'save', name: 'dex' });
    expect(hideModalMock).toHaveBeenCalled();
    expect((window as any).openChar).toHaveBeenCalledWith(7);
  });
});
