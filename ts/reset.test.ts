import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

const RESET_DOM = `
  <div id="msg" class="alert d-none"></div>
  <form id="resetForm">
    <input id="newPassword" value="s3cret12">
    <input id="confirmPassword" value="s3cret12">
    <button type="submit">Reset Password</button>
  </form>
  <div id="resetUnavailable" class="d-none">
    <p class="fw-semibold mb-1">Password reset is disabled on this deployment.</p>
    <p class="mb-2">Password login is retired, so a reset here would be refused.</p>
    <p class="mb-0 small">To restore password reset, set <code>ALLOW_PASSWORD_FALLBACK=true</code> in the server environment and restart.</p>
  </div>
  <div class="text-center mt-3"><a href="/login">Back to login</a></div>`;

/** What GET /api/auth/oidc/status answers; null models an unreachable endpoint. */
type Status = { enabled: boolean; password_enabled: boolean } | null;

type ResetStub = {
  posts: () => Array<{ url: string; init?: RequestInit }>;
  body: (call: { init?: RequestInit }) => unknown;
};

function stubFetch(status: Status): ResetStub {
  const calls: Array<{ url: string; init?: RequestInit }> = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
    const target = String(url);
    calls.push({ url: target, init });
    if (target.includes('/api/auth/oidc/status')) {
      return status ? { ok: true, json: async () => status } : { ok: false, json: async () => ({}) };
    }
    if (target.includes('/api/reset-password')) {
      // A stale or tampered token: reported by the page's existing error path.
      return { ok: false, status: 400, statusText: 'Bad Request', json: async () => ({ error: 'invalid or expired token' }) };
    }
    return { ok: false, json: async () => ({}) };
  }));
  return {
    posts: () => calls.filter(c => c.url.includes('/api/reset-password')),
    body: (call) => JSON.parse(String(call.init?.body)),
  };
}

async function importReset(status: Status) {
  document.body.innerHTML = RESET_DOM;
  // The served page carries `onsubmit="return false"`; keep dispatched submit
  // events from triggering navigation in happy-dom.
  document.getElementById('resetForm')!.addEventListener('submit', (e) => e.preventDefault());
  const stub = stubFetch(status);
  vi.resetModules();
  const module = await import('./reset');
  await module.resetReady;
  return { module, ...stub };
}

function hidden(id: string): boolean {
  return document.getElementById(id)!.classList.contains('d-none');
}

function submitResetForm() {
  document.getElementById('resetForm')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
}

function setUrl(search = '') {
  (window as any).happyDOM.setURL(`http://localhost/reset-password${search}`);
}

beforeEach(() => {
  vi.unstubAllGlobals();
  setUrl('?token=stale-token');
});

afterEach(() => {
  setUrl();
});

describe('password reset policy', () => {
  it('replaces the form with a notice when password login is retired', async () => {
    const { posts } = await importReset({ enabled: true, password_enabled: false });
    expect(hidden('resetForm')).toBe(true);
    expect(hidden('resetUnavailable')).toBe(false);

    // The notice must name the consequence and the flag that reverses it.
    const notice = document.getElementById('resetUnavailable')!.textContent!;
    expect(notice).toContain('retired');
    expect(notice).toContain('ALLOW_PASSWORD_FALLBACK');
    expect(notice).toContain('restart');

    // No handler at all: a dispatched submit sends nothing.
    submitResetForm();
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(posts()).toHaveLength(0);
    expect(hidden('msg')).toBe(true);
  });

  it('keeps the form when password login is offered', async () => {
    await importReset({ enabled: true, password_enabled: true });
    expect(hidden('resetForm')).toBe(false);
    expect(hidden('resetUnavailable')).toBe(true);
  });

  it('still reports a rejected token through the page message', async () => {
    const { posts, body } = await importReset({ enabled: true, password_enabled: true });
    submitResetForm();
    await vi.waitFor(() => expect(posts()).toHaveLength(1));
    expect(posts()[0].init?.method).toBe('POST');
    expect(body(posts()[0])).toEqual({ token: 'stale-token', new_password: 's3cret12' });

    await vi.waitFor(() => expect(hidden('msg')).toBe(false));
    expect(document.getElementById('msg')!.textContent).toContain('invalid or expired token');
  });

  it('still reports a missing token', async () => {
    setUrl();
    const { posts } = await importReset({ enabled: true, password_enabled: true });
    submitResetForm();
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(posts()).toHaveLength(0);
    expect(document.getElementById('msg')!.textContent).toContain('Missing reset token');
  });

  it('keeps the form when the status lookup fails', async () => {
    const { posts } = await importReset(null);
    expect(hidden('resetForm')).toBe(false);
    expect(hidden('resetUnavailable')).toBe(true);
    submitResetForm();
    await vi.waitFor(() => expect(posts()).toHaveLength(1));
  });
});
