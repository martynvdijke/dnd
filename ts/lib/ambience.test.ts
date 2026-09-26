import { describe, it, expect, beforeEach, vi } from 'vitest';

vi.mock('./expose', () => ({ expose: () => {} }));

import { listAmbienceTracks, playAmbience, getCurrentAmbience, stopAmbience } from './ambience';

describe('lib/ambience', () => {
  beforeEach(() => {
    delete (document.body.dataset as any).ambience;
    stopAmbience();
  });

  it('listAmbienceTracks returns expected tracks', () => {
    const tracks = listAmbienceTracks();
    const keys = tracks.map(t => t.key);
    expect(keys).toContain('rain');
    expect(keys).toContain('fire');
    expect(keys).toContain('combat');
    expect(tracks.length).toBeGreaterThanOrEqual(3);
    for (const t of tracks) {
      expect(t.label).toBeTruthy();
      expect(t.icon).toBeTruthy();
    }
  });

  it('playAmbience sets dataset and getCurrentAmbience', () => {
    playAmbience('rain', 0.4);
    expect(document.body.dataset.ambience).toBe('rain');
    expect(getCurrentAmbience()).toBe('rain');
  });

  it('stopAmbience clears dataset and current', () => {
    playAmbience('rain', 0.4);
    stopAmbience();
    expect(document.body.dataset.ambience).toBeUndefined();
    expect(getCurrentAmbience()).toBeNull();
  });

  it('unknown track does not set dataset and is safe', () => {
    stopAmbience();
    playAmbience('nonsense', 0.5);
    expect(document.body.dataset.ambience).toBeUndefined();
    expect(getCurrentAmbience()).toBeNull();
    // calling with unknown should not throw
    expect(() => playAmbience('nonsense')).not.toThrow();
  });
});
