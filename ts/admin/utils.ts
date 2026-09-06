// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, toast, showModal, hideModal } from '../lib/dom';
import { api, currentUser, setApiToken } from './state';
export function capitalize(s: string): string { return s.charAt(0).toUpperCase() + s.slice(1); }
export function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
expose('logout', async function () {
  await api('POST', '/api/logout');
  setApiToken('');
  localStorage.removeItem(`villum-api-token-${currentUser?.username || 'admin'}`);
  window.location.href = '/login';
});
