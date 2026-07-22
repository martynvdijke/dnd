import { describe, it, expect, beforeEach, vi } from 'vitest';
import { parseHash, routeToHash, navigate, getCurrentRoute, initRouter, navigateToInitialHash, resetRouter } from './router';

describe('parseHash', () => {
  it('parses empty hash to characters', () => {
    const r = parseHash('');
    expect(r.view).toBe('characters');
    expect(r.params).toEqual({});
  });

  it('parses #/ to characters', () => {
    const r = parseHash('#/');
    expect(r.view).toBe('characters');
  });

  it('parses simple view', () => {
    const r = parseHash('#/dice');
    expect(r.view).toBe('dice');
    expect(r.params).toEqual({});
  });

  it('parses view with id param', () => {
    const r = parseHash('#/sheet/42');
    expect(r.view).toBe('sheet');
    expect(r.params).toEqual({ id: '42' });
  });

  it('parses singleEncounter with id', () => {
    const r = parseHash('#/singleEncounter/7');
    expect(r.view).toBe('singleEncounter');
    expect(r.params).toEqual({ id: '7' });
  });

  it('parses compendium without id', () => {
    const r = parseHash('#/compendium');
    expect(r.view).toBe('compendium');
    expect(r.params).toEqual({});
  });
});

describe('routeToHash', () => {
  it('serializes simple view', () => {
    expect(routeToHash({ view: 'dice', params: {} })).toBe('#/dice');
  });

  it('serializes view with id', () => {
    expect(routeToHash({ view: 'sheet', params: { id: '42' } })).toBe('#/sheet/42');
  });
});

describe('navigate', () => {
  it('updates current route', () => {
    const r = navigate('dice');
    expect(r.view).toBe('dice');
    expect(getCurrentRoute().view).toBe('dice');
  });

  it('updates location.hash', () => {
    navigate('compendium');
    expect(location.hash).toBe('#/compendium');
  });

  it('sets params', () => {
    const r = navigate('sheet', { id: '99' });
    expect(r.params.id).toBe('99');
    expect(location.hash).toBe('#/sheet/99');
  });
});

describe('initRouter', () => {
  beforeEach(() => resetRouter());

  it('calls onNavigate on hashchange', () => {
    const onNav = vi.fn();
    initRouter(onNav);

    location.hash = '#/dice';
    window.dispatchEvent(new HashChangeEvent('hashchange'));

    expect(onNav).toHaveBeenCalledWith(
      expect.objectContaining({ view: 'dice' })
    );
  });
});

describe('navigateToInitialHash', () => {
  beforeEach(() => {
    location.hash = '';
  });

  it('returns null when no hash', () => {
    const onNav = vi.fn();
    const result = navigateToInitialHash(onNav);
    expect(result).toBeNull();
    expect(onNav).not.toHaveBeenCalled();
  });

  it('navigates to hash when present', () => {
    location.hash = '#/compendium';
    const onNav = vi.fn();
    const result = navigateToInitialHash(onNav);
    expect(result).not.toBeNull();
    expect(result!.view).toBe('compendium');
    expect(onNav).toHaveBeenCalledWith(
      expect.objectContaining({ view: 'compendium' })
    );
  });
});
