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

const LOGIN_DOM = `
  <div id="error" class="alert alert-danger d-none"></div>
  <form id="loginForm">
    <input id="username" value="admin">
    <input id="password" type="password" value="s3cret">
    <button type="submit">Enter</button>
  </form>
  <div id="oidcLogin" class="d-none">
    <p id="oidcOnlyNote" class="d-none"></p>
    <div id="oidcDivider"></div>
    <a id="oidcLink" href="/api/auth/oidc/login">Login with Authelia</a>
  </div>
  <div id="loginUnavailable" class="d-none">
    <p class="fw-semibold mb-1">Login is disabled on this deployment.</p>
    <p class="mb-2">Password login is turned off and Authelia is not configured, so there is no way to sign in right now.</p>
    <p class="mb-0 small">To restore access, set <code>ALLOW_PASSWORD_FALLBACK=true</code> in the server environment and restart, or configure OIDC.</p>
  </div>
  <div id="forgotBlock">
    <a href="#" id="forgotLink">Forgot password?</a>
    <div id="forgotForm" class="d-none"></div>
  </div>
  <span id="themeIcon"></span><span id="themeLabel"></span>`;

/** What GET /api/auth/oidc/status answers; null models an unreachable endpoint. */
type Status = { enabled: boolean; password_enabled: boolean } | null;

type FetchStub = { calls: Array<{ url: string; init?: RequestInit }>; loginPosts: () => Array<{ url: string; init?: RequestInit }> };

function stubFetch(status: Status): FetchStub {
  const calls: Array<{ url: string; init?: RequestInit }> = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    const target = String(url);
    calls.push({ url: target, init });
    if (target.includes('/api/check-setup')) return { ok: true, json: async () => ({ setup: true }) };
    if (target.includes('/api/user/me')) return { ok: false, json: async () => ({}) };
    if (target.includes('/api/auth/oidc/status')) {
      return status ? { ok: true, json: async () => status } : { ok: false, json: async () => ({}) };
    }
    // Password POST: fail it so the handler renders an error instead of navigating.
    if (target.includes('/api/login')) return { ok: false, json: async () => ({ error: 'Invalid credentials' }) };
    return { ok: false, json: async () => ({}) };
  }));
  return { calls, loginPosts: () => calls.filter(c => c.url.includes('/api/login')) };
}

// Importing login.ts runs init(); stub DOM + fetch first, then wait for the
// page to settle so assertions are not racing the status lookup.
async function importLogin(status: Status) {
  stubLocalStorage();
  document.body.innerHTML = LOGIN_DOM;
  // The served page carries `onsubmit="return false"`; keep dispatched submit
  // events from triggering navigation in happy-dom.
  document.getElementById('loginForm')!.addEventListener('submit', (e) => e.preventDefault());
  const stub = stubFetch(status);
  vi.resetModules();
  const module = await import('./login');
  await module.loginReady;
  return { module, ...stub };
}

function hidden(id: string): boolean {
  return document.getElementById(id)!.classList.contains('d-none');
}

/** Password inputs the user could actually type into. */
function visiblePasswordInputs(): number {
  return [...document.querySelectorAll('input[type="password"]')]
    .filter((input) => !input.closest('.d-none')).length;
}

function submitLoginForm() {
  document.getElementById('loginForm')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
}

beforeEach(() => {
  vi.unstubAllGlobals();
});

describe('login policy: password available (break-glass)', () => {
  it('shows the password form and the SSO button when both are available', async () => {
    await importLogin({ enabled: true, password_enabled: true });
    expect(hidden('loginForm')).toBe(false);
    expect(hidden('forgotBlock')).toBe(false);
    expect(hidden('oidcLogin')).toBe(false);
    // With two options the "or" rule belongs on the page.
    expect(hidden('oidcDivider')).toBe(false);
    expect(hidden('oidcOnlyNote')).toBe(true);
    expect(hidden('loginUnavailable')).toBe(true);
    expect(visiblePasswordInputs()).toBe(1);
  });

  it('keeps the SSO button hidden when OIDC is disabled, and logs in with a password', async () => {
    const { loginPosts } = await importLogin({ enabled: false, password_enabled: true });
    expect(hidden('oidcLogin')).toBe(true);
    expect(hidden('loginForm')).toBe(false);
    expect(hidden('loginUnavailable')).toBe(true);

    submitLoginForm();
    await vi.waitFor(() => expect(loginPosts()).toHaveLength(1));
    expect(JSON.parse(String(loginPosts()[0].init?.body))).toEqual({ username: 'admin', password: 's3cret' });
    expect(loginPosts()[0].init?.method).toBe('POST');
    await vi.waitFor(() => expect(hidden('error')).toBe(false));
  });
});

describe('login policy: password retired', () => {
  it('shows Authelia only when OIDC is enabled', async () => {
    await importLogin({ enabled: true, password_enabled: false });
    expect(hidden('loginForm')).toBe(true);
    expect(hidden('forgotBlock')).toBe(true);
    expect(hidden('oidcLogin')).toBe(false);
    // Offered alone, the "or" rule and its copy go away.
    expect(hidden('oidcDivider')).toBe(true);
    expect(hidden('oidcOnlyNote')).toBe(false);
    expect(hidden('loginUnavailable')).toBe(true);
    expect(visiblePasswordInputs()).toBe(0);
  });

  it('says the deployment is unreachable when OIDC is not configured either', async () => {
    await importLogin({ enabled: false, password_enabled: false });
    expect(hidden('loginUnavailable')).toBe(false);
    expect(hidden('loginForm')).toBe(true);
    expect(hidden('oidcLogin')).toBe(true);
    expect(visiblePasswordInputs()).toBe(0);
    // The message must name the break-glass flag, not just say "no".
    expect(document.getElementById('loginUnavailable')!.textContent).toContain('ALLOW_PASSWORD_FALLBACK');
  });

  it('cannot post a password when password login is retired', async () => {
    const { loginPosts } = await importLogin({ enabled: true, password_enabled: false });
    submitLoginForm();
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(loginPosts()).toHaveLength(0);
  });

  it('leaves the page usable when the status lookup fails', async () => {
    await importLogin(null);
    expect(hidden('loginForm')).toBe(false);
    expect(hidden('oidcLogin')).toBe(true);
    expect(hidden('loginUnavailable')).toBe(true);
  });
});

describe('oidcErrorText', () => {
  it('maps known SSO error codes to friendly messages', async () => {
    const { module } = await importLogin(null);
    expect(module.oidcErrorText('oidc_expired')).toContain('expired');
    expect(module.oidcErrorText('oidc_state')).toContain('verification');
    expect(module.oidcErrorText('oidc_email')).toContain('verified email');
    expect(module.oidcErrorText('bogus')).toContain('password login');
  });

  it('does not offer password login when it is retired', async () => {
    const { module } = await importLogin(null);
    expect(module.oidcErrorText('bogus', false)).not.toContain('password login');
    expect(module.oidcErrorText('bogus', false)).toContain('Authelia');
  });

  it('shows a friendly message for an SSO error redirect', async () => {
    (window as any).happyDOM.setURL('http://localhost/login?error=oidc_expired');
    try {
      await importLogin({ enabled: true, password_enabled: false });
      const err = document.getElementById('error')!;
      expect(err.classList.contains('d-none')).toBe(false);
      expect(err.textContent).toContain('expired');
    } finally {
      (window as any).happyDOM.setURL('http://localhost/');
    }
  });
});
