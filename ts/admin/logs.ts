// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, attrEscape, toast } from '../lib/dom';
import { api, logRefreshInterval, setLogRefreshInterval } from './state';
import { renderError } from '../lib/errors';

function startLogAutoRefresh() {
  loadLogLevel();
  loadLogSources();
  loadLogs();
  if (logRefreshInterval) clearInterval(logRefreshInterval);
  setLogRefreshInterval(setInterval(() => {
    loadLogLevel();
    loadLogs();
  }, 5000));
}
function stopLogAutoRefresh() {
  if (logRefreshInterval) {
    clearInterval(logRefreshInterval);
    setLogRefreshInterval(null);
  }
}
async function loadLogSources() {
  try {
    const sources: string[] = await api('GET', '/api/admin/log-sources');
    const sel = document.getElementById('logSourceFilter') as HTMLSelectElement;
    if (!sel) return;
    const current = sel.value;
    sel.innerHTML = '<option value="">All</option>';
    const seen = new Set<string>();
    for (const s of sources) {
      if (!s || seen.has(s)) continue;
      seen.add(s);
      const opt = document.createElement('option');
      opt.value = s;
      opt.textContent = s.charAt(0).toUpperCase() + s.slice(1);
      sel.appendChild(opt);
    }
    if (current && [...sel.options].some(o => o.value === current)) {
      sel.value = current;
    }
  } catch { }
}
async function loadLogs() {
  try {
    const sourceFilter = (document.getElementById('logSourceFilter') as HTMLSelectElement)?.value || '';
    const searchTerm = ((document.getElementById('logSearch') as HTMLInputElement)?.value || '').toLowerCase();
    let url = '/api/admin/logs?limit=200';
    if (sourceFilter) url += '&source=' + encodeURIComponent(sourceFilter);
    const tableContainer = document.getElementById('adminLogs')?.querySelector('.table-responsive');
    const wasAtBottom = tableContainer
      ? tableContainer.scrollHeight - tableContainer.scrollTop - tableContainer.clientHeight < 50
      : false;
    const logs = await api('GET', url);
    (window as any).__lastLogs = logs || [];
    const tbody = document.getElementById('logBody')!;
    if (!logs || logs.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" class="text-muted text-center">No log entries</td></tr>';
      document.getElementById('logCount')!.textContent = '0 entries';
      return;
    }
    const levelBadge: Record<string, string> = {
      debug: 'bg-secondary',
      info: 'bg-info text-dark',
      warn: 'bg-warning text-dark',
      error: 'bg-danger',
    };
    const filtered = searchTerm ? logs.filter((l: any) => JSON.stringify(l).toLowerCase().includes(searchTerm)) : logs;
    tbody.innerHTML = filtered.map((l: any, idx: number) => {
      const badge = levelBadge[l.level] || 'bg-secondary';
      const ts = l.timestamp ? new Date(l.timestamp).toLocaleString() : '-';
      const hasAttrs = l.attributes && Object.keys(l.attributes).length > 0;
      const attrSummary = hasAttrs
        ? Object.entries(l.attributes).map(([k, v]) => `<span class="badge bg-light text-dark me-1">${esc(k)}=${esc(String(v)).substring(0, 30)}</span>`).join('')
        : '';
      const detailId = `logDetail_${idx}`;
      let attrsHtml = hasAttrs
        ? '<dl class="log-detail-attrs mb-0">' +
          Object.entries(l.attributes).map(([k, v]) => `<dt>${esc(k)}</dt><dd><code>${esc(JSON.stringify(v))}</code></dd>`).join('') +
          '</dl>'
        : '<span class="text-muted">No attributes</span>';
      const traceId = (l.attributes && (l.attributes as any)['trace_id']) || '';
      if (traceId) attrsHtml += `<button class="btn btn-outline-secondary btn-sm mt-2 js-copy-trace" data-trace-id="${attrEscape(traceId)}"><i class="fa-solid fa-copy me-1"></i>Copy trace id</button>`;
      return `<tr class="log-row js-toggle-log-detail" style="cursor:pointer" data-detail-id="${attrEscape(detailId)}">
        <td><span class="badge ${badge}">${esc(l.level)}</span></td>
        <td class="small">${esc(ts)}</td>
        <td><code class="small">${esc(l.source || '-')}</code></td>
        <td>${esc(l.message)}</td>
        <td class="small" style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${attrSummary}</td>
      </tr>
      <tr id="${detailId}" class="log-detail-row" style="display:none">
        <td colspan="5">${attrsHtml}</td>
      </tr>`;
    }).join('');
    const entryCount = logs.length;
    const totalCount = (logs[0] as any)._total !== undefined ? (logs[0] as any)._total : entryCount;
    document.getElementById('logCount')!.textContent = entryCount + ' entries' + (totalCount > entryCount ? ` (showing last ${entryCount})` : '');
    tbody.querySelectorAll<HTMLButtonElement>('.js-copy-trace').forEach(btn => {
      btn.addEventListener('click', (e) => { e.stopPropagation(); copyLogTrace(btn.dataset.traceId || ''); });
    });
    tbody.querySelectorAll<HTMLTableRowElement>('.js-toggle-log-detail').forEach(row => {
      row.addEventListener('click', () => toggleLogDetail(row.dataset.detailId || ''));
    });
    if (wasAtBottom && tableContainer) {
      tableContainer.scrollTop = tableContainer.scrollHeight;
    }
  } catch (e: any) {
    document.getElementById('logBody')!.innerHTML = '<tr><td colspan="5" class="text-danger text-center">Failed to load logs</td></tr>';
  }
}
function toggleLogDetail(detailId: string) {
  const detailRow = document.getElementById(detailId);
  if (!detailRow) return;
  const isVisible = detailRow.style.display !== 'none';
  detailRow.style.display = isVisible ? 'none' : 'table-row';
}
expose('loadLogs', loadLogs);
expose('toggleLogDetail', toggleLogDetail);
async function copyLogTrace(traceId: string) {
  try {
    await navigator.clipboard.writeText(traceId);
    toast('Trace id copied');
  } catch {
    const ta = document.createElement('textarea');
    ta.value = traceId;
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand('copy'); toast('Trace id copied'); } catch { toast('Copy failed', true); }
    document.body.removeChild(ta);
  }
}
expose('copyLogTrace', copyLogTrace);
function exportLogs() {
  const logs: any[] = (window as any).__lastLogs || [];
  if (!logs.length) { toast('No log entries to export', true); return; }
  const format = (document.getElementById('logExportFormat') as HTMLSelectElement)?.value || 'json';
  let blob: Blob;
  let name = 'logs.json';
  if (format === 'csv') {
    const cols = ['level', 'source', 'timestamp', 'message', 'attributes'];
    const rows = logs.map((l) => cols.map((c) => {
      const v = c === 'attributes' ? JSON.stringify(l.attributes || {}) : (l[c] ?? '');
      return '"' + String(v).replace(/"/g, '""') + '"';
    }).join(','));
    blob = new Blob([cols.join(',') + '\n' + rows.join('\n')], { type: 'text/csv' });
    name = 'logs.csv';
  } else {
    blob = new Blob([JSON.stringify(logs, null, 2)], { type: 'application/json' });
  }
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = name;
  a.click();
  URL.revokeObjectURL(a.href);
  toast('Logs exported (' + logs.length + ' entries)');
}
expose('exportLogs', exportLogs);
function toggleLogAutoRefresh() {
  const span = document.getElementById('logAutoRefresh');
  if (logRefreshInterval) {
    clearInterval(logRefreshInterval);
    setLogRefreshInterval(null);
    if (span) span.textContent = 'Auto-refresh: off';
  } else {
    setLogRefreshInterval(setInterval(() => { loadLogs(); }, 5000));
    if (span) span.textContent = 'Auto-refresh: on';
  }
}
expose('toggleLogAutoRefresh', toggleLogAutoRefresh);
document.addEventListener('click', (ev) => {
  const t = (ev.target as HTMLElement).closest('#logAutoRefresh');
  if (t) { ev.preventDefault(); toggleLogAutoRefresh(); }
});
async function loadLogLevel() {
  try {
    const res = await api('GET', '/api/admin/log-level');
    const sel = document.getElementById('logLevelSelect') as HTMLSelectElement;
    if (sel && res.level) sel.value = res.level;
  } catch {}
}
async function setLogLevel(level: string) {
  try {
    await api('PUT', '/api/admin/log-level', { level });
    loadLogs();
  } catch (e: any) {
    renderError(e);
  }
}
expose('setLogLevel', setLogLevel);
function clearLogFilters() {
  const sourceFilter = document.getElementById('logSourceFilter') as HTMLSelectElement;
  if (sourceFilter) sourceFilter.value = '';
  loadLogs();
}
expose('clearLogFilters', clearLogFilters);
expose('startLogAutoRefresh', startLogAutoRefresh);
expose('stopLogAutoRefresh', stopLogAutoRefresh);
