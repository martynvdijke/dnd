import { describe, it, expect, beforeEach, vi } from 'vitest';
import { critCelebration, animateHpChange, animateTurnChange } from './animations';

// Mock gsap to avoid actual animation in tests
vi.mock('gsap', () => {
  const timeline = {
    fromTo: vi.fn().mockReturnThis(),
    to: vi.fn().mockReturnThis(),
    set: vi.fn().mockReturnThis(),
    onComplete: null,
  };
  return {
    default: {
      timeline: vi.fn(() => {
        const tl = {
          fromTo: vi.fn().mockReturnThis(),
          to: vi.fn().mockReturnThis(),
          set: vi.fn().mockReturnThis(),
        };
        return tl;
      }),
      matchMedia: vi.fn(() => ({
        add: vi.fn(),
      })),
    },
  };
});

// Mock matchMedia for reduced-motion tests
const mockMatchMedia = (matches: boolean) => {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }));
};

describe('critCelebration', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    mockMatchMedia(false);
  });

  it('creates overlay, text, and 12 spark elements on success', () => {
    const dieEl = document.createElement('div');
    dieEl.getBoundingClientRect = () => ({ left: 100, top: 100, width: 64, height: 64, right: 164, bottom: 164, x: 100, y: 100, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(dieEl);

    critCelebration('success', 'd20', dieEl);

    expect(document.querySelector('.crit-overlay')).toBeTruthy();
    expect(document.querySelector('.crit-text')).toBeTruthy();
    expect(document.querySelectorAll('.crit-spark').length).toBe(12);
  });

  it('creates overlay, text, and 12 spark elements on fail', () => {
    const dieEl = document.createElement('div');
    dieEl.getBoundingClientRect = () => ({ left: 100, top: 100, width: 64, height: 64, right: 164, bottom: 164, x: 100, y: 100, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(dieEl);

    critCelebration('fail', 'd20', dieEl);

    expect(document.querySelector('.crit-overlay')).toBeTruthy();
    expect(document.querySelector('.crit-text')).toBeTruthy();
    expect(document.querySelectorAll('.crit-spark').length).toBe(12);
  });

  it('shows CRITICAL! text for d20 success', () => {
    const dieEl = document.createElement('div');
    dieEl.getBoundingClientRect = () => ({ left: 0, top: 0, width: 64, height: 64, right: 64, bottom: 64, x: 0, y: 0, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(dieEl);

    critCelebration('success', 'd20', dieEl);
    const textEl = document.querySelector('.crit-text') as HTMLElement;
    expect(textEl.textContent).toBe('CRITICAL!');
  });

  it('shows Max on dX! text for non-d20 success', () => {
    const dieEl = document.createElement('div');
    dieEl.getBoundingClientRect = () => ({ left: 0, top: 0, width: 64, height: 64, right: 64, bottom: 64, x: 0, y: 0, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(dieEl);

    critCelebration('success', 'd6', dieEl);
    const textEl = document.querySelector('.crit-text') as HTMLElement;
    expect(textEl.textContent).toBe('Max on d6!');
  });

  it('shows CRITICAL FAIL! text for fail', () => {
    const dieEl = document.createElement('div');
    dieEl.getBoundingClientRect = () => ({ left: 0, top: 0, width: 64, height: 64, right: 64, bottom: 64, x: 0, y: 0, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(dieEl);

    critCelebration('fail', 'd20', dieEl);
    const textEl = document.querySelector('.crit-text') as HTMLElement;
    expect(textEl.textContent).toBe('CRITICAL FAIL!');
  });
});

describe('animateHpChange', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    mockMatchMedia(false);
  });

  it('creates floating damage number for HP decrease', () => {
    const hpEl = document.createElement('div');
    hpEl.getBoundingClientRect = () => ({ left: 100, top: 100, width: 80, height: 20, right: 180, bottom: 120, x: 100, y: 100, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(hpEl);
    const barEl = document.createElement('div');
    document.body.appendChild(barEl);

    animateHpChange(hpEl, barEl, 20, 15, 20);

    const floatEl = document.querySelector('.damage-float') as HTMLElement;
    expect(floatEl).toBeTruthy();
    expect(floatEl.textContent).toBe('-5');
  });

  it('creates floating heal number for HP increase', () => {
    const hpEl = document.createElement('div');
    hpEl.getBoundingClientRect = () => ({ left: 100, top: 100, width: 80, height: 20, right: 180, bottom: 120, x: 100, y: 100, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(hpEl);
    const barEl = document.createElement('div');
    document.body.appendChild(barEl);

    animateHpChange(hpEl, barEl, 10, 18, 20);

    const floatEl = document.querySelector('.damage-float') as HTMLElement;
    expect(floatEl).toBeTruthy();
    expect(floatEl.textContent).toBe('+8');
  });

  it('does nothing when delta is zero', () => {
    const hpEl = document.createElement('div');
    document.body.appendChild(hpEl);
    const barEl = document.createElement('div');
    document.body.appendChild(barEl);

    animateHpChange(hpEl, barEl, 15, 15, 20);

    expect(document.querySelector('.damage-float')).toBeNull();
  });

  it('adds hp-bar-pulse class when HP drops below 25%', () => {
    const hpEl = document.createElement('div');
    hpEl.getBoundingClientRect = () => ({ left: 0, top: 0, width: 80, height: 20, right: 80, bottom: 20, x: 0, y: 0, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(hpEl);
    const barEl = document.createElement('div');
    document.body.appendChild(barEl);

    animateHpChange(hpEl, barEl, 10, 4, 20);

    expect(barEl.classList.contains('hp-bar-pulse')).toBe(true);
  });

  it('removes hp-bar-pulse class when HP is above 25%', () => {
    const hpEl = document.createElement('div');
    hpEl.getBoundingClientRect = () => ({ left: 0, top: 0, width: 80, height: 20, right: 80, bottom: 20, x: 0, y: 0, toJSON: () => {} }) as DOMRect;
    document.body.appendChild(hpEl);
    const barEl = document.createElement('div');
    barEl.classList.add('hp-bar-pulse');
    document.body.appendChild(barEl);

    animateHpChange(hpEl, barEl, 5, 15, 20);

    expect(barEl.classList.contains('hp-bar-pulse')).toBe(false);
  });
});

describe('animateTurnChange', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    mockMatchMedia(false);
  });

  it('adds combatant-row-active class to next row', () => {
    const prevRow = document.createElement('tr');
    const nextRow = document.createElement('tr');
    nextRow.innerHTML = '<td></td><td>Goblin</td><td></td>';

    animateTurnChange(prevRow, nextRow, true);

    expect(nextRow.classList.contains('combatant-row-active')).toBe(true);
  });

  it('adds combatant-row-dimmed class to prev row', () => {
    const prevRow = document.createElement('tr');
    prevRow.classList.add('combatant-row-active');
    const nextRow = document.createElement('tr');
    nextRow.innerHTML = '<td></td><td>Goblin</td><td></td>';

    animateTurnChange(prevRow, nextRow, false);

    expect(prevRow.classList.contains('combatant-row-dimmed')).toBe(true);
    expect(prevRow.classList.contains('combatant-row-active')).toBe(false);
  });

  it('adds turn-tint-monster class when isMonster is true', () => {
    const nextRow = document.createElement('tr');
    nextRow.innerHTML = '<td></td><td>Goblin</td><td></td>';

    animateTurnChange(null, nextRow, true);

    expect(nextRow.classList.contains('turn-tint-monster')).toBe(true);
  });

  it('does not add turn-tint-monster class when isMonster is false', () => {
    const nextRow = document.createElement('tr');
    nextRow.innerHTML = '<td></td><td>Fighter</td><td></td>';

    animateTurnChange(null, nextRow, false);

    expect(nextRow.classList.contains('turn-tint-monster')).toBe(false);
  });

  it('handles null prevRow (after re-render)', () => {
    const nextRow = document.createElement('tr');
    nextRow.innerHTML = '<td></td><td>Goblin</td><td></td>';

    // Should not throw
    animateTurnChange(null, nextRow, true);
    expect(nextRow.classList.contains('combatant-row-active')).toBe(true);
  });
});
