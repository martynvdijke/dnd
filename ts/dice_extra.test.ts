import { describe, it, expect } from 'vitest';
import { parseSides, rollValue, rollUsed, rollFlags, isCritRoll, buildDie } from './dice';

describe('parseSides edge cases', () => {
  it('returns 0 for null-ish strings', () => {
    expect(parseSides('d')).toBe(0);
    expect(parseSides('d-1')).toBe(0);
    expect(parseSides('dx')).toBe(0);
    expect(parseSides(' d20')).toBe(0);
    expect(parseSides('d20 ')).toBe(0);
  });
  it('handles d0 and large sides', () => {
    expect(parseSides('d0')).toBe(0);
    expect(parseSides('d1000')).toBe(1000);
    expect(parseSides('d00')).toBe(0);
  });
  it('is case-insensitive for D', () => {
    expect(parseSides('D100')).toBe(100);
  });
});

describe('rollValue edge', () => {
  it('returns 0 for null/undefined/false-like', () => {
    expect(rollValue(null)).toBe(0);
    expect(rollValue(undefined)).toBe(0);
    expect(rollValue({})).toBe(0);
    expect(rollValue({ value: undefined })).toBe(0);
    expect(rollValue({ value: 0 })).toBe(0);
  });
  it('returns numeric roll directly', () => {
    expect(rollValue(1)).toBe(1);
    expect(rollValue(100)).toBe(100);
  });
  it('prefers object value', () => {
    expect(rollValue({ value: 12 })).toBe(12);
    expect(rollValue({ value: 12, modifierFlags: 'dropped' })).toBe(12);
  });
});

describe('rollUsed edge', () => {
  it('plain numbers are used', () => {
    expect(rollUsed(5)).toBe(true);
  });
  it('object without useInTotal is used', () => {
    expect(rollUsed({ value: 5 })).toBe(true);
    expect(rollUsed({ value: 5, useInTotal: undefined })).toBe(true);
    expect(rollUsed({ value: 5, useInTotal: null as any })).toBe(true);
  });
  it('explicit false is not used', () => {
    expect(rollUsed({ value: 5, useInTotal: false })).toBe(false);
  });
});

describe('rollFlags edge', () => {
  it('returns empty for numbers', () => {
    expect(rollFlags(6)).toBe('');
  });
  it('returns flags when present', () => {
    expect(rollFlags({ modifierFlags: 'dropped' })).toBe('dropped');
    expect(rollFlags({ modifierFlags: '' })).toBe('');
    expect(rollFlags({})).toBe('');
  });
});

describe('isCritRoll table', () => {
  const table: Array<[number, number, boolean, string | null]> = [
    [20, 20, true, 'success'],
    [1, 20, true, 'fail'],
    [10, 20, true, null],
    [6, 6, true, 'success'],
    [1, 6, true, 'fail'],
    [4, 4, true, 'success'],
    [1, 4, true, 'fail'],
    [12, 12, true, 'success'],
    [100, 100, true, 'success'],
    [1, 100, true, 'fail'],
    [20, 20, false, null],
    [1, 20, false, null],
    [2, 3, true, null],
    [3, 3, true, null],
    [0, 0, true, null],
    [1, 2, true, null],
  ];
  it.each(table)('value=%i sides=%i used=%s => %s', (value, sides, used, expected) => {
    expect(isCritRoll(value, sides, used)).toBe(expected);
  });
});

describe('buildDie variations', () => {
  it('d100 value 100 shows 00', () => {
    const html = buildDie(100, 100, 'd100');
    expect(html).toContain('>00<');
    expect(html).toContain('d100');
  });
  it('d100 edge sides >=100 uses d100 class', () => {
    expect(buildDie(5, 120, 'd120')).toContain('d100');
    expect(buildDie(5, 200, 'd200')).toContain('d100');
  });
  it('normal sides use exact class', () => {
    expect(buildDie(3, 6, 'd6', 'rolling')).toContain('class="die d6 rolling"');
    expect(buildDie(4, 20, 'd20', 'settled')).toContain('class="die d20 settled"');
  });
  it('encodes data attributes', () => {
    const html = buildDie(15, 20, 'd20', 'settled');
    expect(html).toContain('data-sides="20"');
    expect(html).toContain('data-value="15"');
    expect(html).toContain('>d20<');
  });
  it('handles no extraClass', () => {
    const html = buildDie(2, 4, 'd4');
    expect(html).toContain('die d4');
  });
});
