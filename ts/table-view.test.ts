import { beforeEach, describe, expect, it, vi } from 'vitest';

const { api } = vi.hoisted(() => ({ api: vi.fn() }));
vi.mock('./lib/api', () => ({ api }));
vi.mock('./lib/state', () => ({ currentChar: null, currentCampaign: null, currentUser: null }));

import { hideTable, toggleTableViewDock } from './table-view';

function renderTable(): HTMLElement {
  document.body.innerHTML = `
    <div id="tableView" class="table-view-docked">
      <button data-testid="table-view-dock-toggle" title="Expand table view"></button>
      <div id="tableBody"></div>
    </div>`;
  return document.getElementById('tableView') as HTMLElement;
}

describe('live table HUD dock state', () => {
  beforeEach(() => {
    api.mockReset();
    document.body.className = '';
  });

  it('toggles the expanded class and button label', () => {
    const view = renderTable();
    const toggle = view.querySelector('[data-testid="table-view-dock-toggle"]') as HTMLElement;

    toggleTableViewDock();
    expect(view.classList.contains('table-view-expanded')).toBe(true);
    expect(toggle.title).toBe('Dock table view');

    toggleTableViewDock();
    expect(view.classList.contains('table-view-expanded')).toBe(false);
    expect(toggle.title).toBe('Expand table view');
  });

  it('clears the HUD body flag when hidden', () => {
    const view = renderTable();
    document.body.classList.add('table-hud-open');
    hideTable();
    expect(view.style.display).toBe('none');
    expect(document.body.classList.contains('table-hud-open')).toBe(false);
  });

  it('keeps the docked class alongside expanded mode', () => {
    const view = renderTable();
    toggleTableViewDock();
    expect(view.classList.contains('table-view-docked')).toBe(true);
    expect(view.classList.contains('table-view-expanded')).toBe(true);
  });
});
