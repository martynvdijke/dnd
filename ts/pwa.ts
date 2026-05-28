const SW_VERSION = 'v1';

export function registerSW(): void {
  if (!('serviceWorker' in navigator)) return;
  navigator.serviceWorker.register('/sw.js', { scope: '/' })
    .then(() => console.log('[SW] registered'))
    .catch((err) => console.warn('[SW] registration failed:', err));
}

export function captureInstallPrompt(): void {
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault();
    (window as any)._deferredInstallPrompt = e;
    const moreMenu = document.getElementById('moreMenu');
    if (moreMenu) {
      const installBtn = document.createElement('button');
      installBtn.className = 'fab-menu-item';
      installBtn.innerHTML = '<i class="fa-solid fa-download me-2" aria-hidden="true"></i>Install App';
      installBtn.onclick = () => showInstallPrompt();
      moreMenu.appendChild(installBtn);
    }
  });

  window.addEventListener('appinstalled', () => {
    (window as any)._deferredInstallPrompt = null;
    console.log('[PWA] installed');
  });
}

export function showInstallPrompt(): void {
  const prompt = (window as any)._deferredInstallPrompt;
  if (!prompt) return;
  prompt.prompt();
  prompt.userChoice.then((choice: { outcome: string }) => {
    if (choice.outcome === 'dismissed') {
      const dismissed = parseInt(localStorage.getItem('villum-install-dismissed') || '0');
      if (Date.now() - dismissed > 30 * 24 * 60 * 60 * 1000) {
        localStorage.setItem('villum-install-dismissed', String(Date.now()));
      }
    }
    (window as any)._deferredInstallPrompt = null;
  });
}

export function isOffline(): boolean {
  return !navigator.onLine;
}

(window as any).registerSW = registerSW;
(window as any).showInstallPrompt = showInstallPrompt;
(window as any).isOffline = isOffline;
