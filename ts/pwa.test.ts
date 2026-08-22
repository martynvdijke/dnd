import { describe, it, expect } from 'vitest';
import { pushSupported, urlBase64ToUint8Array, subscribeToPush, unsubscribeFromPush } from './pwa';

describe('urlBase64ToUint8Array', () => {
  it('round-trips bytes through base64url encoding', () => {
    // Bytes chosen so standard base64 contains +/-/_ chars and padding.
    const raw = new Uint8Array([251, 255, 191, 239, 190, 1, 2, 3]);
    const b64 = btoa(String.fromCharCode(...raw));
    const urlSafe = b64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

    const out = urlBase64ToUint8Array(urlSafe);
    expect(out).toBeInstanceOf(Uint8Array);
    expect(Array.from(out)).toEqual(Array.from(raw));
  });

  it('decodes a VAPID-key-sized payload', () => {
    // 65-byte uncompressed P-256 point shape (0x04 prefix), all bytes distinct-ish
    const raw = new Uint8Array(65);
    raw[0] = 4;
    for (let i = 1; i < 65; i++) raw[i] = (i * 7) % 256;
    const b64 = btoa(String.fromCharCode(...raw));
    const urlSafe = b64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

    const out = urlBase64ToUint8Array(urlSafe);
    expect(out.length).toBe(65);
    expect(out[0]).toBe(4);
  });
});

describe('pushSupported', () => {
  it('returns false when service worker container lacks push APIs', () => {
    // happy-dom's navigator has no serviceWorker by default
    expect(pushSupported()).toBe(false);
  });

  it('returns true when all push APIs are present', () => {
    const nav = navigator as any;
    const w = window as any;
    nav.serviceWorker = {};
    w.PushManager = function PushManagerStub() {};
    w.Notification = { requestPermission: async () => 'granted' };
    try {
      expect(pushSupported()).toBe(true);
    } finally {
      delete nav.serviceWorker;
      delete w.PushManager;
      delete w.Notification;
    }
  });
});

describe('subscribe/unsubscribe guards', () => {
  it('subscribeToPush resolves unavailable without push support', async () => {
    expect(await subscribeToPush()).toBe('unavailable');
  });

  it('unsubscribeFromPush resolves false without push support', async () => {
    expect(await unsubscribeFromPush()).toBe(false);
  });
});
