/**
 * Transfer UI — import/export via the villum-transfer envelope format.
 *
 * Integrates with the transfer backend (handlers/transfer.go):
 *   POST /api/transfer/export        — export selected entity types
 *   GET  /api/transfer/export/campaign/:id — campaign-scoped export
 *   POST /api/transfer/import        — import a transfer envelope
 *   GET  /api/transfer/logs          — list import history
 */
import { api } from './lib/api';
import { showModal, hideModal, esc, toast } from './lib/dom';
import { expose } from './lib/expose';

// ─── Types ───

interface TransferMeta {
  version: string;
  exported_by: string;
  exported_at: string;
  source_version: string;
}

interface TransferEntity {
  type: string;
  data: Record<string, any>;
}

interface TransferEntityResult {
  type: string;
  name: string;
  status: 'created' | 'updated' | 'skipped' | 'error';
  error?: string;
  new_id?: number;
}

interface TransferImportResult {
  success: boolean;
  total_entities: number;
  results: TransferEntityResult[];
}

// ─── Export ───

/**
 * Open a modal for exporting data.
 * @param preSelectedType  Optional entity type to pre-select (e.g. "character")
 * @param campaignId       Optional campaign ID to scope the export
 */
export function showTransferExport(preSelectedType?: string, campaignId?: number): void {
  const allTypes = [
    { value: 'character', label: 'Characters' },
    { value: 'npc', label: 'NPCs' },
    { value: 'campaign', label: 'Campaigns' },
    { value: 'location', label: 'Locations' },
    { value: 'shop', label: 'Shops' },
    { value: 'faction', label: 'Factions' },
  ];

  const typeChips = allTypes.map(t => `
    <label class="transfer-type-chip${preSelectedType === t.value ? ' active' : ''}">
      <input type="checkbox" value="${t.value}" ${preSelectedType === t.value ? 'checked' : ''}>
      <span>${esc(t.label)}</span>
    </label>
  `).join('');

  const campaignSection = campaignId
    ? `<p class="small text-muted mb-0">Scoped to campaign #${campaignId}</p>`
    : '';

  showModal('Export Data', `
    <p class="text-muted fst-italic small mb-3">Choose entity types to include in the export.</p>
    ${campaignSection}
    <div class="mb-3">
      <label class="form-label fw-medium">Entity Types</label>
      <div class="transfer-type-grid">${typeChips}</div>
    </div>
    <div class="d-flex gap-2">
      <button class="btn btn-primary flex-grow-1" onclick="transferDoExport(${campaignId ?? 'null'})">
        <i class="fa-solid fa-download me-1"></i>Export
      </button>
      <button class="btn btn-outline-secondary" onclick="showTransferLogs()">
        <i class="fa-solid fa-clock-rotate-left me-1"></i>Logs
      </button>
    </div>
  `);
}

/**
 * Perform export and download as JSON.
 * Called from onclick in the export modal.
 */
