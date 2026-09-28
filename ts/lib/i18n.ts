// Lightweight, dependency-free i18n layer.
// ponytail: flat dictionaries + attribute scanning; upgrade to ICU/plurals only
// if a locale actually needs them.
import { expose } from './expose';

export type Locale = 'en' | 'nl';

const STORAGE_KEY = 'villum.locale';

const en: Record<string, string> = {
  'nav.characters': 'Characters',
  'nav.party': 'Party View',
  'nav.partyShort': 'Party',
  'nav.compendium': 'Compendium',
  'nav.dice': 'Dice',
  'nav.encounters': 'Encounters',
  'nav.oneshots': 'One-Shots',
  'nav.combat': 'Combat',
  'nav.timeline': 'Timeline',
  'nav.world': 'World',
  'nav.sessions': 'Sessions',
  'nav.factions': 'Factions',
  'nav.shops': 'Shops',
  'nav.knowledge': 'Knowledge',
  'nav.copilot': 'Copilot',
  'nav.admin': 'Admin',
  'nav.shared': 'Shared Links',
  'nav.more': 'More',
  'action.logout': 'Logout',
  'action.signOut': 'Sign Out',
  'search.placeholder': 'Search everything...',
  'search.aria': 'Search',
  'theme.toggle': 'Toggle dark/light mode',
  'offline.banner': 'You are offline — some features may be unavailable',
  'skip.link': 'Skip to content',
  'locale.label': 'Language',
};

const nl: Record<string, string> = {
  'nav.characters': 'Personages',
  'nav.party': 'Partijoverzicht',
  'nav.partyShort': 'Partij',
  'nav.compendium': 'Compendium',
  'nav.dice': 'Dobbelstenen',
  'nav.encounters': 'Ontmoetingen',
  'nav.oneshots': 'Losse avonturen',
  'nav.combat': 'Gevecht',
  'nav.timeline': 'Tijdlijn',
  'nav.world': 'Wereld',
  'nav.sessions': 'Sessies',
  'nav.factions': 'Facties',
  'nav.shops': 'Winkels',
  'nav.knowledge': 'Kennis',
  'nav.copilot': 'Copiloot',
  'nav.admin': 'Beheer',
  'nav.shared': 'Deelbare links',
  'nav.more': 'Meer',
  'action.logout': 'Uitloggen',
  'action.signOut': 'Afmelden',
  'search.placeholder': 'Alles doorzoeken...',
  'search.aria': 'Zoeken',
  'theme.toggle': 'Donkere/lichte modus wisselen',
  'offline.banner': 'Je bent offline — sommige functies zijn mogelijk niet beschikbaar',
  'skip.link': 'Naar inhoud',
  'locale.label': 'Taal',
};

export const dictionaries: Record<Locale, Record<string, string>> = { en, nl };

export const availableLocales: { code: Locale; label: string }[] = [
  { code: 'en', label: 'English' },
  { code: 'nl', label: 'Nederlands' },
];

let activeLocale: Locale = 'en';

// Some runtimes (Node without --localstorage-file, private browsing) don't expose
// localStorage; fall back to an in-memory store so preference still works per session.
const memoryStore: Record<string, string> = {};

function storageGet(key: string): string | null {
  try {
    const ls = (globalThis as { localStorage?: Storage }).localStorage;
    const v = ls ? ls.getItem(key) : null;
    if (typeof v === 'string') return v;
  } catch {
    // ignore
  }
  return memoryStore[key] ?? null;
}

function storageSet(key: string, value: string): void {
  try {
    const ls = (globalThis as { localStorage?: Storage }).localStorage;
    if (ls) {
      ls.setItem(key, value);
      return;
    }
  } catch {
    // ignore
  }
  memoryStore[key] = value;
}

function isLocale(value: string): value is Locale {
  return value === 'en' || value === 'nl';
}

export function detectLocale(): Locale {
  const stored = storageGet(STORAGE_KEY);
  if (stored && isLocale(stored)) return stored;
  const nav = typeof navigator !== 'undefined' ? navigator.language || '' : '';
  const prefix = nav.slice(0, 2).toLowerCase();
  return isLocale(prefix) ? prefix : 'en';
}

export function getLocale(): Locale {
  return activeLocale;
}

export function t(key: string, params?: Record<string, string | number>): string {
  const dict = dictionaries[activeLocale] || en;
  let value = dict[key] ?? en[key] ?? key;
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      value = value.split(`{${k}}`).join(String(v));
    }
  }
  return value;
}

export function applyTranslations(root: ParentNode = document): void {
  root.querySelectorAll<HTMLElement>('[data-i18n]').forEach((el) => {
    const key = el.getAttribute('data-i18n');
    if (key) el.textContent = t(key);
  });
  root.querySelectorAll<HTMLElement>('[data-i18n-placeholder]').forEach((el) => {
    const key = el.getAttribute('data-i18n-placeholder');
    if (key) el.setAttribute('placeholder', t(key));
  });
  root.querySelectorAll<HTMLElement>('[data-i18n-title]').forEach((el) => {
    const key = el.getAttribute('data-i18n-title');
    if (key) el.setAttribute('title', t(key));
  });
  root.querySelectorAll<HTMLElement>('[data-i18n-aria-label]').forEach((el) => {
    const key = el.getAttribute('data-i18n-aria-label');
    if (key) el.setAttribute('aria-label', t(key));
  });
}

export function setLocale(code: string): void {
  if (!isLocale(code)) return;
  activeLocale = code;
  storageSet(STORAGE_KEY, code);
  document.documentElement.setAttribute('lang', code);
  const select = document.getElementById('localeSelect') as HTMLSelectElement | null;
  if (select) select.value = code;
  applyTranslations();
}

/** Clear the stored preference (used for tests and a future "system default" option). */
export function resetLocale(): void {
  try {
    (globalThis as { localStorage?: Storage }).localStorage?.removeItem(STORAGE_KEY);
  } catch {
    // ignore
  }
  delete memoryStore[STORAGE_KEY];
}

export function initI18n(): void {
  activeLocale = detectLocale();
  document.documentElement.setAttribute('lang', activeLocale);
  const select = document.getElementById('localeSelect') as HTMLSelectElement | null;
  if (select) select.value = activeLocale;
  applyTranslations();
}

expose('setLocale', setLocale);
expose('applyTranslations', applyTranslations);
