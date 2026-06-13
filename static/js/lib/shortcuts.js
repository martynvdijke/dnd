import { hideModal, showModal } from './dom';
import { toggleTheme } from './theme';
const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'feats', 'companions', 'crafting', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'notes', 'graph', 'analytics', 'details', 'dice'];
export function getSections() {
    return sections;
}
export function showShortcutsHelp() {
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
export function initShortcuts() {
    document.addEventListener('keydown', (e) => {
        const target = e.target;
        const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT';
        if (e.key === 'Escape') {
            hideModal();
            return;
        }
        if (isInput)
            return;
        if (e.key === '?') {
            showShortcutsHelp();
            return;
        }
        if (e.key === 't' || e.key === 'T') {
            toggleTheme();
            return;
        }
        if (e.key === 'n' && window.getCurrentView?.() === 'characters') {
            window.newChar?.();
            return;
        }
        if (e.key === 'd' && window.getCurrentView?.() !== 'sheet') {
            window.showView?.('dice');
            window.renderDiceTab?.();
            setTimeout(() => {
                const input = document.getElementById('diceExpr');
                if (input)
                    input.focus();
            }, 100);
            return;
        }
        if (e.key === 'p') {
            window.showParty?.();
            return;
        }
        if (e.key === 'c') {
            window.showCompendium?.();
            return;
        }
        if (e.key === '/' && window.getCurrentView?.() === 'characters') {
            e.preventDefault();
            const search = document.querySelector('#charSearch');
            if (search)
                search.focus();
            return;
        }
        if (window.getCurrentView?.() === 'sheet') {
            const num = parseInt(e.key);
            if (num >= 1 && num <= 9) {
                const idx = num - 1;
                if (idx < sections.length) {
                    window.switchTab?.(sections[idx]);
                }
            }
        }
    });
}
