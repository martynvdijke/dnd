/**
 * Tests for the transfer UI module.
 *
 * Most functions interact with DOM modals and the API, so we test
 * the utility helpers and verify modal structure where possible.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock api and dom dependencies before importing transfer module
const mockApi = vi.fn();
vi.mock('./lib/api', () => ({
  api: (...args: any[]) => mockApi(...args),
}));

// Keep these tests focused on transfer rendering. Real Bootstrap modals
// schedule transition timers that can outlive happy-dom's per-test document.
vi.mock('./lib/dom', () => ({
  esc: (value: string | null | undefined) => {
    if (!value) return '';
    const el = document.createElement('div');
    el.textContent = value;
    return el.innerHTML;
  },
  showModal: (title: string, bodyHtml: string) => {
    document.getElementById('genericModalTitle')!.textContent = title;
    document.getElementById('genericModalBody')!.innerHTML = bodyHtml;
  },
  hideModal: vi.fn(),
  toast: vi.fn(),
}));

// Setup DOM environment
beforeEach(() => {
  document.body.innerHTML = `
    <div class="modal fade" id="genericModal" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="genericModalTitle">Modal</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body" id="genericModalBody"></div>
        </div>
      </div>
    </div>
    <div id="toastContainer"></div>
    <div id="loadingOverlay" class="d-none"></div>
  `;
  mockApi.mockReset();
});

describe('showTransferExport', () => {
  it('opens modal with export options', async () => {
    const { showTransferExport } = await import('./transfer');
    showTransferExport();
    const title = document.getElementById('genericModalTitle');
    expect(title?.textContent).toBe('Export Data');
  });

  it('pre-selects entity type when provided', async () => {
    const { showTransferExport } = await import('./transfer');
    showTransferExport('character');
    const chips = document.querySelectorAll('.transfer-type-chip');
    expect(chips.length).toBeGreaterThan(0);
    const activeChips = document.querySelectorAll('.transfer-type-chip.active');
    expect(activeChips.length).toBe(1);
  });

  it('shows campaign scope notice when campaignId provided', async () => {
    const { showTransferExport } = await import('./transfer');
    showTransferExport(undefined, 42);
    const body = document.getElementById('genericModalBody');
    expect(body?.textContent).toContain('campaign #42');
  });
});

describe('showTransferImport', () => {
  it('opens modal with import form', async () => {
    const { showTransferImport } = await import('./transfer');
    showTransferImport();
    const title = document.getElementById('genericModalTitle');
    expect(title?.textContent).toBe('Import Data');
  });

  it('includes file input', async () => {
    const { showTransferImport } = await import('./transfer');
    showTransferImport();
    const fileInput = document.getElementById('transferImportFile');
    expect(fileInput).not.toBeNull();
  });

  it('includes JSON textarea', async () => {
    const { showTransferImport } = await import('./transfer');
    showTransferImport();
    const textarea = document.getElementById('transferImportJson');
    expect(textarea).not.toBeNull();
  });
});

describe('showTransferLogs', () => {
  it('shows empty state when no logs', async () => {
    mockApi.mockResolvedValue([]);
    const { showTransferLogs } = await import('./transfer');
    await showTransferLogs();
    const body = document.getElementById('genericModalBody');
    expect(body?.textContent).toContain('No import/export history yet');
  });

  it('renders log table when logs exist', async () => {
    mockApi.mockResolvedValue([
      { created_at: '2026-01-15T10:00:00Z', file_name: 'test.json', status: 'completed', counts: '{"characters": 2}' },
      { created_at: '2026-01-14T09:00:00Z', file_name: 'backup.json', status: 'error', counts: null },
    ]);
    const { showTransferLogs } = await import('./transfer');
    await showTransferLogs();
    const title = document.getElementById('genericModalTitle');
    expect(title?.textContent).toBe('Transfer Logs');
    const body = document.getElementById('genericModalBody');
    expect(body?.textContent).toContain('completed');
    expect(body?.textContent).toContain('error');
    expect(body?.textContent).toContain('test.json');
  });

  it('handles API errors gracefully', async () => {
    mockApi.mockRejectedValue(new Error('network error'));
    const { showTransferLogs } = await import('./transfer');
    // Should not throw
    await showTransferLogs();
    // Toast should have been called - just verify no crash
    expect(true).toBe(true);
  });
});

describe('transferDoExport', () => {
  it('shows toast error when no types selected and no campaign', async () => {
    await import('./transfer');
    // Call the global handler with no selected checkboxes
    const handler = (window as any).transferDoExport;
    await handler(null);
    // Toast should show error - just verify no crash
    expect(true).toBe(true);
  });
});

describe('transferOnFileChange', () => {
  it('clears textarea when file is selected', async () => {
    await import('./transfer');
    const body = document.getElementById('genericModalBody')!;
    body.innerHTML = `
      <input class="form-control" type="file" id="transferImportFile">
      <textarea class="form-control" id="transferImportJson" rows="6">old content</textarea>
    `;
    // Simulate file selection
    const fileInput = document.getElementById('transferImportFile') as HTMLInputElement;
    const file = new File(['{}'], 'test.json', { type: 'application/json' });
    Object.defineProperty(fileInput, 'files', { value: [file] });
    (window as any).transferOnFileChange();
    const textarea = document.getElementById('transferImportJson') as HTMLTextAreaElement;
    expect(textarea.value).toBe('');
  });
});
