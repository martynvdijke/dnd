import { describe, it, expect, beforeEach } from 'vitest';
import { highlightMatch, getRecents, addRecent, clearRecents } from './search';

function stubStorage() {
  const store: Record<string, string> = {};
  (globalThis as any).localStorage = {
    getItem: (k: string) => store[k] ?? null,
    setItem: (k: string, v: string) => { store[k] = v; },
    removeItem: (k: string) => { delete store[k]; },
    clear: () => { Object.keys(store).forEach(k => delete store[k]); },
    get length() { return Object.keys(store).length; },
    key: (i: number) => Object.keys(store)[i] ?? null,
  };
  return store;
}

beforeEach(() => stubStorage());

describe('highlightMatch extra', () => {
  it('escapes regex special chars: . * + ? ^ $ { } ( ) | [ ] \\', () => {
    expect(highlightMatch('a.b*c+d?e^f$g{h}i(j)k|l[m]n\\o', 'a.b*c+d?e^f$g{h}i(j)k|l[m]n\\o')).toContain('<mark>');
  });
  it('handles query with multiple spaces', () => {
    const r = highlightMatch('fireball spell', 'fire  spell');
    expect(r).toContain('<mark>fire</mark>');
    expect(r).toContain('<mark>spell</mark>');
  });
  it('returns text when query empty', () => {
    expect(highlightMatch('hello', '')).toBe('hello');
  });
  it('returns text on invalid regex edge — fallback', () => {
    // highlightMatch catches errors internally
    expect(highlightMatch('hello', '')).toBe('hello');
    // query that could throw if not escaped — but it is escaped so should highlight
    expect(highlightMatch('hello [world]', '[world]')).toContain('<mark>');
  });
  it('highlights case-insensitive matches', () => {
    expect(highlightMatch('Fireball', 'fireball')).toBe('<mark>Fireball</mark>');
  });
});

describe('getRecents/addRecent/clearRecents extra', () => {
  it('getRecents returns [] on invalid JSON', () => {
    const store = stubStorage();
    store['villum-search-recents'] = 'not-json[';
    expect(getRecents()).toEqual([]);
  });
  it('addRecent dedups and moves to front', () => {
    addRecent('a');
    addRecent('b');
    addRecent('a');
    expect(getRecents()).toEqual(['a', 'b']);
  });
  it('addRecent slices to 10', () => {
    for (let i = 0; i < 20; i++) addRecent(`q${i}`);
    const recents = getRecents();
    expect(recents.length).toBe(10);
    expect(recents[0]).toBe('q19');
    expect(recents[9]).toBe('q10');
  });
  it('clearRecents empties', () => {
    addRecent('hello');
    clearRecents();
    expect(getRecents()).toEqual([]);
  });
  it('ignores whitespace-only query', () => {
    addRecent('   ');
    expect(getRecents()).toEqual([]);
  });
});
