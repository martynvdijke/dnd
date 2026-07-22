import { describe, it, expect, beforeEach } from 'vitest';
import { highlightMatch, getRecents, addRecent, clearRecents } from './search';

// happy-dom doesn't provide localStorage by default
beforeEach(() => {
  const store: Record<string, string> = {};
  (globalThis as any).localStorage = {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, val: string) => { store[key] = val; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { Object.keys(store).forEach(k => delete store[k]); },
    get length() { return Object.keys(store).length; },
    key: (i: number) => Object.keys(store)[i] ?? null,
  };
});

describe('highlightMatch', () => {
  it('wraps matching terms in <mark> tags', () => {
    expect(highlightMatch('Fireball spell', 'fire')).toBe('<mark>Fire</mark>ball spell');
  });

  it('is case-insensitive', () => {
    expect(highlightMatch('Fireball', 'FIRE')).toBe('<mark>Fire</mark>ball');
  });

  it('splits multi-word queries and highlights each word', () => {
    const result = highlightMatch('fireball spell level', 'fire level');
    expect(result).toContain('<mark>fire</mark>');
    expect(result).toContain('<mark>level</mark>');
  });

  it('returns original text when query is empty', () => {
    expect(highlightMatch('Hello', '')).toBe('Hello');
  });

  it('escapes regex special characters in query', () => {
    const result = highlightMatch('hello (world)', '(world)');
    expect(result).toContain('<mark>');
    expect(result).toContain('(world)');
  });

  it('returns original text unchanged when no match', () => {
    expect(highlightMatch('Hello world', 'xyz')).toBe('Hello world');
  });
});

describe('recent searches', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('returns empty array initially', () => {
    expect(getRecents()).toEqual([]);
  });

  it('adds a recent search', () => {
    addRecent('fireball');
    expect(getRecents()).toEqual(['fireball']);
  });

  it('deduplicates recent searches moving newest to front', () => {
    addRecent('fireball');
    addRecent('magic missile');
    addRecent('fireball');
    expect(getRecents()).toEqual(['fireball', 'magic missile']);
  });

  it('ignores empty or whitespace queries', () => {
    addRecent('');
    addRecent('   ');
    expect(getRecents()).toEqual([]);
  });

  it('caps at MAX_RECENTS (10)', () => {
    for (let i = 0; i < 15; i++) {
      addRecent(`query-${i}`);
    }
    expect(getRecents().length).toBe(10);
    expect(getRecents()[0]).toBe('query-14');
  });

  it('clearRecents removes all recents', () => {
    addRecent('test');
    clearRecents();
    expect(getRecents()).toEqual([]);
  });
});
