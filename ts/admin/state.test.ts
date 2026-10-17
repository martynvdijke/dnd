import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { api, setApiToken, setCsrfToken, setUnauthorizedHandler } from './state';

describe('admin api client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.resetAllMocks();
    setCsrfToken('');
    setApiToken('');
    setUnauthorizedHandler(null);
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('attaches the bearer header on mutations only', async () => {
    setApiToken('vlt_admin');
    globalThis.fetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({}) }) as any;

    await api('POST', '/api/admin/thing', { a: 1 });
    const opts = (globalThis.fetch as any).mock.calls[0][1];
    expect(opts.headers['Authorization']).toBe('Bearer vlt_admin');

    await api('GET', '/api/admin/thing');
    expect((globalThis.fetch as any).mock.calls[1][1].headers['Authorization']).toBeUndefined();
  });

  it('re-provisions and retries once when a mutation 401s', async () => {
    setApiToken('vlt_stale');
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts: any) => {
      if (opts.headers.Authorization === 'Bearer vlt_stale') {
        return Promise.resolve({ ok: false, status: 401, statusText: 'Unauthorized', json: () => Promise.resolve({ error: 'invalid API token' }) });
      }
      return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve({ saved: true }) });
    }) as any;
    const refresh = vi.fn().mockImplementation(async () => { setApiToken('vlt_fresh'); });
    setUnauthorizedHandler(refresh);

    const data = await api('PUT', '/api/admin/thing/1', { a: 1 });

    expect(data).toEqual({ saved: true });
    expect(refresh).toHaveBeenCalledTimes(1);
    expect((globalThis.fetch as any).mock.calls.length).toBe(2);
    expect((globalThis.fetch as any).mock.calls[1][1].headers.Authorization).toBe('Bearer vlt_fresh');
  });

  it('throws the response error when the request fails without recovery', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      statusText: 'Bad Request',
      json: () => Promise.resolve({ error: 'nope' }),
    }) as any;

    await expect(api('POST', '/api/admin/thing', {})).rejects.toThrow('nope');
  });
});
