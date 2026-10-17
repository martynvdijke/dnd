import { showLoading, hideLoading } from './dom';

let csrfToken = '';
let apiToken = '';

export function setCsrfToken(token: string): void {
  csrfToken = token;
}

export function getCsrfToken(): string {
  return csrfToken;
}

export function setApiToken(token: string): void {
  apiToken = token;
}

export function getApiToken(): string {
  return apiToken;
}

export function clearApiToken(): void {
  apiToken = '';
}

type UnauthorizedHandler = () => Promise<void> | void;

let onUnauthorized: UnauthorizedHandler | null = null;

// setUnauthorizedHandler registers a callback that re-provisions the API token.
// The app wires this to the bootstrap so a stale token (rotated or revoked on
// another device) self-heals instead of wedging every mutation on a 401.
export function setUnauthorizedHandler(handler: UnauthorizedHandler | null): void {
  onUnauthorized = handler;
}

function isMutation(method: string): boolean {
  const m = method.toUpperCase();
  return m !== 'GET' && m !== 'HEAD' && m !== 'OPTIONS';
}

function sendRequest(method: string, path: string, body?: unknown): Promise<Response> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  if (apiToken && isMutation(method)) {
    headers['Authorization'] = `Bearer ${apiToken}`;
  }
  const opts: RequestInit = { method, headers, credentials: 'include' };
  if (body !== undefined) opts.body = JSON.stringify(body);
  return fetch(path, opts);
}

export async function api<T = any>(method: string, path: string, body?: unknown): Promise<T> {
  showLoading();
  try {
    let res = await sendRequest(method, path, body);
    // GETs are exempt from the API-token check, so a read can succeed while a
    // mutation 401s with a stale token. Re-provision once and retry rather than
    // making the user reload the page.
    if (res.status === 401 && isMutation(method) && apiToken && onUnauthorized) {
      try {
        await onUnauthorized();
      } catch {
        // Keep the original 401 if re-provisioning fails.
      }
      res = await sendRequest(method, path, body);
    }
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw Object.assign(new Error(err.error || (err as any).message || 'Request failed'), {
        status: res.status,
        payload: err,
      });
    }
    return res.json();
  } finally {
    hideLoading();
  }
}
