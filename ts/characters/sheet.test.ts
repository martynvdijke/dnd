import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setCurrentChar, setCurrentTab } from '../lib/state';
import * as state from '../lib/state';
import { renderStepper, autoSaveField, stepperField, updateField, editStepperValue, renderSheet } from './sheet';

vi.mock('../lib/api', () => ({ api: vi.fn(), getApiToken: () => '' }));
vi.mock('./stats', () => ({ renderStats: vi.fn(), renderXPBar: vi.fn(() => '') }));
vi.mock('./combat', () => ({ renderCombat: vi.fn() }));
vi.mock('../dice', () => ({ renderDiceTab: vi.fn() }));
vi.mock('../lib/save', () => ({ markDirty: vi.fn(), isDirty: () => false, isSaving: () => false, saveCharacter: vi.fn() }));
vi.mock('./resources', () => ({ wealthTotalGp: () => 0 }));
// sheet imports save module which uses api — already mocked

describe('renderStepper', () => {
  it('returns stepper span with field/value/delta and aria', () => {
    const html = renderStepper('ac', 15, 1, 0, 20, 'AC');
    expect(html).toContain('stepper');
    expect(html).toContain('ac');
    expect(html).toContain('>15<');
    expect(html).toContain('aria-label="Increase AC"');
    expect(html).toContain('aria-label="Decrease AC"');
    expect(html).toContain("stepperField('ac'");
  });

  it('handles min/max undefined', () => {
    const html = renderStepper('speed', 30, 5);
    expect(html).toContain('undefined');
    expect(html).toContain('30');
  });

  it('applies size class lg', () => {
    expect(renderStepper('hp_max', 20, 1, 1, undefined, 'HP', 'lg')).toContain('stepper-lg');
  });

  it('applies sm as currency-stepper', () => {
    expect(renderStepper('gp', 5, 1, 0, undefined, 'GP', 'sm')).toContain('currency-stepper');
  });

  it('defaults label to field name', () => {
    const html = renderStepper('myField', 3, 1);
    expect(html).toContain('Increase myField');
  });
});

describe('autoSaveField / updateField / stepperField', () => {
  beforeEach(() => {
    setCurrentChar({ id: 1, name: 'Test', concentrating: false, classes: [], level: 1, race: 'Human', class: 'Fighter', hp_current: 10, hp_max: 10, temp_hp: 0, ac: 10, initiative: 0, speed: 30, proficiency_bonus: 2, portrait_url: null } as any);
    document.body.innerHTML = '<div id="sheetName"></div><div id="sheetSubtitle"></div><div id="tabBar"></div><div id="sheetView"></div><div id="statsSection"></div><div id="combatSection"></div>';
    (window as any).canEditCharacter = true;
  });

  it('autoSaveField updates numeric field', () => {
    const input = document.createElement('input');
    input.value = '18';
    autoSaveField('ac', input as any);
    expect((state as any).currentChar.ac).toBe(18);
  });

  it('autoSaveField handles checkbox', () => {
    const cb = document.createElement('input');
    cb.type = 'checkbox';
    cb.checked = true;
    autoSaveField('concentrating', cb as any);
    expect((state as any).currentChar.concentrating).toBe(true);
  });

  it('updateField sets value directly', () => {
    updateField('ac', 99);
    expect((state as any).currentChar.ac).toBe(99);
  });

  it('stepperField increments with clamp min/max', () => {
    (state as any).currentChar.testVal = 5;
    stepperField('testVal', 10, 0, 8);
    expect((state as any).currentChar.testVal).toBe(8);
    stepperField('testVal', -20, 0, 10);
    expect((state as any).currentChar.testVal).toBe(0);
  });

  it('stepperField no-op when no currentChar', () => {
    setCurrentChar(null as any);
    expect(() => stepperField('ac', 1)).not.toThrow();
    // restore
    setCurrentChar({ id: 1, ac: 10 } as any);
  });
});

describe('editStepperValue', () => {
  beforeEach(() => {
    setCurrentChar({ id: 1, hp_max: 20 } as any);
    document.body.innerHTML = '';
  });

  it('replaces element content with input and focuses', () => {
    const span = document.createElement('span');
    span.textContent = '20';
    editStepperValue('hp_max', span as any);
    expect(span.innerHTML).toContain('stepper-inline-input');
    expect(span.querySelector('input')).not.toBeNull();
  });
});

describe('renderSheet lazy-loading', () => {
  const htmxProcess = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    htmxProcess.mockClear();
    (globalThis as any).htmx = { process: htmxProcess };
    (window as any).renderCrafting = vi.fn();
    (window as any).renderDetails = vi.fn();
    (window as any).renderJournal = vi.fn();
    (window as any).updateSaveBtnState = vi.fn();
    (window as any).canEditCharacter = true;
    setCurrentChar({ id: 7, name: 'Hero', race: 'Human', class: 'Fighter', level: 3, hp_current: 10, hp_max: 20, ac: 15, portrait_url: null, classes: [] } as any);
  });

  function baseDOM() {
    document.body.innerHTML = `
      <div id="sheetName"></div><div id="sheetSubtitle"></div>
      <div id="tabBar"></div><div id="sheetView"><h2>Sheet</h2></div>
      <div id="statsSection"></div><div id="combatSection"></div>
      <div id="spellsSection"></div><div id="inventorySection"></div>
      <div id="resourcesSection"></div><div id="featuresSection"></div>
      <div id="featsSection"></div><div id="companionsSection"></div>
      <div id="craftingSection"></div><div id="journalSection"></div>
      <div id="notesSection"></div><div id="detailsSection"></div>
      <div id="diceSection"></div><div id="partySection"></div>
    `;
  }

  it('calls window.renderJournal when currentTab is journal', () => {
    setCurrentTab('journal');
    baseDOM();
    renderSheet();
    expect((window as any).renderJournal).toHaveBeenCalled();
    expect(htmxProcess).not.toHaveBeenCalled();
  });

  it('sets hx-get/hx-trigger/hx-swap and calls htmx.process for htmxTabs (notes)', () => {
    setCurrentTab('notes');
    baseDOM();
    const notesEl = document.getElementById('notesSection')!;
    expect(notesEl.hasAttribute('hx-get')).toBe(false);
    renderSheet();
    expect(notesEl.getAttribute('hx-get')).toBe('/htmx/notes?character_id=7');
    expect(notesEl.getAttribute('hx-trigger')).toBe('load');
    expect(notesEl.getAttribute('hx-swap')).toBe('innerHTML');
    expect(htmxProcess).toHaveBeenCalledWith(notesEl);
  });

  it('does not overwrite existing hx-get on second render', () => {
    setCurrentTab('notes');
    baseDOM();
    renderSheet();
    htmxProcess.mockClear();
    renderSheet();
    // second call should not re-process because hx-get already set
    expect(htmxProcess).not.toHaveBeenCalled();
  });
});
