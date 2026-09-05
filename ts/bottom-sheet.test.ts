import { describe, it, expect, beforeEach, vi } from 'vitest';
import { openBottomSheet, closeBottomSheet } from './bottom-sheet';

function setMobile(isMobile: boolean) {
  Object.defineProperty(window, 'innerWidth', { value: isMobile ? 500 : 1024, writable: true, configurable: true });
}

beforeEach(() => {
  document.body.innerHTML = '';
  setMobile(true);
  vi.useFakeTimers();
});

describe('openBottomSheet mobile', () => {
  it('creates bottom sheet DOM on mobile', () => {
    openBottomSheet({ id: 'test1', title: 'Hello', content: '<p>body</p>' });
    const sheet = document.getElementById('bottom-sheet-test1');
    expect(sheet).not.toBeNull();
    expect(sheet!.innerHTML).toContain('Hello');
    expect(sheet!.innerHTML).toContain('<p>body</p>');
  });

  it('removes existing sheet with same id before creating new one', () => {
    openBottomSheet({ id: 'dup', title: 'First', content: 'a' });
    openBottomSheet({ id: 'dup', title: 'Second', content: 'b' });
    const sheets = document.querySelectorAll('#bottom-sheet-dup');
    expect(sheets.length).toBe(1);
    expect(sheets[0].innerHTML).toContain('Second');
  });

  it('close button dismisses sheet and calls onDismiss', () => {
    const onDismiss = vi.fn();
    openBottomSheet({ id: 'dismiss', title: 'T', content: 'c', onDismiss });
    const sheet = document.getElementById('bottom-sheet-dismiss')!;
    const closeBtn = sheet.querySelector('.bottom-sheet-close') as HTMLElement;
    closeBtn.click();
    // removal delayed 300ms
    expect(sheet.classList.contains('open')).toBe(false);
    vi.advanceTimersByTime(310);
    expect(document.getElementById('bottom-sheet-dismiss')).toBeNull();
    expect(onDismiss).toHaveBeenCalledOnce();
  });

  it('backdrop click dismisses', () => {
    const onDismiss = vi.fn();
    openBottomSheet({ id: 'bd', title: 'T', content: 'c', onDismiss });
    const sheet = document.getElementById('bottom-sheet-bd')!;
    const backdrop = sheet.querySelector('.bottom-sheet-backdrop') as HTMLElement;
    backdrop.click();
    vi.advanceTimersByTime(310);
    expect(document.getElementById('bottom-sheet-bd')).toBeNull();
    expect(onDismiss).toHaveBeenCalledOnce();
  });

  it('adds open class after rAF', async () => {
    vi.useRealTimers();
    openBottomSheet({ id: 'raf', title: 'T', content: 'c' });
    const sheet = document.getElementById('bottom-sheet-raf')!;
    await new Promise<void>((r) => requestAnimationFrame(() => r()));
    // allow rAF to fire
    await new Promise<void>((r) => setTimeout(r, 10));
    expect(sheet.classList.contains('open')).toBe(true);
    vi.useFakeTimers();
  });
});

describe('openBottomSheet desktop', () => {
  it('calls window.showModal instead of creating sheet', () => {
    setMobile(false);
    const showModal = vi.fn();
    (window as any).showModal = showModal;
    openBottomSheet({ id: 'desk', title: 'Desk Title', content: 'desk body' });
    expect(showModal).toHaveBeenCalledWith('Desk Title', 'desk body');
    expect(document.getElementById('bottom-sheet-desk')).toBeNull();
    delete (window as any).showModal;
  });
});

describe('closeBottomSheet', () => {
  it('removes active sheet after delay', () => {
    openBottomSheet({ id: 'close', title: 'T', content: 'c' });
    expect(document.getElementById('bottom-sheet-close')).not.toBeNull();
    closeBottomSheet();
    vi.advanceTimersByTime(310);
    expect(document.getElementById('bottom-sheet-close')).toBeNull();
  });

  it('is no-op when no active sheet', () => {
    expect(() => closeBottomSheet()).not.toThrow();
  });
});
