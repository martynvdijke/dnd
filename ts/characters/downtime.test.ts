import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderDowntime } from './downtime';
import { setCurrentChar } from '../lib/state';

const apiMock = vi.fn();
vi.mock('../lib/api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));

const toastMock = vi.fn();
const showModalMock = vi.fn();
const hideModalMock = vi.fn();
vi.mock('../lib/dom', async (importOriginal) => {
  const orig: any = await importOriginal();
  return { ...orig, toast: (...a: unknown[]) => toastMock(...a), showModal: (...a: unknown[]) => showModalMock(...a), hideModal: (...a: unknown[]) => hideModalMock(...a), esc: (s: string | null | undefined) => s ?? '' };
});
vi.mock('../lib/expose', () => ({ expose: () => {} }));

describe('characters/downtime', () => {
  beforeEach(() => {
    apiMock.mockReset();
    apiMock.mockResolvedValue([]);
    toastMock.mockReset();
    showModalMock.mockReset();
    hideModalMock.mockReset();
    setCurrentChar({ id: 1, name: 'Test' } as any);
    document.body.innerHTML = '<div id="downtimeSection"></div>';
  });

  it('renderDowntime renders activity name and status badge', async () => {
    apiMock.mockResolvedValueOnce([
      { id: 1, name: 'Training Arc', activity_type: 'training', status: 'in-progress', dc: 15, days_required: 5, days_completed: 2, total_cost: 10, reward: '+1 skill', description: 'lifting' },
    ]);
    await renderDowntime();
    const el = document.getElementById('downtimeSection')!;
    expect(el.textContent).toContain('Training Arc');
    expect(el.textContent).toContain('in-progress');
    // status badge class for in-progress is bg-info
    expect(el.innerHTML).toContain('bg-info');
  });

  it('renderDowntime shows empty state when no activities', async () => {
    apiMock.mockResolvedValueOnce([]);
    await renderDowntime();
    const el = document.getElementById('downtimeSection')!;
    expect(el.textContent).toContain('No Downtime Activities');
  });

  it('renderDowntime is no-op without char or section', async () => {
    setCurrentChar(null);
    await renderDowntime();
    expect(apiMock).not.toHaveBeenCalled();

    setCurrentChar({ id: 1 } as any);
    document.body.innerHTML = '';
    await renderDowntime();
    expect(apiMock).not.toHaveBeenCalled();
  });

  it('renderDowntime handles api failure', async () => {
    apiMock.mockRejectedValueOnce(new Error('boom'));
    await renderDowntime();
    expect(document.getElementById('downtimeSection')!.textContent).toContain('boom');
  });
});
