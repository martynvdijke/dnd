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

export async function api(method: string, path: string, body?: any): Promise<any> {
  showLoading();
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  if (apiToken && method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
    headers['Authorization'] = `Bearer ${apiToken}`;
  }
  const opts: RequestInit = { method, headers, credentials: 'include' };
  if (body !== undefined) opts.body = JSON.stringify(body);
  try {
    const res = await fetch(path, opts);
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(err.error || 'Request failed');
    }
    return res.json();
  } finally {
    hideLoading();
  }
}
