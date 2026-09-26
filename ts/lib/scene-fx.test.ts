import { describe, it, expect, beforeEach } from 'vitest';
import { SCENE_FX, playSceneEffect } from './scene-fx';

describe('lib/scene-fx', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });

  it('SCENE_FX contains expected effects', () => {
    expect(SCENE_FX).toContain('fire');
    expect(SCENE_FX).toContain('smoke');
    expect(SCENE_FX).toContain('lightning');
    expect(SCENE_FX.length).toBeGreaterThanOrEqual(3);
  });

  it('playSceneEffect appends div with correct classes', () => {
    playSceneEffect('fire');
    const el = document.body.querySelector('.scene-fx.scene-fx-fire');
    expect(el).not.toBeNull();
    expect(el!.tagName).toBe('DIV');
  });

  it('playSceneEffect does nothing for unknown effect', () => {
    playSceneEffect('nonsense');
    expect(document.body.querySelector('.scene-fx-nonsense')).toBeNull();
    expect(document.body.querySelector('.scene-fx')).toBeNull();
  });
});
