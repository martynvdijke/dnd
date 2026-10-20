import { describe, it, expect, vi, beforeEach } from 'vitest';

// happy-dom doesn't provide localStorage by default
function stubLocalStorage() {
  const store: Record<string, string> = {};
  vi.stubGlobal('localStorage', {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, val: string) => { store[key] = val; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { Object.keys(store).forEach(k => delete store[k]); },
    get length() { return Object.keys(store).length; },
    key: (i: number) => Object.keys(store)[i] ?? null,
  });
}

const SETUP_DOM = `
  <div id="error" class="alert alert-danger d-none"></div>
  <form id="setupForm">
    <input id="username" value="admin">
    <input id="password" value="s3cret12">
    <input id="confirm" value="s3cret12">
    <button type="submit">Forge the Sigil</button>
  </form>
  <div id="setupUnavailable" class="d-none">
    <p class="fw-semibold mb-1">The first admin cannot be created here.</p>
    <p class="mb-2">Password login is turned off on this deployment, and this form creates the first admin by setting a password.</p>
    <p class="mb-0 small">Create the first admin by signing in through Authelia, or set <code>ALLOW_PASSWORD_FALLBACK=true</code> in the server environment and restart to use this form.</p>
  </div>
  <span id="themeIcon"></span><span id="themeLabel"></span>`;

/** What GET /api/auth/oidc/status answers; null models an unreachable endpoint. */
type Status = { enabled: boolean; password_enabled: boolean } | null;

type FetchStub = { loginPosts: () => Array<{ url: string; init?: RequestInit }> };

function stubFetch(status: Status): FetchStub {
  const calls: Array<{ url: string; init?: RequestInit }> = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    const target = String(url);
    calls.push({ url: target, init });
    // Setup is required: the page must not bounce to /login.
    if (target.includes('/api/check-setup')) return { ok: true, json: async () => ({ setup: false }) };
    if (target.includes('/api/auth/oidc/status')) {
      return status ? { ok: true, json: async () => status } : { ok: false, json: async () => ({}) };
    }
    // Password POST: fail it so the handler renders an error instead of navigating.
    if (target.includes('/api/login')) {
      return { ok: false, status: 403, statusText: 'Forbidden', text: async () => 'password login is disabled' };
    }
    return { ok: false, json: async () => ({}) };
  }));
  return { loginPosts: () => calls.filter(c => c.url.includes('/api/login')) };
}

async function importSetup(status: Status) {
  stubLocalStorage();
  document.body.innerHTML = SETUP_DOM;
  // The served page carries `onsubmit="return false"`; keep dispatched submit
  // events from triggering navigation in happy-dom.
  document.getElementById('setupForm')!.addEventListener('submit', (e) => e.preventDefault());
  const stub = stubFetch(status);
  vi.resetModules();
  const module = await import('./setup');
  await module.setupReady;
  return { module, ...stub };
}

function hidden(id: string): boolean {
  return document.getElementById(id)!.classList.contains('d-none');
}

function submitSetupForm() {
  document.getElementById('setupForm')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
}

beforeEach(() => {
  vi.unstubAllGlobals();
});

describe('first-admin setup', () => {
  it('creates the first admin when the break-glass flag allows password login', async () => {
    const { loginPosts } = await importSetup({ enabled: true, password_enabled: true });
    expect(hidden('setupForm')).toBe(false);
    expect(hidden('setupUnavailable')).toBe(true);

    submitSetupForm();
    await vi.waitFor(() => expect(loginPosts()).toHaveLength(1));
    expect(JSON.parse(String(loginPosts()[0].init?.body))).toEqual({
      username: 'admin', password: 's3cret12', setup: true,
    });
  });

  it('still validates the form it does show', async () => {
    const { loginPosts } = await importSetup({ enabled: false, password_enabled: true });
    (document.getElementById('confirm') as HTMLInputElement).value = 'different';
    submitSetupForm();
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(loginPosts()).toHaveLength(0);
    expect(document.getElementById('error')!.textContent).toContain('do not match');
  });

  it('replaces the form with a notice when password login is retired', async () => {
    const { loginPosts } = await importSetup({ enabled: true, password_enabled: false });
    expect(hidden('setupForm')).toBe(true);
    expect(hidden('setupUnavailable')).toBe(false);
    // The notice must name both ways out: Authelia, or the break-glass flag.
    const notice = document.getElementById('setupUnavailable')!.textContent!;
    expect(notice).toContain('Authelia');
    expect(notice).toContain('ALLOW_PASSWORD_FALLBACK');

    submitSetupForm();
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(loginPosts()).toHaveLength(0);
  });

  it('keeps the form when the status lookup fails', async () => {
    await importSetup(null);
    expect(hidden('setupForm')).toBe(false);
    expect(hidden('setupUnavailable')).toBe(true);
  });
});
