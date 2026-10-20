import { describe, it, expect, vi, afterEach } from 'vitest';
import { fetchAuthStatus, passwordLoginOffered, autheliaOffered } from './auth-status';

describe('fetchAuthStatus', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('returns the parsed status', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => ({ enabled: true, password_enabled: false }) })));
    expect(await fetchAuthStatus()).toEqual({ enabled: true, password_enabled: false });
  });

  it('returns null when the endpoint is not ok', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: false, json: async () => ({}) })));
    expect(await fetchAuthStatus()).toBeNull();
  });

  it('returns null when the request fails', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => { throw new Error('offline'); }));
    expect(await fetchAuthStatus()).toBeNull();
  });
});

describe('passwordLoginOffered', () => {
  it('follows the break-glass flag when the server reports it', () => {
    expect(passwordLoginOffered({ enabled: true, password_enabled: true })).toBe(true);
    expect(passwordLoginOffered({ enabled: true, password_enabled: false })).toBe(false);
  });

  it('keeps password login when the response has no flag or no answer', () => {
    // A server that does not report the flag never retired password login.
    expect(passwordLoginOffered({ enabled: true })).toBe(true);
    expect(passwordLoginOffered(null)).toBe(true);
  });
});

describe('autheliaOffered', () => {
  it('is on only when the server says OIDC is usable', () => {
    expect(autheliaOffered({ enabled: true, password_enabled: false })).toBe(true);
    expect(autheliaOffered({ enabled: false, password_enabled: true })).toBe(false);
    expect(autheliaOffered(null)).toBe(false);
  });
});
