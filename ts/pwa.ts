import { expose } from './lib/expose';
const SW_VERSION = 'v1';

export function registerSW(): void {
  if (!('serviceWorker' in navigator)) return;
  // Reload once when an updated service worker (a new release) takes
  // control, so users get the fresh bundle without a manual refresh.
  // The first-ever install also claims the page; guard on whether the
  // page was already controlled so first visits don't double-load.
  const wasControlled = !!navigator.serviceWorker.controller;
  let refreshing = false;
  navigator.serviceWorker.addEventListener('controllerchange', () => {
    if (!wasControlled || refreshing) return;
    refreshing = true;
    window.location.reload();
  });
  navigator.serviceWorker.register('/sw.js', { scope: '/' })
    .then(() => console.log('[SW] registered'))
    .catch((err) => console.warn('[SW] registration failed:', err));
}

export function captureInstallPrompt(): void {
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault();
    expose('_deferredInstallPrompt', e);
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
    expose('_deferredInstallPrompt', null);
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
    expose('_deferredInstallPrompt', null);
  });
}

export function isOffline(): boolean {
  return !navigator.onLine;
}

// ─── Web Push ───

/** True only when the browser can actually subscribe (SW + PushManager). */
export function pushSupported(): boolean {
  return (
    'serviceWorker' in navigator &&
    'PushManager' in window &&
    'Notification' in window
  );
}

/** Convert a base64url VAPID public key to the bytes pushManager expects. */
export function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(base64);
  const output = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) output[i] = raw.charCodeAt(i);
  return output;
}

type SubscribeOutcome = 'subscribed' | 'denied' | 'unavailable' | 'error';

/**
 * Opt-in flow: runs only from an explicit user action. Requests permission,
 * subscribes with the server's VAPID public key and registers the endpoint.
 */
export async function subscribeToPush(): Promise<SubscribeOutcome> {
  if (!pushSupported()) return 'unavailable';
  try {
    if (Notification.permission === 'denied') return 'denied';

    let publicKey: string | null = null;
    try {
      const res = await fetch('/api/push/vapid-public-key');
      if (!res.ok) return 'unavailable';
      const data = await res.json();
      publicKey = data.public_key || null;
    } catch {
      return 'unavailable';
    }
    if (!publicKey) return 'unavailable';

    const permission =
      Notification.permission === 'granted'
        ? 'granted'
        : await Notification.requestPermission();
    if (permission !== 'granted') return 'denied';

    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey) as BufferSource,
    });
    const res = await fetch('/api/push/subscribe', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(sub.toJSON()),
    });
    if (!res.ok) return 'error';
    return 'subscribed';
  } catch {
    return 'error';
  }
}

/** Mirrors subscribe: unregister server-side, then locally. */
export async function unsubscribeFromPush(): Promise<boolean> {
  if (!pushSupported()) return false;
  try {
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.getSubscription();
    if (!sub) return true;
    await fetch('/api/push/unsubscribe', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ endpoint: sub.endpoint }),
    });
    await sub.unsubscribe();
    return true;
  } catch {
    return false;
  }
}

expose('registerSW', registerSW);
expose('showInstallPrompt', showInstallPrompt);
expose('isOffline', isOffline);
