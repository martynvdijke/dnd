/**
 * Tests for entity-picker module.
 */
import { describe, it, expect, beforeEach, afterEach } from 'vitest';

// Load after DOM setup
let showEntityPicker: Function;
let hideEntityPicker: Function;

beforeEach(async () => {
  document.body.innerHTML = '<div id="app"></div>';
  const mod = await import('./entity-picker');
  showEntityPicker = mod.showEntityPicker;
  hideEntityPicker = mod.hideEntityPicker;
});

afterEach(() => {
  const overlay = document.getElementById('entityPickerOverlay');
  if (overlay) overlay.remove();
});

describe('entity picker', () => {
  it('creates overlay and panel on show', () => {
    showEntityPicker('Link Entity', () => {});
    const overlay = document.getElementById('entityPickerOverlay');
    expect(overlay).toBeTruthy();
    expect(overlay!.style.display).toBe('flex');
    const panel = document.getElementById('entityPickerPanel');
    expect(panel).toBeTruthy();
  });

  it('sets the title in the header', () => {
    showEntityPicker('Pick a Target', () => {});
    const title = document.querySelector('.entity-picker-title');
    expect(title?.textContent).toBe('Pick a Target');
  });

  it('hides overlay on hideEntityPicker', () => {
    showEntityPicker('Test', () => {});
    hideEntityPicker();
    const overlay = document.getElementById('entityPickerOverlay');
    expect(overlay?.style.display).toBe('none');
  });

  it('renders type filter chips', () => {
    showEntityPicker('Test', () => {});
    const chips = document.querySelectorAll('.entity-picker-chip');
    expect(chips.length).toBeGreaterThan(10);
    // First chip should be "All"
    expect(chips[0].textContent).toContain('All');
  });

  it('renders an empty state initially', () => {
    showEntityPicker('Test', () => {});
    const results = document.getElementById('entityPickerResults');
    expect(results?.textContent).toContain('Start typing to search');
  });

  it('clicks backdrop to close', () => {
    showEntityPicker('Test', () => {});
    const overlay = document.getElementById('entityPickerOverlay')!;
    overlay.click();
    expect(overlay.style.display).toBe('none');
  });
});
