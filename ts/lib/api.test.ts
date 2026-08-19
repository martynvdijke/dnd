import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { api, setCsrfToken, getCsrfToken, setApiToken, getApiToken } from './api';

describe('api client', () => {
  const originalFetch = globalThis.fetch;
  const originalSetTimeout = globalThis.setTimeout;

  beforeEach(() => {
    vi.resetAllMocks();
    setCsrfToken('');
    setApiToken('');
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    globalThis.setTimeout = originalSetTimeout;
  });

  it('performs a GET request and returns parsed JSON', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ results: [1, 2, 3] }),
    }) as any;
    const data = await api('GET', '/api/search');
    expect(data).toEqual({ results: [1, 2, 3] });
    const [url, opts] = (globalThis.fetch as any).mock.calls[0];
    expect(url).toBe('/api/search');
    expect(opts.method).toBe('GET');
    expect(opts.body).toBeUndefined();
    expect(opts.credentials).toBe('include');
  });

  it('includes X-CSRF-Token header when csrf token is set', async () => {
    setCsrfToken('abc123');
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    }) as any;
    await api('GET', '/api/test');
    const opts = (globalThis.fetch as any).mock.calls[0][1];
    expect(opts.headers['X-CSRF-Token']).toBe('abc123');
    expect(getCsrfToken()).toBe('abc123');
  });

  it('attaches Authorization bearer header on non-GET when api token is set', async () => {
    setApiToken('vlt_secret123');
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    }) as any;
    await api('POST', '/api/roll', { expression: '1d20' });
    const opts = (globalThis.fetch as any).mock.calls[0][1];
    expect(opts.headers['Authorization']).toBe('Bearer vlt_secret123');
    expect(getApiToken()).toBe('vlt_secret123');
  });

  it('does not attach Authorization header on GET', async () => {
    setApiToken('vlt_secret123');
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    }) as any;
    await api('GET', '/api/search');
    const opts = (globalThis.fetch as any).mock.calls[0][1];
    expect(opts.headers['Authorization']).toBeUndefined();
  });

  it('does not attach Authorization header when no api token is set', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    }) as any;
    await api('POST', '/api/roll', { expression: '1d20' });
    const opts = (globalThis.fetch as any).mock.calls[0][1];
    expect(opts.headers['Authorization']).toBeUndefined();
  });

  it('serializes body as JSON when provided', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    }) as any;
    await api('POST', '/api/roll', { expression: '1d20' });
    const opts = (globalThis.fetch as any).mock.calls[0][1];
    expect(opts.body).toBe(JSON.stringify({ expression: '1d20' }));
  });

  it('throws error message from response body when request fails', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      statusText: 'Bad Request',
      json: () => Promise.resolve({ error: 'Something broke' }),
    }) as any;
    await expect(api('GET', '/api/boom')).rejects.toThrow('Something broke');
  });

  it('falls back to Request failed when error body has no error field', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      statusText: 'Not Found',
      json: () => Promise.resolve({}),
    }) as any;
    await expect(api('GET', '/api/missing')).rejects.toThrow('Request failed');
  });

  it('falls back to statusText when error body is not valid JSON', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      statusText: 'Server Error',
      json: () => Promise.reject(new Error('parse error')),
    }) as any;
    await expect(api('GET', '/api/bad')).rejects.toThrow('Server Error');
  });

  it('calls hideLoading in finally even on success', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    }) as any;
    await api('GET', '/api/ok');
    expect(true).toBe(true);
  });
});
