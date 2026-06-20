import { describe, it, expect } from 'vitest';
import { parseSides, rollValue, rollUsed, rollFlags, buildDie } from './dice';

describe('parseSides', () => {
  it('extracts sides from a die label', () => {
    expect(parseSides('d20')).toBe(20);
    expect(parseSides('d6')).toBe(6);
    expect(parseSides('d4')).toBe(4);
    expect(parseSides('d100')).toBe(100);
  });

  it('is case-insensitive', () => {
    expect(parseSides('D20')).toBe(20);
    expect(parseSides('D6')).toBe(6);
  });

  it('returns 0 for invalid labels', () => {
    expect(parseSides('foo')).toBe(0);
    expect(parseSides('')).toBe(0);
    expect(parseSides('20')).toBe(0);
    expect(parseSides('2d6')).toBe(0);
  });
});

describe('rollValue', () => {
  it('returns the number directly for plain number rolls', () => {
    expect(rollValue(15)).toBe(15);
    expect(rollValue(0)).toBe(0);
    expect(rollValue(-3)).toBe(-3);
  });

  it('extracts value from roll objects', () => {
    expect(rollValue({ value: 20, useInTotal: true })).toBe(20);
    expect(rollValue({ value: 8, useInTotal: false })).toBe(8);
  });

  it('returns 0 for objects without value', () => {
    expect(rollValue({})).toBe(0);
    expect(rollValue(null)).toBe(0);
    expect(rollValue(undefined)).toBe(0);
  });
});

describe('rollUsed', () => {
  it('returns true for plain numbers', () => {
    expect(rollUsed(15)).toBe(true);
    expect(rollUsed(0)).toBe(true);
  });

  it('returns true when useInTotal is not false', () => {
    expect(rollUsed({ value: 10, useInTotal: true })).toBe(true);
    expect(rollUsed({ value: 10 })).toBe(true);
    expect(rollUsed({ value: 10, useInTotal: undefined })).toBe(true);
  });

  it('returns false when useInTotal is explicitly false', () => {
    expect(rollUsed({ value: 5, useInTotal: false })).toBe(false);
  });
});

describe('rollFlags', () => {
  it('returns empty string for plain numbers', () => {
    expect(rollFlags(15)).toBe('');
    expect(rollFlags(0)).toBe('');
  });

  it('returns modifierFlags from roll objects', () => {
    expect(rollFlags({ value: 10, modifierFlags: 'dropped' })).toBe('dropped');
    expect(rollFlags({ value: 10, modifierFlags: 'rerolled' })).toBe('rerolled');
  });

  it('returns empty string when modifierFlags is missing', () => {
    expect(rollFlags({ value: 10 })).toBe('');
    expect(rollFlags({ value: 10, modifierFlags: '' })).toBe('');
  });
});

describe('buildDie', () => {
  it('produces a die-wrapper with label and die element', () => {
    const html = buildDie(15, 20, 'd20', 'settled');
    expect(html).toContain('die-wrapper');
    expect(html).toContain('die-label');
    expect(html).toContain('>d20<');
    expect(html).toContain('class="die d20 settled"');
  });

  it('displays the value in a die-value span', () => {
    const html = buildDie(7, 8, 'd8', 'rolling');
    expect(html).toContain('die-value');
    expect(html).toContain('>7<');
  });

  it('shows "00" for d100 value of 100', () => {
    const html = buildDie(100, 100, 'd100', 'settled');
    expect(html).toContain('>00<');
  });

  it('shows normal numbers for d100 values under 100', () => {
    const html = buildDie(50, 100, 'd100', 'settled');
    expect(html).toContain('>50<');
    expect(html).not.toContain('>00<');
  });

  it('includes data-sides and data-value attributes', () => {
    const html = buildDie(12, 12, 'd12', 'settled');
    expect(html).toContain('data-sides="12"');
    expect(html).toContain('data-value="12"');
  });

  it('applies extra classes correctly', () => {
    const html = buildDie(20, 20, 'd20', 'settled dice-crit-success');
    expect(html).toContain('dice-crit-success');
    expect(html).toContain('settled');
  });

  it('uses d100 class for sides >= 100', () => {
    const html = buildDie(30, 100, 'd100', 'rolling');
    expect(html).toContain('class="die d100 rolling"');
  });

  it('defaults extraClass to empty', () => {
    const html = buildDie(3, 4, 'd4');
    expect(html).toContain('class="die d4 "');
  });
});
