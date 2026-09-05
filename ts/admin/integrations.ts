// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, toast } from '../lib/dom';
import { api, aiEndpointEditId, setAiEndpointEditId } from './state';
import { renderError } from '../lib/errors';
async function loadBackupSettings() {
  try {
    const settings = await api('GET', '/api/backup/settings');
    (document.getElementById('backupEnabled') as HTMLInputElement).checked = settings.enabled;
    (document.getElementById('backupInterval') as HTMLInputElement).value = settings.interval_days || 7;
    (document.getElementById('backupKeepCount') as HTMLInputElement).value = settings.keep_count || 7;
  } catch {}
}
expose('saveBackupSettings', async function () {
  try {
    await api('PUT', '/api/backup/settings', {
      enabled: (document.getElementById('backupEnabled') as HTMLInputElement).checked,
      interval_days: +(document.getElementById('backupInterval') as HTMLInputElement).value || 7,
      keep_count: +(document.getElementById('backupKeepCount') as HTMLInputElement).value || 7,
    });
    toast('Settings saved');
  } catch (e: any) { renderError(e); }
});
async function loadBackupList() {
  try {
    const backups = await api('GET', '/api/backup/list');
    const el = document.getElementById('backupList')!;
    el.innerHTML = backups.length > 0
      ? `<table class="table table-hover mb-0"><thead><tr><th>Name</th><th>Size</th></tr></thead><tbody>
          ${backups.map((b: any) => `<tr><td>${esc(b.name)}</td><td>${formatSize(b.size)}</td></tr>`).join('')}
        </tbody></table>`
      : '<p class="text-muted p-3">No backups yet</p>';
  } catch {}
}
function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
expose('triggerBackup', async function () {
  try {
    const result = await api('POST', '/api/backup/trigger');
    toast('Backup created: ' + result.path);
    loadBackupList();
  } catch (e: any) { renderError(e); }
});
async function loadEmailSettings() {
  try {
    const s = await api('GET', '/api/admin/email-settings');
    (document.getElementById('emailEnabled') as HTMLInputElement).checked = s.enabled;
    (document.getElementById('smtpHost') as HTMLInputElement).value = s.smtp_host || '';
    (document.getElementById('smtpPort') as HTMLInputElement).value = s.smtp_port || 587;
    (document.getElementById('smtpUsername') as HTMLInputElement).value = s.username || '';
    (document.getElementById('smtpFrom') as HTMLInputElement).value = s.from_addr || '';
    if (s.has_password) {
      (document.getElementById('smtpPassword') as HTMLInputElement).placeholder = 'Password is set (leave blank to keep)';
    }
  } catch {}
}
expose('saveEmailSettings', async function () {
  try {
    await api('POST', '/api/admin/email-settings', {
      smtp_host: (document.getElementById('smtpHost') as HTMLInputElement).value,
      smtp_port: +(document.getElementById('smtpPort') as HTMLInputElement).value || 587,
      username: (document.getElementById('smtpUsername') as HTMLInputElement).value,
      password: (document.getElementById('smtpPassword') as HTMLInputElement).value,
      from_addr: (document.getElementById('smtpFrom') as HTMLInputElement).value,
      enabled: (document.getElementById('emailEnabled') as HTMLInputElement).checked,
    });
    toast('Email settings saved');
  } catch (e: any) { renderError(e); }
});
expose('testEmailSettings', async function () {
  try {
    await api('POST', '/api/admin/email-settings', {
      smtp_host: (document.getElementById('smtpHost') as HTMLInputElement).value,
      smtp_port: +(document.getElementById('smtpPort') as HTMLInputElement).value || 587,
      username: (document.getElementById('smtpUsername') as HTMLInputElement).value,
      password: (document.getElementById('smtpPassword') as HTMLInputElement).value,
      from_addr: (document.getElementById('smtpFrom') as HTMLInputElement).value,
      enabled: (document.getElementById('emailEnabled') as HTMLInputElement).checked,
      test: true,
    });
    toast('Test email sent! Check your inbox.');
  } catch (e: any) { renderError(e); }
});
async function loadPushSettings() {
  try {
    const s = await api('GET', '/api/admin/push-settings');
    (document.getElementById('pushPublicKey') as HTMLInputElement).value = s.public_key || '';
    (document.getElementById('pushSubject') as HTMLInputElement).value = s.subject || '';
    (document.getElementById('pushLeadMinutes') as HTMLInputElement).value = s.lead_minutes ?? 60;
    (document.getElementById('pushSessionLeadDays') as HTMLInputElement).value = s.session_lead_days ?? 1;
  } catch {}
}
expose('savePushSettings', async function () {
  try {
    const s = await api('POST', '/api/admin/push-settings', {
      subject: (document.getElementById('pushSubject') as HTMLInputElement).value,
      lead_minutes: +(document.getElementById('pushLeadMinutes') as HTMLInputElement).value || undefined,
      session_lead_days: +(document.getElementById('pushSessionLeadDays') as HTMLInputElement).value,
      generate_keys: (document.getElementById('pushGenerateKeys') as HTMLInputElement).checked,
    });
    (document.getElementById('pushPublicKey') as HTMLInputElement).value = s.public_key || '';
    (document.getElementById('pushGenerateKeys') as HTMLInputElement).checked = false;
    toast('Push settings saved');
  } catch (e: any) { renderError(e); }
});
expose('testPushNotification', async function () {
  try {
    const r = await api('POST', '/api/admin/test-push');
    toast(`Test push sent to ${r.sent} subscription${r.sent === 1 ? '' : 's'}`);
  } catch (e: any) { renderError(e); }
});
async function loadUmamiSettings() {
  try {
    const s = await api('GET', '/api/admin/umami-settings');
    (document.getElementById('umamiEnabled') as HTMLInputElement).checked = s.enabled;
    (document.getElementById('umamiHostname') as HTMLInputElement).value = s.tracker_hostname || '';
    (document.getElementById('umamiWebsiteID') as HTMLInputElement).value = s.website_id || '';
    (document.getElementById('umamiShareData') as HTMLInputElement).checked = s.share_data;
    (document.getElementById('umamiAdminTracking') as HTMLInputElement).checked = s.enable_admin_tracking;
  } catch {}
}
expose('saveUmamiSettings', async function () {
  try {
    await api('POST', '/api/admin/umami-settings', {
      enabled: (document.getElementById('umamiEnabled') as HTMLInputElement).checked,
      tracker_hostname: (document.getElementById('umamiHostname') as HTMLInputElement).value,
      website_id: (document.getElementById('umamiWebsiteID') as HTMLInputElement).value,
      share_data: (document.getElementById('umamiShareData') as HTMLInputElement).checked,
      enable_admin_tracking: (document.getElementById('umamiAdminTracking') as HTMLInputElement).checked,
    });
    toast('Analytics settings saved');
  } catch (e: any) { renderError(e); }
});
async function loadOTelSettings() {
  try {
    const s = await api('GET', '/api/admin/otel-settings');
    (document.getElementById('otelEnabled') as HTMLInputElement).checked = s.enabled;
    (document.getElementById('otelEndpoint') as HTMLInputElement).value = s.endpoint || '';
  } catch {}
}
expose('saveOTelSettings', async function () {
  try {
    await api('POST', '/api/admin/otel-settings', {
      enabled: (document.getElementById('otelEnabled') as HTMLInputElement).checked,
      endpoint: (document.getElementById('otelEndpoint') as HTMLInputElement).value,
    });
    toast('Telemetry settings saved');
  } catch (e: any) { renderError(e); }
});
let dummy=0;
async function loadAIEndpoints() {
  try {
    const endpoints = await api('GET', '/api/admin/ai-endpoints');
    const tbody = document.querySelector('#aiEndpointTable tbody')!;
    tbody.innerHTML = endpoints.map((ep: any) => `
      <tr>
        <td>${esc(ep.name)}</td>
        <td><span class="badge ${ep.type === 'text' ? 'badge-primary' : 'badge-secondary'}">${ep.type}</span></td>
        <td>${esc(ep.model)}</td>
        <td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(ep.base_url)}">${esc(ep.base_url)}</td>
        <td>${ep.tags && ep.tags.length ? ep.tags.map((t: string) => `<span class="badge bg-secondary me-1">${esc(t)}</span>`).join('') : '-'}</td>
        <td>${ep.enabled ? '<span class="text-success"><i class="fa-solid fa-check"></i></span>' : '<span class="text-danger"><i class="fa-solid fa-xmark"></i></span>'}</td>
        <td>
          <button class="btn btn-outline-primary btn-sm" onclick="editAIEndpoint(${ep.id})" title="Edit"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-danger btn-sm" onclick="deleteAIEndpoint(${ep.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
          <button class="btn btn-outline-info btn-sm" onclick="testAIEndpoint(${ep.id})" title="Test Connection"><i class="fa-solid fa-flask"></i></button>
        </td>
      </tr>
    `).join('');
  } catch (e: any) { renderError(e); }
}
expose('loadAIEndpoints', loadAIEndpoints);
function toggleAIEndpointFields() {
  const type = (document.getElementById('aiEndpointType') as HTMLSelectElement).value;
  document.getElementById('aiEndpointTextFields')!.style.display = type === 'text' ? 'flex' : 'none';
  document.getElementById('aiEndpointImageSizeField')!.style.display = type === 'image' ? 'block' : 'none';
}
expose('toggleAIEndpointFields', toggleAIEndpointFields);
expose('showAddAIEndpoint', function () {
  setAiEndpointEditId(null);
  document.getElementById('aiEndpointModalTitle')!.textContent = 'Add AI Endpoint';
  (document.getElementById('aiEndpointId') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointName') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointType') as HTMLSelectElement).value = 'text';
  (document.getElementById('aiEndpointBaseURL') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointTemperature') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointMaxTokens') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointImageSize') as HTMLSelectElement).value = '';
  (document.getElementById('aiEndpointTags') as HTMLInputElement).value = '';
  (document.getElementById('aiEndpointEnabled') as HTMLInputElement).checked = true;
  (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).placeholder = 'sk-...';
  toggleAIEndpointFields();
  const modal = new (window as any).bootstrap.Modal(document.getElementById('aiEndpointModal')!);
  modal.show();
});
expose('editAIEndpoint', async function (id: number) {
  setAiEndpointEditId(id);
  try {
    const ep = await api('GET', `/api/admin/ai-endpoints/${id}`);
    document.getElementById('aiEndpointModalTitle')!.textContent = 'Edit AI Endpoint';
    (document.getElementById('aiEndpointId') as HTMLInputElement).value = String(id);
    (document.getElementById('aiEndpointName') as HTMLInputElement).value = ep.name;
    (document.getElementById('aiEndpointType') as HTMLSelectElement).value = ep.type;
    (document.getElementById('aiEndpointBaseURL') as HTMLInputElement).value = ep.base_url;
    (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).value = '';
    (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).placeholder = 'Leave blank to keep current';
    (document.getElementById('aiEndpointModel') as HTMLInputElement).value = ep.model;
    (document.getElementById('aiEndpointTemperature') as HTMLInputElement).value = ep.temperature != null ? String(ep.temperature) : '';
    (document.getElementById('aiEndpointMaxTokens') as HTMLInputElement).value = ep.max_tokens != null ? String(ep.max_tokens) : '';
    (document.getElementById('aiEndpointImageSize') as HTMLSelectElement).value = ep.image_size || '';
    (document.getElementById('aiEndpointTags') as HTMLInputElement).value = ep.tags ? ep.tags.join(', ') : '';
    (document.getElementById('aiEndpointEnabled') as HTMLInputElement).checked = ep.enabled;
    toggleAIEndpointFields();
    const modal = new (window as any).bootstrap.Modal(document.getElementById('aiEndpointModal')!);
    modal.show();
  } catch (e: any) { renderError(e); }
});
expose('saveAIEndpoint', async function () {
  const name = (document.getElementById('aiEndpointName') as HTMLInputElement).value.trim();
  const type = (document.getElementById('aiEndpointType') as HTMLSelectElement).value;
  const base_url = (document.getElementById('aiEndpointBaseURL') as HTMLInputElement).value.trim();
  const api_key = (document.getElementById('aiEndpointAPIKey') as HTMLInputElement).value;
  const model = (document.getElementById('aiEndpointModel') as HTMLInputElement).value.trim();
  const temperatureStr = (document.getElementById('aiEndpointTemperature') as HTMLInputElement).value;
  const maxTokensStr = (document.getElementById('aiEndpointMaxTokens') as HTMLInputElement).value;
  const imageSize = (document.getElementById('aiEndpointImageSize') as HTMLSelectElement).value;
  const tagsStr = (document.getElementById('aiEndpointTags') as HTMLInputElement).value;
  const enabled = (document.getElementById('aiEndpointEnabled') as HTMLInputElement).checked;
  if (!name || !type || !base_url || !model) { toast('Name, Type, Base URL, and Model are required', true); return; }
  if (!aiEndpointEditId && !api_key) { toast('API Key is required for new endpoints', true); return; }
  const body: any = { name, type, base_url, model, enabled };
  if (api_key) body.api_key = api_key;
  if (temperatureStr) body.temperature = parseFloat(temperatureStr);
  if (maxTokensStr) body.max_tokens = parseInt(maxTokensStr, 10);
  if (imageSize) body.image_size = imageSize;
  body.tags = tagsStr ? tagsStr.split(',').map((t: string) => t.trim()).filter((t: string) => t) : [];
  try {
    if (aiEndpointEditId) {
      await api('PUT', `/api/admin/ai-endpoints/${aiEndpointEditId}`, body);
      toast('Endpoint updated');
    } else {
      await api('POST', '/api/admin/ai-endpoints', body);
      toast('Endpoint created');
    }
    const modalEl = document.getElementById('aiEndpointModal')!;
    const modal = (window as any).bootstrap.Modal.getInstance(modalEl);
    if (modal) modal.hide();
    loadAIEndpoints();
  } catch (e: any) { renderError(e); }
});
expose('deleteAIEndpoint', async function (id: number) {
  if (!confirm('Delete this AI endpoint? This cannot be undone.')) return;
  try { await api('DELETE', `/api/admin/ai-endpoints/${id}`); loadAIEndpoints(); toast('Endpoint deleted'); } catch (e: any) { renderError(e); }
});
expose('testAIEndpoint', async function (id: number) {
  const btn = event?.target as HTMLElement;
  if (btn) btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i>';
  try {
    const result = await api('POST', `/api/admin/ai-endpoints/${id}/test`);
    if (result.success) toast('Connection successful (status ' + result.status + ')');
    else toast('Test failed: ' + (result.error || 'Unknown error'), true);
  } catch (e: any) { renderError(e); }
  if (btn) btn.innerHTML = '<i class="fa-solid fa-flask"></i>';
});
expose('loadBackupSettings', loadBackupSettings);
expose('loadBackupList', loadBackupList);
expose('loadEmailSettings', loadEmailSettings);
expose('loadPushSettings', loadPushSettings);
expose('loadUmamiSettings', loadUmamiSettings);
expose('loadOTelSettings', loadOTelSettings);
