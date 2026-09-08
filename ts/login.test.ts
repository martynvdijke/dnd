import { describe, it, expect, vi, beforeEach } from 'vitest';

// happy-dom doesn't provide localStorage by default
function stubLocalStorage() {
  const store: Record<string, string> = {};
  (globalThis as any).localStorage = {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, val: string) => { store[key] = val; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { Object.keys(store).forEach(k => delete store[k]); },
    get length() { return Object.keys(store).length; },
    key: (i: number) => Object.keys(store)[i] ?? null,
  };
}

const LOGIN_DOM = `
  <form id="loginForm">
    <input id="username" value=""><input id="password" value="">
    <button type="submit">Enter</button>
  </form>
  <div id="error" class="d-none"></div>
  <div id="oidcLogin" class="d-none"></div>
  <span id="themeIcon"></span><span id="themeLabel"></span>`;

// Importing login.ts runs init(); stub DOM + fetch first.
async function importLogin(statusEnabled: boolean) {
  stubLocalStorage();
  document.body.innerHTML = LOGIN_DOM;
  vi.stubGlobal('fetch', vi.fn(async (url: string) => {
    if (String(url).includes('/api/check-setup')) return { ok: true, json: async () => ({ setup: true }) };
    if (String(url).includes('/api/user/me')) return { ok: false, json: async () => ({}) };
    if (String(url).includes('/api/auth/oidc/status')) return { ok: true, json: async () => ({ enabled: statusEnabled }) };
    return { ok: false, json: async () => ({}) };
  }));
  vi.resetModules();
  return import('./login');
}

beforeEach(() => {
  vi.unstubAllGlobals();
});

describe('oidcErrorText', () => {
  it('maps known SSO error codes to friendly messages', async () => {
    const m = await importLogin(false);
    expect(m.oidcErrorText('oidc_expired')).toContain('expired');
    expect(m.oidcErrorText('oidc_state')).toContain('verification');
    expect(m.oidcErrorText('oidc_email')).toContain('verified email');
    expect(m.oidcErrorText('bogus')).toContain('password login');
  });
});

describe('initOIDC', () => {
  it('keeps the SSO button hidden when OIDC is disabled', async () => {
    const m = await importLogin(false);
    await m.initOIDC();
    expect(document.getElementById('oidcLogin')?.classList.contains('d-none')).toBe(true);
  });

  it('reveals the SSO button when OIDC is enabled', async () => {
    const m = await importLogin(true);
    await m.initOIDC();
    expect(document.getElementById('oidcLogin')?.classList.contains('d-none')).toBe(false);
  });

  it('shows a friendly message for an SSO error redirect', async () => {
    stubLocalStorage();
    document.body.innerHTML = LOGIN_DOM;
    (window as any).happyDOM.setURL('http://localhost/login?error=oidc_expired');
    vi.stubGlobal('fetch', vi.fn(async (url: string) => {
      if (String(url).includes('/api/check-setup')) return { ok: true, json: async () => ({ setup: true }) };
      if (String(url).includes('/api/user/me')) return { ok: false, json: async () => ({}) };
      return { ok: false, json: async () => ({}) };
    }));
    vi.resetModules();
    const m = await import('./login');
    await m.initOIDC();
    const err = document.getElementById('error')!;
    expect(err.classList.contains('d-none')).toBe(false);
    expect(err.textContent).toContain('expired');
    (window as any).happyDOM.setURL('http://localhost/');
  });
});
