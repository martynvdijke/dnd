import type { BottomSheetConfig } from './types';
import { expose } from './lib/expose';

let activeSheet: HTMLElement | null = null;
let startY = 0;
let currentY = 0;

export function openBottomSheet(config: BottomSheetConfig): void {
  const isMobile = window.innerWidth <= 768;
  if (!isMobile) {
    (window as any).showModal?.(config.title, config.content);
    return;
  }

  const existing = document.getElementById('bottom-sheet-' + config.id);
  if (existing) {
    existing.remove();
  }

  const sheet = document.createElement('div');
  sheet.id = 'bottom-sheet-' + config.id;
  sheet.className = 'bottom-sheet';
  sheet.innerHTML = `
    <div class="bottom-sheet-backdrop"></div>
    <div class="bottom-sheet-panel">
      <div class="bottom-sheet-handle"><div class="bottom-sheet-handle-bar"></div></div>
      <div class="bottom-sheet-header">
        <h5 class="bottom-sheet-title">${config.title}</h5>
        <button class="btn btn-sm btn-outline-secondary bottom-sheet-close">&times;</button>
      </div>
      <div class="bottom-sheet-body">${config.content}</div>
    </div>
  `;
  document.body.appendChild(sheet);
  activeSheet = sheet;

  const panel = sheet.querySelector('.bottom-sheet-panel') as HTMLElement;
  const backdrop = sheet.querySelector('.bottom-sheet-backdrop') as HTMLElement;
  const handle = sheet.querySelector('.bottom-sheet-handle') as HTMLElement;
  const closeBtn = sheet.querySelector('.bottom-sheet-close') as HTMLElement;

  requestAnimationFrame(() => {
    sheet.classList.add('open');
  });

  const dismiss = () => {
    sheet.classList.remove('open');
    setTimeout(() => {
      sheet.remove();
      activeSheet = null;
      config.onDismiss?.();
    }, 300);
  };

  backdrop.addEventListener('click', dismiss);
  closeBtn.addEventListener('click', dismiss);

  let dragging = false;
  handle.addEventListener('touchstart', (e) => {
    startY = e.touches[0].clientY;
    dragging = true;
  }, { passive: true });
  document.addEventListener('touchmove', (e) => {
    if (!dragging) return;
    currentY = e.touches[0].clientY;
    const diff = currentY - startY;
    if (diff > 0) {
      panel.style.transform = `translateY(${diff}px)`;
    }
  }, { passive: true });
  document.addEventListener('touchend', () => {
    if (!dragging) return;
    dragging = false;
    const diff = currentY - startY;
    if (diff > 100) {
      dismiss();
    } else {
      panel.style.transform = '';
    }
    startY = 0;
    currentY = 0;
  }, { passive: true });
}

export function closeBottomSheet(): void {
  if (activeSheet) {
    activeSheet.classList.remove('open');
    setTimeout(() => {
      activeSheet?.remove();
      activeSheet = null;
    }, 300);
  }
}

expose('openBottomSheet', openBottomSheet);
expose('closeBottomSheet', closeBottomSheet);
