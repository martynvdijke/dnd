// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, toast } from '../lib/dom';
import { api } from './state';
import { renderError } from '../lib/errors';

function formatDate(value: string): string {
  const d = new Date(value);
  if (isNaN(d.getTime())) return value;
  return d.toLocaleString();
}

async function loadApiTokens() {
  try {
    const tokens = await api('GET', '/api/tokens');
    const tbody = document.getElementById('apiTokensTable');
    if (!tbody) return;
    tbody.innerHTML = '';
    if (!Array.isArray(tokens) || tokens.length === 0) {
      tbody.innerHTML = '<tr><td colspan="7" class="text-muted text-center py-3">No API tokens yet. Create one to authorize application writes.</td></tr>';
      return;
    }
    for (const t of tokens) {
      const revoked = !!t.revoked_at;
      const expired = !!t.expires_at && new Date(t.expires_at) <= new Date();
      const status = revoked ? '<span class="badge bg-danger">Revoked</span>'
        : expired ? '<span class="badge bg-warning text-dark">Expired</span>'
        : '<span class="badge bg-success">Active</span>';
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${esc(t.name || '—')}</td>
        <td><code>${esc(t.prefix || '')}…</code></td>
        <td>${esc(formatDate(t.created_at))}</td>
        <td>${t.expires_at ? esc(formatDate(t.expires_at)) : 'Never'}</td>
        <td>${t.last_used_at ? esc(formatDate(t.last_used_at)) : '—'}</td>
        <td>${status}</td>
        <td>
          ${revoked || expired ? '' : `<button class="btn btn-sm btn-outline-warning me-1" onclick="rotateApiToken(${t.id})" title="Rotate secret"><i class="fa-solid fa-rotate me-1" aria-hidden="true"></i>Rotate</button><button class="btn btn-sm btn-outline-danger" onclick="revokeApiToken(${t.id})" title="Revoke token"><i class="fa-solid fa-ban me-1" aria-hidden="true"></i>Revoke</button>`}
        </td>`;
      tbody.appendChild(tr);
    }
  } catch {
    toast('Failed to load API tokens', true);
  }
}
function showCreateTokenForm() {
  const form = document.getElementById('createTokenForm');
  const secret = document.getElementById('newTokenSecret');
  if (form) form.style.display = 'block';
  if (secret) secret.style.display = 'none';
}
function hideCreateTokenForm() {
  const form = document.getElementById('createTokenForm');
  if (form) form.style.display = 'none';
}
async function createApiToken() {
  const nameEl = document.getElementById('newTokenName') as HTMLInputElement | null;
  const expiryEl = document.getElementById('newTokenExpiry') as HTMLInputElement | null;
  const name = nameEl?.value.trim() || '';
  const expiresInDays = Math.max(0, Number(expiryEl?.value || 0));
  try {
    const res = await api('POST', '/api/tokens', { name, expires_in_days: expiresInDays });
    const secretEl = document.getElementById('newTokenSecretValue') as HTMLInputElement | null;
    if (secretEl) secretEl.value = res.token || '';
    const secretBox = document.getElementById('newTokenSecret');
    if (secretBox) secretBox.style.display = 'block';
    hideCreateTokenForm();
    if (nameEl) nameEl.value = '';
    toast('Token created — copy the secret now');
    loadApiTokens();
  } catch {
    toast('Failed to create token', true);
  }
}
async function revokeApiToken(id: number) {
  if (!confirm('Revoke this API token? It will stop working immediately.')) return;
  try {
    await api('DELETE', `/api/tokens/${id}`);
    toast('Token revoked');
    loadApiTokens();
  } catch {
    toast('Failed to revoke token', true);
  }
}
async function rotateApiToken(id: number) {
  if (!confirm('Rotate this API token? The current secret will stop working immediately.')) return;
  try {
    const res = await api('POST', `/api/tokens/${id}/rotate`);
    const secretEl = document.getElementById('newTokenSecretValue') as HTMLInputElement | null;
    if (secretEl) secretEl.value = res.token || '';
    const secretBox = document.getElementById('newTokenSecret');
    if (secretBox) secretBox.style.display = 'block';
    toast('Token rotated — copy the new secret now');
    loadApiTokens();
  } catch {
    toast('Failed to rotate token', true);
  }
}
async function copyNewTokenSecret() {
  const el = document.getElementById('newTokenSecretValue') as HTMLInputElement | null;
  const secret = el?.value || '';
  if (!secret) { toast('No secret to copy', true); return; }
  try {
    await navigator.clipboard.writeText(secret);
    toast('Secret copied');
  } catch {
    const ta = document.createElement('textarea');
    ta.value = secret;
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand('copy'); toast('Secret copied'); } catch { toast('Failed to copy secret', true); }
    document.body.removeChild(ta);
  }
}
expose('loadApiTokens', loadApiTokens);
expose('showCreateTokenForm', showCreateTokenForm);
expose('hideCreateTokenForm', hideCreateTokenForm);
expose('createApiToken', createApiToken);
expose('revokeApiToken', revokeApiToken);
expose('rotateApiToken', rotateApiToken);
expose('copyNewTokenSecret', copyNewTokenSecret);
