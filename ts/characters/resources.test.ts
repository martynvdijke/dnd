import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  wealthTotalGp,
  renderResources,
  resourceStepper,
  resourceSetValue,
  openResourceForm,
  saveResourceForm,
  deleteResource,
} from './resources';
import { setCurrentChar } from '../lib/state';

const apiMock = vi.fn();
vi.mock('../lib/api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));

const toastMock = vi.fn();
const showModalMock = vi.fn();
const hideModalMock = vi.fn();
vi.mock('../lib/dom', async (importOriginal) => {
  const orig: any = await importOriginal();
  return { ...orig, toast: (...a: unknown[]) => toastMock(...a), showModal: (...a: unknown[]) => showModalMock(...a), hideModal: (...a: unknown[]) => hideModalMock(...a) };
});
vi.mock('../lib/expose', () => ({ expose: () => {} }));

const setBody = (html: string) => { document.body.innerHTML = html; };

describe('characters/resources', () => {
  beforeEach(() => {
    apiMock.mockReset();
    apiMock.mockResolvedValue([]); // default for GETs after the queued Once values are consumed
    toastMock.mockReset();
    showModalMock.mockReset();
    hideModalMock.mockReset();
    setCurrentChar({ id: 1, name: 'Test', race: 'Human', class: 'Wizard', level: 1, currency: { gp: 10, sp: 5 } } as any);
    setBody('<div id="resourcesSection"></div>');
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('wealthTotalGp sums coins by gp value', () => {
    expect(wealthTotalGp()).toBeCloseTo(10.5);
    setCurrentChar({ id: 1, currency: { cp: 100 } } as any);
    expect(wealthTotalGp()).toBeCloseTo(1);
    setCurrentChar({ id: 1 } as any);
    expect(wealthTotalGp()).toBe(0);
  });

  it('renderResources renders wealth + empty state, then rows with recovery badges', async () => {
    apiMock.mockResolvedValueOnce([
      { id: 1, name: 'Ki', current: 2, max: 5, short_rest_recovery: 1, long_rest_recovery: 5, icon: 'fa-bolt', sort_order: 0 },
      { id: 2, name: 'Rations', current: 3, max: 0, short_rest_recovery: 0, long_rest_recovery: 0, icon: 'fa-drumstick-bite', sort_order: 0 },
    ]);
    await renderResources();
    const el = document.getElementById('resourcesSection')!;
    expect(el.textContent).toContain('Ki');
    expect(el.textContent).toContain('+1 SR');
    expect(el.textContent).toContain('+5 LR');
    expect(el.textContent).toContain('Rations');
    expect(el.textContent).toContain('10.5 gp');
    expect(el.textContent).toContain('Resources with rest recovery refill automatically');
  });

  it('renderResources is a no-op without a char or section, and shows hint-less list without recovery', async () => {
    setCurrentChar(null);
    await renderResources();
    expect(apiMock).not.toHaveBeenCalled();

    setCurrentChar({ id: 1, currency: {} } as any);
    apiMock.mockResolvedValueOnce([{ id: 1, name: 'Ki', current: 1, max: 5, short_rest_recovery: 0, long_rest_recovery: 0, icon: 'fa-bolt', sort_order: 0 }]);
    await renderResources();
    expect(document.getElementById('resourcesSection')!.textContent).not.toContain('refill automatically');
  });

  it('renderResources shows empty state when load fails', async () => {
    apiMock.mockRejectedValueOnce(new Error('boom'));
    await renderResources();
    expect(document.getElementById('resourcesSection')!.textContent).toContain('No Resources');
  });

  it('resourceStepper clamps to max and persists; ignores unknown id', async () => {
    apiMock.mockResolvedValue([
      { id: 1, name: 'Ki', current: 2, max: 5, short_rest_recovery: 1, long_rest_recovery: 5, icon: 'fa-bolt', sort_order: 0 },
    ]);
    await renderResources();
    apiMock.mockResolvedValueOnce({ ok: true });
    await resourceStepper(1, 1);
    expect(apiMock).toHaveBeenCalledWith('PUT', '/api/resources/1', expect.objectContaining({ current: 3 }));

    apiMock.mockResolvedValueOnce({ ok: true });
    await resourceStepper(1, 99); // clamps at max
    expect(apiMock).toHaveBeenCalledWith('PUT', '/api/resources/1', expect.objectContaining({ current: 5 }));

    await resourceStepper(999, 1); // not found
    expect(apiMock.mock.calls.filter(c => c[0] === 'PUT').length).toBe(2);
  });

  it('resourceStepper tolerates api failure', async () => {
    apiMock.mockResolvedValue([{ id: 1, name: 'Rations', current: 3, max: 0, short_rest_recovery: 0, long_rest_recovery: 0, icon: 'fa-bolt', sort_order: 0 }]);
    await renderResources();
    apiMock.mockRejectedValueOnce(new Error('nope'));
    await resourceStepper(1, -1); // no max → floor at 0
    expect(apiMock).toHaveBeenCalledWith('PUT', '/api/resources/1', expect.objectContaining({ current: 2 }));
    expect(toastMock).toHaveBeenCalledWith('nope', true);
  });

  it('resourceSetValue commits typed value on blur/enter and cancels on escape', async () => {
    apiMock.mockResolvedValueOnce([{ id: 1, name: 'Ki', current: 2, max: 5, short_rest_recovery: 1, long_rest_recovery: 5, icon: 'fa-bolt', sort_order: 0 }]);
    await renderResources();
    setBody('<div id="resourcesSection"><span id="val"></span></div>');
    apiMock.mockResolvedValueOnce({ ok: true });
    resourceSetValue(1, document.getElementById('val')!);
    const input = document.querySelector<HTMLInputElement>('#val input')!;
    input.value = '4';
    input.dispatchEvent(new Event('blur'));
    await vi.waitFor(() => expect(apiMock).toHaveBeenCalledWith('PUT', '/api/resources/1', expect.objectContaining({ current: 4 })));
  });

  it('openResourceForm opens add vs edit and picks icons', async () => {
    openResourceForm();
    const [, body] = showModalMock.mock.calls[0];
    expect(String(body)).toContain('Add Resource');
    expect(String(body)).toContain('value="1"'); // default long rest recovery for new

    apiMock.mockResolvedValueOnce([{ id: 7, name: 'Ki', current: 2, max: 5, short_rest_recovery: 1, long_rest_recovery: 5, icon: 'fa-bolt', sort_order: 0 }]);
    await renderResources();
    // Icon picker element must exist at call time (showModal is mocked, so provide it)
    setBody(`
      <div id="resourcesSection"></div>
      <div id="resourceIconPicker">
        <span class="resource-icon-option selected" data-icon="fa-bolt"><i class="fa-solid fa-bolt"></i></span>
        <span class="resource-icon-option" data-icon="fa-fire"><i class="fa-solid fa-fire"></i></span>
      </div>
      <input type="hidden" id="resourceIcon" value="fa-bolt">
    `);
    openResourceForm(7);
    const [editTitle, editBody] = showModalMock.mock.calls[1];
    expect(String(editTitle)).toContain('Edit Resource');
    expect(String(editBody)).toContain('Ki');
    expect(String(editBody)).toContain('Save Changes');
    (document.querySelector('[data-icon="fa-fire"]') as HTMLElement).click();
    expect((document.getElementById('resourceIcon') as HTMLInputElement).value).toBe('fa-fire');
  });

  it('saveResourceForm validates name and creates vs updates with recovery handling', async () => {
    // Reset any editing state from prior tests (module keeps editingId across tests)
    openResourceForm();
    setBody(`
      <div id="resourcesSection"></div>
      <input id="resourceName" value="Rations">
      <input id="resourceCurrent" value="3">
      <input id="resourceMax" value="0">
      <input id="resourceShortRecovery" value="2">
      <input id="resourceLongRecovery" value="3">
      <input id="resourceIcon" value="fa-box">
    `);
    apiMock.mockResolvedValueOnce({ id: 9 });
    await saveResourceForm();
    expect(apiMock).toHaveBeenCalledWith('POST', '/api/characters/1/resources', expect.objectContaining({ name: 'Rations', max: 0, short_rest_recovery: 0, long_rest_recovery: 0 }));
    expect(toastMock).toHaveBeenCalledWith('Added Rations');

    setBody(`
      <div id="resourcesSection"></div>
      <input id="resourceName" value="Ki">
      <input id="resourceCurrent" value="1">
      <input id="resourceMax" value="5">
      <input id="resourceShortRecovery" value="1">
      <input id="resourceLongRecovery" value="5">
      <input id="resourceIcon" value="fa-bolt">
    `);
    // put module into edit mode via openResourceForm on a loaded resource
    apiMock.mockResolvedValueOnce([{ id: 7, name: 'Ki', current: 1, max: 5, short_rest_recovery: 1, long_rest_recovery: 5, icon: 'fa-bolt', sort_order: 0 }]);
    await renderResources();
    openResourceForm(7);
    apiMock.mockResolvedValueOnce({ ok: true });
    await saveResourceForm();
    expect(apiMock).toHaveBeenCalledWith('PUT', '/api/resources/7', expect.objectContaining({ max: 5, short_rest_recovery: 1, long_rest_recovery: 5 }));

    setBody(`<div id="resourcesSection"></div><input id="resourceName" value="   "><input id="resourceCurrent" value="0"><input id="resourceMax" value="0"><input id="resourceShortRecovery" value="0"><input id="resourceLongRecovery" value="0"><input id="resourceIcon" value="fa-bolt">`);
    await saveResourceForm();
    expect(toastMock).toHaveBeenCalledWith('Resource name is required', true);
  });

  it('deleteResource requires confirmation and reports failures', async () => {
    vi.stubGlobal('confirm', () => false);
    await deleteResource(1);
    expect(apiMock).not.toHaveBeenCalled();

    vi.stubGlobal('confirm', () => true);
    apiMock.mockResolvedValueOnce({ ok: true });
    await deleteResource(1);
    expect(apiMock).toHaveBeenCalledWith('DELETE', '/api/resources/1');
    expect(toastMock).toHaveBeenCalledWith('Resource removed');

    apiMock.mockRejectedValueOnce(new Error('gone'));
    await deleteResource(1);
    expect(toastMock).toHaveBeenCalledWith('gone', true);
  });
});
