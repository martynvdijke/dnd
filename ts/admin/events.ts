// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, attrEscape, toast, showModal } from '../lib/dom';
import { api, campaignEventEditId, setCampaignEventEditId } from './state';
import { renderError } from '../lib/errors';
expose('toggleEventSourceFields', function () {
  const isIcal = (document.getElementById('sourceTypeIcal') as HTMLInputElement).checked;
  (document.getElementById('eventsIcalUrlField') as HTMLElement).style.display = isIcal ? '' : 'none';
  (document.getElementById('eventsGoogleFields') as HTMLElement).style.display = isIcal ? 'none' : '';
  (document.getElementById('eventsGoogleOnlyFields') as HTMLElement).style.display = isIcal ? 'none' : '';
});
expose('toggleEventsAuthFields', function () {
  const sa = (document.getElementById('authMethodServiceAccount') as HTMLInputElement).checked;
  (document.getElementById('eventsServiceAccountFields') as HTMLElement).style.display = sa ? '' : 'none';
  (document.getElementById('eventsOAuthFields') as HTMLElement).style.display = sa ? 'none' : '';
});
async function loadEventsSettings() {
  try {
    const s = await api('GET', '/api/admin/events-settings');
    (document.getElementById('eventsCalendarId') as HTMLInputElement).value = s.calendar_id || '';
    (document.getElementById('eventsTags') as HTMLInputElement).value = s.tags || '';
    (document.getElementById('eventsCacheTTL') as HTMLInputElement).value = s.cache_ttl_seconds || 300;
    (document.getElementById('eventsColorLabels') as HTMLInputElement).value = s.color_labels || '';
    (document.getElementById('eventsIcalUrl') as HTMLInputElement).value = s.ical_url || '';
    const sourceType = s.source_type || 'google_api';
    if (sourceType === 'ical') (document.getElementById('sourceTypeIcal') as HTMLInputElement).checked = true;
    else (document.getElementById('sourceTypeGoogleApi') as HTMLInputElement).checked = true;
    const filterMode = s.filter_mode || 'text';
    const fmRadio = document.getElementById('filterMode' + filterMode.charAt(0).toUpperCase() + filterMode.slice(1)) as HTMLInputElement;
    if (fmRadio) fmRadio.checked = true;
    const authMethod = s.auth_method || 'service_account';
    if (authMethod === 'oauth') (document.getElementById('authMethodOAuth') as HTMLInputElement).checked = true;
    else (document.getElementById('authMethodServiceAccount') as HTMLInputElement).checked = true;
    (document.getElementById('eventsCredentialsJson') as HTMLTextAreaElement).value = s.credentials_json || '';
    (document.getElementById('eventsOAuthClientId') as HTMLInputElement).value = s.oauth_client_id || '';
    (document.getElementById('eventsOAuthClientSecret') as HTMLInputElement).value = s.oauth_client_secret || '';
    (document.getElementById('eventsOAuthRefreshToken') as HTMLInputElement).value = s.oauth_refresh_token || '';
    (window as any).toggleEventSourceFields();
    (window as any).toggleEventsAuthFields();
  } catch (e: any) { renderError(e); }
}
expose('loadEventsSettings', loadEventsSettings);
expose('saveEventsSettings', async function () {
  const sourceTypeEl = document.querySelector('input[name="eventsSourceType"]:checked') as HTMLInputElement;
  const filterModeEl = document.querySelector('input[name="eventsFilterMode"]:checked') as HTMLInputElement;
  const authMethodEl = document.querySelector('input[name="eventsAuthMethod"]:checked') as HTMLInputElement;
  const body: any = {
    source_type: sourceTypeEl ? sourceTypeEl.value : 'google_api',
    ical_url: (document.getElementById('eventsIcalUrl') as HTMLInputElement).value.trim(),
    calendar_id: (document.getElementById('eventsCalendarId') as HTMLInputElement).value.trim(),
    tags: (document.getElementById('eventsTags') as HTMLInputElement).value.trim(),
    cache_ttl_seconds: parseInt((document.getElementById('eventsCacheTTL') as HTMLInputElement).value) || 300,
    color_labels: (document.getElementById('eventsColorLabels') as HTMLInputElement).value.trim(),
    filter_mode: filterModeEl ? filterModeEl.value : 'text',
    auth_method: authMethodEl ? authMethodEl.value : 'service_account',
    credentials_json: (document.getElementById('eventsCredentialsJson') as HTMLTextAreaElement).value.trim(),
    oauth_client_id: (document.getElementById('eventsOAuthClientId') as HTMLInputElement).value.trim(),
    oauth_client_secret: (document.getElementById('eventsOAuthClientSecret') as HTMLInputElement).value.trim(),
    oauth_refresh_token: (document.getElementById('eventsOAuthRefreshToken') as HTMLInputElement).value.trim(),
  };
  try { await api('PUT', '/api/admin/events-settings', body); toast('Events settings saved'); } catch (e: any) { renderError(e); }
});
expose('clearEventsCache', async function () {
  try { await api('POST', '/api/admin/events-cache/clear'); toast('Events cache cleared'); } catch (e: any) { renderError(e); }
});
async function loadEventsPublicLink() {
  try {
    const res = await api('GET', '/api/admin/events/public-link');
    const input = document.getElementById('eventsPublicLink') as HTMLInputElement;
    if (input) input.value = res.url || '';
    const img = document.getElementById('eventsQRImg') as HTMLImageElement;
    if (img) img.src = '/api/admin/events/qr';
  } catch (e: any) { renderError(e); }
}
expose('loadEventsPublicLink', loadEventsPublicLink);
expose('copyPublicLink', async function () {
  const input = document.getElementById('eventsPublicLink') as HTMLInputElement;
  try {
    await navigator.clipboard.writeText(input.value);
    const label = document.getElementById('copyLinkLabel');
    if (label) label.textContent = 'Copied!';
    setTimeout(() => { const l = document.getElementById('copyLinkLabel'); if (l) l.textContent = 'Copy link'; }, 2000);
  } catch { input.select(); document.execCommand('copy'); }
});
expose('openEventsPage', function () {
  const input = document.getElementById('eventsPublicLink') as HTMLInputElement;
  window.open(input.value, '_blank');
});
expose('downloadQR', function () {
  const img = document.getElementById('eventsQRImg') as HTMLImageElement;
  const a = document.createElement('a'); a.href = img.src; a.download = 'events-qr.png';
  document.body.appendChild(a); a.click(); a.remove();
});
expose('shareCampaignEvent', async function (id: number, slug: string) {
  try {
    const res = await api('GET', '/api/admin/events/public-link?slug=' + encodeURIComponent(slug));
    const url = res.url || '';
    const qrSrc = '/api/admin/events/qr?slug=' + encodeURIComponent(slug);
    showModal('Share Event Page', `
      <p class="text-muted">Share this campaign's public events page with players.</p>
      <div class="mb-3">
        <label class="form-label">Public URL</label>
        <div class="input-group">
          <input type="text" id="campaignShareUrl" class="form-control font-monospace" readonly value="${esc(url)}" style="font-size:0.85rem">
          <button class="btn btn-outline-primary" id="campaignCopyLinkBtn" title="Copy to clipboard"><i class="fa-solid fa-copy me-1"></i><span id="campaignCopyLinkLabel">Copy link</span></button>
        </div>
      </div>
      <div class="text-center">
        <div class="border rounded p-2 d-inline-block bg-light">
          <img id="campaignShareQRImg" src="${qrSrc}" alt="QR Code" width="160" height="160" style="image-rendering:pixelated">
        </div>
        <div class="mt-2">
          <button class="btn btn-outline-secondary" id="campaignDownloadQRBtn"><i class="fa-solid fa-download me-1"></i>Download QR</button>
        </div>
      </div>
    `);
    document.getElementById('campaignCopyLinkBtn')!.addEventListener('click', async () => {
      const input = document.getElementById('campaignShareUrl') as HTMLInputElement;
      try {
        await navigator.clipboard.writeText(input.value);
        const label = document.getElementById('campaignCopyLinkLabel');
        if (label) label.textContent = 'Copied!';
        setTimeout(() => { const l = document.getElementById('campaignCopyLinkLabel'); if (l) l.textContent = 'Copy link'; }, 2000);
      } catch { input.select(); document.execCommand('copy'); }
    });
    document.getElementById('campaignDownloadQRBtn')!.addEventListener('click', () => {
      const img = document.getElementById('campaignShareQRImg') as HTMLImageElement;
      const a = document.createElement('a'); a.href = img.src; a.download = 'events-qr-' + slug + '.png';
      document.body.appendChild(a); a.click(); a.remove();
    });
  } catch (e: any) { renderError(e); }
});
async function loadCampaignEventSettings() {
  try {
    const campaigns = await api('GET', '/api/admin/events-campaigns');
    const tbody = document.getElementById('campaignEventBody')!;
    if (!campaigns || campaigns.length === 0) {
      tbody.innerHTML = '<tr><td colspan="9" class="text-muted text-center py-3">No campaign event pages configured. Add one to create a public event page for a campaign.</td></tr>';
      return;
    }
    tbody.innerHTML = campaigns.map((c: any) => {
      const filterParts: string[] = [];
      if (c.filter_mode && c.filter_mode !== 'text') filterParts.push(c.filter_mode);
      if (c.color_labels) filterParts.push('color:' + c.color_labels);
      const sourceType = c.source_type || 'google_api';
      const sourceLabel = sourceType === 'ical' ? 'iCal' : 'GCal API';
      const sourceDetail = sourceType === 'ical' ? (c.ical_url || '(global)') : (c.calendar_id || '(global)');
      return `
      <tr>
        <td><strong>${esc(c.display_name)}</strong></td>
        <td><code>${esc(c.slug)}</code></td>
        <td style="max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(sourceDetail)}"><small class="text-muted">${sourceLabel}</small> ${esc(sourceDetail)}</td>
        <td style="max-width:120px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(c.tags || '')}">${esc(c.tags || '(global)')}</td>
        <td><small class="text-muted">${filterParts.length ? esc(filterParts.join(', ')) : 'text'}</small></td>
        <td><small class="text-muted">${sourceType === 'ical' ? '—' : (c.auth_method === 'oauth' ? 'OAuth' : 'Service Acct')}</small></td>
        <td>${c.is_active ? '<span class="text-success"><i class="fa-solid fa-check"></i></span>' : '<span class="text-muted"><i class="fa-solid fa-xmark"></i></span>'}</td>
        <td class="text-nowrap">
          <button class="btn btn-outline-success btn-sm py-0 js-share-campaign-event" data-id="${c.id}" data-slug="${attrEscape(c.slug)}" title="Share link & QR"><i class="fa-solid fa-share-nodes"></i></button>
          <a href="/events/c/${esc(c.slug)}" class="btn btn-outline-info btn-sm py-0" target="_blank" title="View public page"><i class="fa-solid fa-eye"></i></a>
        </td>
        <td class="text-nowrap">
          <button class="btn btn-outline-warning btn-sm py-0" onclick="clearCampaignCache(${c.id})" title="Clear cache for this campaign page"><i class="fa-solid fa-eraser"></i></button>
          <button class="btn btn-outline-primary btn-sm py-0" onclick="editCampaignEventSetting(${c.id})" title="Edit"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-danger btn-sm py-0" onclick="deleteCampaignEventSetting(${c.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
        </td>
      </tr>`;
    }).join('');
    const tbody2 = document.getElementById('campaignEventBody')!;
    tbody2.querySelectorAll<HTMLButtonElement>('.js-share-campaign-event').forEach(btn => {
      btn.addEventListener('click', () => (window as any).shareCampaignEvent(Number(btn.dataset.id), btn.dataset.slug || ''));
    });
  } catch (e: any) {
    const tbody = document.getElementById('campaignEventBody');
    if (tbody) tbody.innerHTML = '<tr><td colspan="9" class="text-danger text-center">Failed to load: ' + esc(e.message) + '</td></tr>';
  }
}
expose('loadCampaignEventSettings', loadCampaignEventSettings);
expose('toggleCampaignAuthFields', function () {
  const sa = (document.getElementById('campaignAuthMethodServiceAccount') as HTMLInputElement).checked;
  (document.getElementById('campaignServiceAccountFields') as HTMLElement).style.display = sa ? '' : 'none';
  (document.getElementById('campaignOAuthFields') as HTMLElement).style.display = sa ? 'none' : '';
});
expose('toggleCampaignEventSourceFields', function () {
  const isIcal = (document.getElementById('campaignSourceTypeIcal') as HTMLInputElement).checked;
  (document.getElementById('campaignIcalUrlField') as HTMLElement).style.display = isIcal ? '' : 'none';
  (document.getElementById('campaignGoogleFields') as HTMLElement).style.display = isIcal ? 'none' : '';
  (document.getElementById('campaignGoogleOnlyFields') as HTMLElement).style.display = isIcal ? 'none' : '';
});
expose('showAddCampaignEvent', function () {
  setCampaignEventEditId(null);
  (window as any).campaignSlugAuto = true;
  document.getElementById('campaignEventModalTitle')!.textContent = 'Add Campaign Event Page';
  (document.getElementById('campaignEventId') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventDisplayName') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventSlug') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventCalendarId') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventTags') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventCacheTTL') as HTMLInputElement).value = '300';
  (document.getElementById('campaignEventColorLabels') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventIcalUrl') as HTMLInputElement).value = '';
  (document.getElementById('campaignSourceTypeGoogleApi') as HTMLInputElement).checked = true;
  (document.getElementById('campaignFilterModeText') as HTMLInputElement).checked = true;
  (document.getElementById('campaignAuthMethodServiceAccount') as HTMLInputElement).checked = true;
  (document.getElementById('campaignEventCredentialsJson') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventOAuthClientId') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventOAuthClientSecret') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventOAuthRefreshToken') as HTMLInputElement).value = '';
  (document.getElementById('campaignEventIsActive') as HTMLInputElement).checked = true;
  updateCampaignSlugPreview();
  (document.getElementById('campaignServiceAccountFields') as HTMLElement).style.display = 'none';
  (document.getElementById('campaignOAuthFields') as HTMLElement).style.display = 'none';
  (window as any).toggleCampaignEventSourceFields();
  new (window as any).bootstrap.Modal(document.getElementById('campaignEventModal')!).show();
});
expose('editCampaignEventSetting', async function (id: number) {
  setCampaignEventEditId(id);
  (window as any).campaignSlugAuto = false;
  try {
    const c = await api('GET', '/api/admin/events-campaigns/' + id);
    document.getElementById('campaignEventModalTitle')!.textContent = 'Edit Campaign Event Page';
    (document.getElementById('campaignEventId') as HTMLInputElement).value = String(id);
    (document.getElementById('campaignEventDisplayName') as HTMLInputElement).value = c.display_name || '';
    (document.getElementById('campaignEventSlug') as HTMLInputElement).value = c.slug || '';
    (document.getElementById('campaignEventCalendarId') as HTMLInputElement).value = c.calendar_id || '';
    (document.getElementById('campaignEventTags') as HTMLInputElement).value = c.tags || '';
    (document.getElementById('campaignEventCacheTTL') as HTMLInputElement).value = c.cache_ttl_seconds || 300;
    (document.getElementById('campaignEventColorLabels') as HTMLInputElement).value = c.color_labels || '';
    (document.getElementById('campaignEventIcalUrl') as HTMLInputElement).value = c.ical_url || '';
    const sourceType = c.source_type || 'google_api';
    if (sourceType === 'ical') (document.getElementById('campaignSourceTypeIcal') as HTMLInputElement).checked = true;
    else (document.getElementById('campaignSourceTypeGoogleApi') as HTMLInputElement).checked = true;
    const filterMode = c.filter_mode || 'text';
    const fmRadio = document.getElementById('campaignFilterMode' + filterMode.charAt(0).toUpperCase() + filterMode.slice(1)) as HTMLInputElement;
    if (fmRadio) fmRadio.checked = true;
    const authMethod = c.auth_method || 'service_account';
    if (authMethod === 'oauth') (document.getElementById('campaignAuthMethodOAuth') as HTMLInputElement).checked = true;
    else (document.getElementById('campaignAuthMethodServiceAccount') as HTMLInputElement).checked = true;
    (document.getElementById('campaignEventCredentialsJson') as HTMLInputElement).value = c.credentials_json || '';
    (document.getElementById('campaignEventOAuthClientId') as HTMLInputElement).value = c.oauth_client_id || '';
    (document.getElementById('campaignEventOAuthClientSecret') as HTMLInputElement).value = c.oauth_client_secret || '';
    (document.getElementById('campaignEventOAuthRefreshToken') as HTMLInputElement).value = c.oauth_refresh_token || '';
    (document.getElementById('campaignEventIsActive') as HTMLInputElement).checked = c.is_active;
    updateCampaignSlugPreview();
    (window as any).toggleCampaignEventSourceFields();
    (window as any).toggleCampaignAuthFields();
    new (window as any).bootstrap.Modal(document.getElementById('campaignEventModal')!).show();
  } catch (e: any) { renderError(e); }
});
function updateCampaignSlugPreview() {
  const slug = (document.getElementById('campaignEventSlug') as HTMLInputElement).value.trim() || 'your-slug';
  const preview = document.getElementById('campaignSlugPreview');
  if (preview) preview.textContent = slug;
}
function slugifyCampaignName(name: string): string { return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, ''); }
(window as any).campaignSlugAuto = false;
document.addEventListener('input', function (e: Event) {
  const el = e.target as HTMLElement;
  if (el && el.id === 'campaignEventSlug') { (window as any).campaignSlugAuto = false; updateCampaignSlugPreview(); }
  if (el && el.id === 'campaignEventDisplayName') {
    if ((window as any).campaignSlugAuto) {
      (document.getElementById('campaignEventSlug') as HTMLInputElement).value = slugifyCampaignName((el as HTMLInputElement).value);
      updateCampaignSlugPreview();
    }
  }
});
expose('saveCampaignEventSetting', async function () {
  const displayName = (document.getElementById('campaignEventDisplayName') as HTMLInputElement).value.trim();
  const slug = (document.getElementById('campaignEventSlug') as HTMLInputElement).value.trim();
  if (!displayName || !slug) { toast('Display name and slug are required', true); return; }
  const sourceTypeEl = document.querySelector('input[name="campaignSourceType"]:checked') as HTMLInputElement;
  const filterModeEl = document.querySelector('input[name="campaignFilterMode"]:checked') as HTMLInputElement;
  const authMethodEl = document.querySelector('input[name="campaignAuthMethod"]:checked') as HTMLInputElement;
  const body: any = {
    display_name: displayName, slug: slug,
    source_type: sourceTypeEl ? sourceTypeEl.value : 'google_api',
    ical_url: (document.getElementById('campaignEventIcalUrl') as HTMLInputElement).value.trim(),
    calendar_id: (document.getElementById('campaignEventCalendarId') as HTMLInputElement).value.trim(),
    tags: (document.getElementById('campaignEventTags') as HTMLInputElement).value.trim(),
    color_labels: (document.getElementById('campaignEventColorLabels') as HTMLInputElement).value.trim(),
    filter_mode: filterModeEl ? filterModeEl.value : 'text',
    auth_method: authMethodEl ? authMethodEl.value : 'service_account',
    credentials_json: (document.getElementById('campaignEventCredentialsJson') as HTMLInputElement).value.trim(),
    oauth_client_id: (document.getElementById('campaignEventOAuthClientId') as HTMLInputElement).value.trim(),
    oauth_client_secret: (document.getElementById('campaignEventOAuthClientSecret') as HTMLInputElement).value.trim(),
    oauth_refresh_token: (document.getElementById('campaignEventOAuthRefreshToken') as HTMLInputElement).value.trim(),
    cache_ttl_seconds: parseInt((document.getElementById('campaignEventCacheTTL') as HTMLInputElement).value) || 300,
    is_active: (document.getElementById('campaignEventIsActive') as HTMLInputElement).checked
  };
  try {
    if (campaignEventEditId) { body.id = campaignEventEditId; await api('PUT', '/api/admin/events-campaigns/' + campaignEventEditId, body); toast('Campaign page updated'); }
    else { await api('POST', '/api/admin/events-campaigns', body); toast('Campaign page created'); }
    const modalEl = document.getElementById('campaignEventModal')!;
    const modal = (window as any).bootstrap.Modal.getInstance(modalEl);
    if (modal) modal.hide();
    loadCampaignEventSettings();
  } catch (e: any) { renderError(e); }
});
expose('deleteCampaignEventSetting', async function (id: number) {
  if (!confirm('Delete this campaign event page? The public page will no longer be available.')) return;
  try { await api('DELETE', '/api/admin/events-campaigns/' + id); toast('Campaign page deleted'); loadCampaignEventSettings(); } catch (e: any) { renderError(e); }
});
expose('clearCampaignCache', async function (id: number) {
  try { await api('POST', '/api/admin/events-campaigns/' + id + '/clear-cache'); toast('Campaign cache cleared'); loadCampaignEventSettings(); } catch (e: any) { renderError(e); }
});
