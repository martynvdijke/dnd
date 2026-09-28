import { describe, it, expect, vi, beforeEach } from 'vitest';

const exposed: Record<string, any> = vi.hoisted(() => ({}));

vi.mock('../lib/expose', () => ({ expose: (n: string, f: any) => { exposed[n] = f; return f; } }));
vi.mock('../lib/dom', () => ({ esc: (s: any) => String(s ?? ''), showModal: vi.fn(), hideModal: vi.fn(), toast: vi.fn() }));
vi.mock('../lib/api', () => ({ api: vi.fn(), getCsrfToken: () => '', getApiToken: () => '', clearApiToken: vi.fn() }));
vi.mock('../lib/state', () => ({ currentChar: { id: 7, name: 'Aria' }, currentUser: null, setCurrentChar: vi.fn() }));
vi.mock('../lib/refresh', () => ({ refreshChar: vi.fn() }));
vi.mock('../lib/errors', () => ({ renderError: vi.fn() }));
vi.mock('../characters/sheet', () => ({ renderSheet: vi.fn(), updateField: vi.fn() }));
vi.mock('../navigation', () => ({ showView: vi.fn() }));
vi.mock('../characters/list', () => ({ loadCharacters: vi.fn() }));
vi.mock('../file-picker', () => ({ FilePicker: { pick: vi.fn(), close: vi.fn() } }));

import './character-ops';

describe('character-ops exportPdf', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('opens the printable HTML sheet and invokes print', () => {
    const print = vi.fn();
    const addEventListener = vi.fn((ev: string, cb: () => void) => {
      if (ev === 'load') cb();
    });
    const open = vi.fn(() => ({ addEventListener, print }));
    vi.stubGlobal('open', open);

    exposed.exportPdf();

    expect(open).toHaveBeenCalledWith('/api/characters/7/print?format=html', '_blank');
    expect(print).toHaveBeenCalled();

    vi.unstubAllGlobals();
  });
});
