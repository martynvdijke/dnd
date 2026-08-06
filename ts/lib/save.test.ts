import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getAutoSaveInterval, setAutoSaveInterval, markDirty, markClean, isDirty, isSaving, startAutoSave, stopAutoSave, saveCharacter } from './save';
import { setCurrentChar } from './state';

// Mock api so saveCharacter does not hit the network
const apiMock = vi.fn();
vi.mock('./api', () => ({ api: (...args: unknown[]) => apiMock(...args) }));

const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (k: string) => store[k] ?? null,
    setItem: (k: string, v: string) => { store[k] = String(v); },
    removeItem: (k: string) => { delete store[k]; },
    clear: () => { store = {}; },
  };
})();
vi.stubGlobal('localStorage', localStorageMock);

describe('lib/save', () => {
  beforeEach(() => {
    localStorageMock.clear();
    apiMock.mockReset();
    markClean();
    stopAutoSave();
    setCurrentChar({ id: 1, name: 'Test', race: 'Human', class: 'Wizard', level: 1 } as any);
  });
  afterEach(() => {
    stopAutoSave();
    vi.useRealTimers();
  });

  it('getAutoSaveInterval defaults to 12', () => {
    expect(getAutoSaveInterval()).toBe(12);
  });

  it('getAutoSaveInterval reads stored value', () => {
    localStorageMock.setItem('villum-autosave-interval', '30');
    expect(getAutoSaveInterval()).toBe(30);
  });

  it('setAutoSaveInterval writes localStorage', () => {
    setAutoSaveInterval(45);
    expect(localStorageMock.getItem('villum-autosave-interval')).toBe('45');
  });

  it('markDirty/markClean toggle dirty flag', () => {
    expect(isDirty()).toBe(false);
    markDirty();
    expect(isDirty()).toBe(true);
    markClean();
    expect(isDirty()).toBe(false);
  });

  it('saveCharacter PUTs current char, flips isSaving, and marks clean on success', async () => {
    let resolveApi: (v: unknown) => void;
    apiMock.mockImplementation(() => new Promise((res) => { resolveApi = res; }));
    const p = saveCharacter();
    expect(isSaving()).toBe(true);
    resolveApi!({ id: 1, name: 'Test', race: 'Human', class: 'Wizard', level: 1 });
    await p;
    expect(isSaving()).toBe(false);
    expect(isDirty()).toBe(false);
    expect(apiMock).toHaveBeenCalledTimes(1);
    expect(apiMock.mock.calls[0][0]).toBe('PUT');
    expect(apiMock.mock.calls[0][1]).toBe('/api/characters/1');
  });

  it('scheduler saves when dirty and stops when clean', () => {
    vi.useFakeTimers();
    startAutoSave();
    markDirty();
    expect(apiMock).not.toHaveBeenCalled();
    vi.advanceTimersByTime(12000);
    expect(apiMock).toHaveBeenCalledTimes(1);
    // after save success, dirty is false — next tick no call
    vi.advanceTimersByTime(12000);
    expect(apiMock).toHaveBeenCalledTimes(1);
  });

  it('stopAutoSave halts the scheduler', () => {
    vi.useFakeTimers();
    startAutoSave();
    markDirty(); // starts scheduler via ensureScheduler
    stopAutoSave(); // must clear the running interval
    vi.advanceTimersByTime(60000);
    expect(apiMock).not.toHaveBeenCalled();
  });
});

describe('autoSaveField integration (7.2)', () => {
  beforeEach(() => {
    localStorageMock.clear();
    apiMock.mockReset();
    markClean();
    stopAutoSave();
    setCurrentChar({ id: 1, name: 'Test', race: 'Human', class: 'Wizard', level: 1 } as any);
  });
  afterEach(() => {
    stopAutoSave();
    vi.useRealTimers();
  });

  it('autoSaveField marks dirty and does not call the API directly', async () => {
    const { autoSaveField } = await import('../characters/sheet');
    vi.useFakeTimers();
    autoSaveField('race', { value: 'Elf' } as any);
    expect(isDirty()).toBe(true);
    expect(apiMock).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1000);
    expect(apiMock).not.toHaveBeenCalled(); // scheduler not started — markDirty only
  });
});