expose('transferDoExport', async function (campaignId: number | null) {
  const checked = document.querySelectorAll<HTMLInputElement>('.transfer-type-chip input:checked');
  const types = Array.from(checked).map(cb => cb.value);
  if (types.length === 0 && !campaignId) {
    toast('Select at least one entity type', true);
    return;
  }

  try {
    let data: { villum_transfer: TransferMeta; entities: TransferEntity[] };
    if (campaignId) {
      data = await api('GET', `/api/transfer/export/campaign/${campaignId}`);
    } else {
      data = await api('POST', '/api/transfer/export', { types });
    }

    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    const ts = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
    a.download = `villum-export-${ts}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    hideModal();
    toast(`Exported ${data.entities.length} entity(ies)`);
  } catch (e: any) {
    toast('Export failed: ' + e.message, true);
  }
});

// ─── Import ───

/**
 * Open a modal for importing a transfer file.
 */
export function showTransferImport(): void {
  showModal('Import Data', `
    <p class="text-muted fst-italic small mb-3">Upload a villum-transfer JSON file or paste the content.</p>
    <div class="mb-3">
      <label class="form-label">File</label>
      <input class="form-control" type="file" id="transferImportFile" accept=".json" onchange="transferOnFileChange()">
    </div>
    <div class="mb-3">
      <label class="form-label">Or paste JSON</label>
      <textarea class="form-control" id="transferImportJson" rows="6" style="font-family:monospace;font-size:0.8rem" placeholder='{"villum_transfer":{...},"entities":[...]}'></textarea>
    </div>
    <button class="btn btn-primary w-100" onclick="transferDoImport()">
      <i class="fa-solid fa-file-import me-1"></i>Preview Import
    </button>
  `);
}

/**
 * Read the import source (file or textarea) and return parsed JSON.
 */
async function readImportSource(): Promise<any> {
  const fileEl = document.getElementById('transferImportFile') as HTMLInputElement;
  const jsonEl = document.getElementById('transferImportJson') as HTMLTextAreaElement;

  if (fileEl.files && fileEl.files[0]) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        try {
          resolve(JSON.parse(reader.result as string));
        } catch {
          reject(new Error('Invalid JSON in file'));
        }
      };
      reader.onerror = () => reject(new Error('Failed to read file'));
      reader.readAsText(fileEl.files![0]);
    });
  }
  if (jsonEl.value.trim()) {
    return JSON.parse(jsonEl.value.trim());
  }
  throw new Error('Select a file or paste JSON');
}

expose('transferOnFileChange', function () {
  // Clear textarea when a file is selected
  const fileEl = document.getElementById('transferImportFile') as HTMLInputElement;
  if (fileEl.files && fileEl.files[0]) {
    (document.getElementById('transferImportJson') as HTMLTextAreaElement).value = '';
  }
});

/**
 * Run a dry-run import, then let the user confirm the actual import.
 */
expose('transferDoImport', async function () {
  let data: any;
  try {
    data = await readImportSource();
  } catch (e: any) {
    toast(e.message, true);
    return;
  }

  // Validate structure
  if (!data.villum_transfer || !Array.isArray(data.entities)) {
    toast('Not a valid villum-transfer file', true);
    return;
  }

  try {
    // Dry run first
    const dryResult: TransferImportResult = await api('POST', '/api/transfer/import?dry_run=true', data);
    const results = dryResult.results || [];

    const summaryHtml = `
      <div class="mb-3">
        <p class="mb-1">File: <strong>${esc(data.villum_transfer.source_version || data.villum_transfer.version || '?')}</strong></p>
        <p class="mb-1">Entities: <strong>${dryResult.total_entities ?? results.length}</strong></p>
      </div>
      <div class="transfer-preview-results">
        ${results.map(r => `
          <div class="transfer-result-row transfer-result-${r.status}">
            <span class="transfer-result-icon">
              ${r.status === 'created' ? '✓' : r.status === 'error' ? '✗' : '—'}
            </span>
            <span><strong>${esc(r.type)}</strong>: ${esc(r.name)}</span>
            <span class="transfer-result-status">${r.status}${r.new_id ? ` (id: ${r.new_id})` : ''}</span>
            ${r.error ? `<span class="text-danger small d-block">${esc(r.error)}</span>` : ''}
          </div>
        `).join('')}
      </div>
    `;

    const containerEl = document.getElementById('genericModalBody')!;
    containerEl.innerHTML = `
      <h5>Import Preview</h5>
      ${summaryHtml}
      <div class="d-flex gap-2 mt-3">
        <button class="btn btn-success flex-grow-1" onclick="transferConfirmImport()">
          <i class="fa-solid fa-check me-1"></i>Confirm Import
        </button>
        <button class="btn btn-outline-secondary" onclick="showTransferImport()">
          <i class="fa-solid fa-arrow-left me-1"></i>Back
        </button>
      </div>
    `;
    // Store data for confirmation step
    expose('__transferPendingData', data);
  } catch (e: any) {
    toast('Import preview failed: ' + e.message, true);
  }
});

/**
 * Execute the confirmed import after dry run preview.
 */
expose('transferConfirmImport', async function () {
  const data = (window as any).__transferPendingData;
  if (!data) {
    toast('No import data. Please start again.', true);
    return;
  }

  try {
    const result: TransferImportResult = await api('POST', '/api/transfer/import', data);
    expose('__transferPendingData', null);

    // Show result
    const results = result.results || [];
    const created = results.filter(r => r.status === 'created').length;
    const errors = results.filter(r => r.status === 'error').length;

    const resultHtml = `
      <div class="mb-3">
        <p class="mb-1">Result: ${result.success ? '<span class="text-success">Import complete</span>' : '<span class="text-warning">Import completed with issues</span>'}</p>
        <p class="mb-1">Created: <strong>${created}</strong> | Errors: <strong>${errors}</strong></p>
      </div>
      <div class="transfer-preview-results">
        ${results.map(r => `
          <div class="transfer-result-row transfer-result-${r.status}">
            <span class="transfer-result-icon">
              ${r.status === 'created' ? '✓' : r.status === 'error' ? '✗' : '—'}
            </span>
            <span><strong>${esc(r.type)}</strong>: ${esc(r.name)}</span>
            <span class="transfer-result-status">${r.status}</span>
            ${r.error ? `<span class="text-danger small d-block">${esc(r.error)}</span>` : ''}
          </div>
        `).join('')}
      </div>
    `;

    const containerEl = document.getElementById('genericModalBody')!;
    containerEl.innerHTML = `
      <h5>Import Results</h5>
      ${resultHtml}
      <div class="mt-3">
        <button class="btn btn-primary w-100" onclick="hideModal()">
          <i class="fa-solid fa-check me-1"></i>Done
        </button>
      </div>
    `;

    toast(`Import complete: ${created} created, ${errors} errors`);
    // Reload current view data
    if ((window as any).loadCharacters) (window as any).loadCharacters();
  } catch (e: any) {
    toast('Import failed: ' + e.message, true);
  }
});

// ─── Transfer Logs ───

/**
 * Open a modal showing import/export history.
 */
export async function showTransferLogs(): Promise<void> {
  try {
    const logs: any[] = await api('GET', '/api/transfer/logs');

    if (!logs || logs.length === 0) {
      showModal('Transfer Logs', `
        <p class="text-muted fst-italic">No import/export history yet.</p>
        <button class="btn btn-secondary w-100" onclick="hideModal()">Close</button>
      `);
      return;
    }

    const rowsHtml = logs.map((log: any) => {
      const ts = log.created_at ? new Date(log.created_at).toLocaleString() : '-';
      let countsInfo = '';
      if (log.counts) {
        try {
          const counts = typeof log.counts === 'string' ? JSON.parse(log.counts) : log.counts;
          countsInfo = Object.entries(counts).map(([k, v]) => `${k}: ${v}`).join(', ');
        } catch { countsInfo = String(log.counts); }
      }
      const badgeClass = log.status === 'completed' ? 'bg-success' : log.status === 'error' ? 'bg-danger' : 'bg-secondary';
      return `<tr>
        <td><span class="badge ${badgeClass}">${esc(log.status)}</span></td>
        <td class="small">${esc(ts)}</td>
        <td><code class="small">${esc(log.file_name || log.doc_type || '-')}</code></td>
        <td class="small text-muted">${esc(countsInfo)}</td>
      </tr>`;
    }).join('');

    showModal('Transfer Logs', `
      <div class="table-responsive" style="max-height:400px;overflow-y:auto">
        <table class="table table-sm table-hover mb-0">
          <thead><tr>
            <th>Status</th>
            <th>Date</th>
            <th>File</th>
            <th>Counts</th>
          </tr></thead>
          <tbody>${rowsHtml}</tbody>
        </table>
      </div>
      <div class="mt-3 d-flex gap-2">
        <button class="btn btn-outline-primary flex-grow-1" onclick="showTransferExport()">
          <i class="fa-solid fa-download me-1"></i>New Export
        </button>
        <button class="btn btn-outline-primary flex-grow-1" onclick="showTransferImport()">
          <i class="fa-solid fa-file-import me-1"></i>New Import
        </button>
        <button class="btn btn-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) {
    toast('Failed to load transfer logs: ' + e.message, true);
  }
}

// ─── Init — wire global references ───

expose('showTransferExport', showTransferExport);
expose('showTransferImport', showTransferImport);
expose('showTransferLogs', showTransferLogs);
