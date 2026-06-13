import { hideModal, showModal } from './dom';
import { toggleTheme } from './theme';

const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'feats', 'companions', 'crafting', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'notes', 'graph', 'analytics', 'details', 'dice'];

export function getSections(): string[] {
  return sections;
}

export function showShortcutsHelp(): void {
  showModal('Keyboard Shortcuts', `
    <div class="shortcut-grid">
      <div class="d-flex justify-content-between py-1"><span><kbd>n</kbd> New Character</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>d</kbd> Dice Roller</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>p</kbd> Party View</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>c</kbd> Compendium</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>/</kbd> Search Characters</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>1</kbd>-<kbd>9</kbd> Sheet Tabs</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>Esc</kbd> Close Modal</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>?</kbd> This Help</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>T</kbd> Toggle Theme</span></div>
    </div>
  `);
}

export function initShortcuts(): void {
  document.addEventListener('keydown', (e) => {
    const target = e.target as HTMLElement;
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT';

    if (e.key === 'Escape') {
      hideModal();
      return;
    }

    if (isInput) return;

    if (e.key === '?') {
      showShortcutsHelp();
      return;
    }

    if (e.key === 't' || e.key === 'T') {
      toggleTheme();
      return;
    }

    if (e.key === 'n' && (window as any).getCurrentView?.() === 'characters') {
      (window as any).newChar?.();
      return;
    }

    if (e.key === 'd' && (window as any).getCurrentView?.() !== 'sheet') {
      (window as any).showView?.('dice');
      (window as any).renderDiceTab?.();
      setTimeout(() => {
        const input = document.getElementById('diceExpr');
        if (input) (input as HTMLInputElement).focus();
      }, 100);
      return;
    }

    if (e.key === 'p') {
      (window as any).showParty?.();
      return;
    }

    if (e.key === 'c') {
      (window as any).showCompendium?.();
      return;
    }

    if (e.key === '/' && (window as any).getCurrentView?.() === 'characters') {
      e.preventDefault();
      const search = document.querySelector<HTMLInputElement>('#charSearch');
      if (search) search.focus();
      return;
    }

    if ((window as any).getCurrentView?.() === 'sheet') {
      const num = parseInt(e.key);
      if (num >= 1 && num <= 9) {
        const idx = num - 1;
        if (idx < sections.length) {
          (window as any).switchTab?.(sections[idx]);
        }
      }
    }
  });
}
