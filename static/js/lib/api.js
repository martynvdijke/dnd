import { showLoading, hideLoading } from './dom';
let csrfToken = '';
export function setCsrfToken(token) {
    csrfToken = token;
}
export function getCsrfToken() {
    return csrfToken;
}
export async function api(method, path, body) {
    showLoading();
    const headers = { 'Content-Type': 'application/json' };
    if (csrfToken)
        headers['X-CSRF-Token'] = csrfToken;
    const opts = { method, headers, credentials: 'include' };
    if (body !== undefined)
        opts.body = JSON.stringify(body);
    try {
        const res = await fetch(path, opts);
        if (!res.ok) {
            const err = await res.json().catch(() => ({ error: res.statusText }));
            throw new Error(err.error || 'Request failed');
        }
        return res.json();
    }
    finally {
        hideLoading();
    }
}
