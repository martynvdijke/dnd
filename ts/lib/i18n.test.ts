import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { t, getLocale, setLocale, detectLocale, applyTranslations, availableLocales, dictionaries, resetLocale } from './i18n';

describe('i18n', () => {
  beforeEach(() => {
    resetLocale();
    setLocale('en');
    document.body.innerHTML = '';
  });

  afterEach(() => {
    resetLocale();
  });

  it('translates known keys for the active locale', () => {
    expect(t('nav.characters')).toBe('Characters');
    setLocale('nl');
    expect(t('nav.characters')).toBe('Personages');
    expect(getLocale()).toBe('nl');
  });

  it('interpolates parameters', () => {
    dictionaries.en['test.param'] = 'Hello {name}!';
    try {
      expect(t('test.param', { name: 'Aria' })).toBe('Hello Aria!');
    } finally {
      delete dictionaries.en['test.param'];
    }
  });

  it('falls back to the key for unknown keys', () => {
    expect(t('does.not.exist')).toBe('does.not.exist');
  });

  it('lists available locales', () => {
    expect(availableLocales.map((l) => l.code)).toEqual(['en', 'nl']);
  });

  it('detects a stored locale', () => {
    setLocale('nl');
    expect(detectLocale()).toBe('nl');
  });

  it('detects the browser locale when nothing is stored', () => {
    resetLocale();
    const original = navigator.language;
    Object.defineProperty(navigator, 'language', { value: 'nl-NL', configurable: true });
    try {
      expect(detectLocale()).toBe('nl');
    } finally {
      Object.defineProperty(navigator, 'language', { value: original, configurable: true });
    }
  });

  it('applies translations to text and placeholder attributes', () => {
    document.body.innerHTML =
      '<span data-i18n="nav.dice">Dice</span>' +
      '<input id="q" data-i18n-placeholder="search.placeholder" placeholder="Search everything...">';
    setLocale('nl');
    const span = document.querySelector('[data-i18n="nav.dice"]') as HTMLElement;
    const input = document.getElementById('q') as HTMLInputElement;
    expect(span.textContent).toBe('Dobbelstenen');
    expect(input.getAttribute('placeholder')).toBe('Alles doorzoeken...');
  });

  it('persists the locale selection', () => {
    setLocale('nl');
    expect(detectLocale()).toBe('nl');
  });
});
