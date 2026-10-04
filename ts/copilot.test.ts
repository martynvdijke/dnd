import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

const { exposed } = vi.hoisted(() => ({ exposed: {} as Record<string, any> }));

vi.mock('./lib/expose', () => ({
  expose: (name: string, value: any) => { exposed[name] = value; return value; },
}));
vi.mock('./lib/dom', () => ({
  esc: (s: any) => {
    if (!s) return '';
    const map: Record<string, string> = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };
    return String(s).replace(/[&<>"']/g, c => map[c]);
  },
  toast: vi.fn(),
}));
vi.mock('./lib/api', () => ({ api: vi.fn() }));
vi.mock('./navigation', () => ({ showView: vi.fn() }));
vi.mock('./lib/state', () => ({ currentCampaign: { id: 42 }, setCurrentCampaign: vi.fn() }));

import { api } from './lib/api';
import { toast } from './lib/dom';
import {
  buildExtractErrorMessage,
  recapPreviewHtml,
  extractedDraftRowHtml,
  getTranscriptId,
} from './copilot';
import './copilot';

const mockApi = api as unknown as ReturnType<typeof vi.fn>;
const mockToast = toast as unknown as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockApi.mockReset();
  (mockToast as any).mockReset?.();
  document.body.innerHTML = '<div id="copilotContent"></div><div id="toastContainer"></div><div id="loadingOverlay" class="d-none"></div>';
  (window as any).__copilotTranscriptId = null;
  exposed._copilotTestReset?.();
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) } as any));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('copilot pure helpers', () => {
  it('buildExtractErrorMessage combines error and message', () => {
    expect(buildExtractErrorMessage({ error: 'AI disabled', message: 'enable it' })).toBe('AI disabled: enable it');
    expect(buildExtractErrorMessage({ error: 'Bad' })).toBe('Bad');
    expect(buildExtractErrorMessage({ message: 'only msg' })).toBe('only msg');
    expect(buildExtractErrorMessage({})).toBe('Extraction failed');
    expect(buildExtractErrorMessage(null)).toBe('Extraction failed');
  });

  it('recapPreviewHtml escapes and preserves line breaks', () => {
    const html = recapPreviewHtml({ title: '<b>Title</b>', content: 'line1\nline2<script>' });
    expect(html).toContain('&lt;b&gt;Title&lt;/b&gt;');
    expect(html).toContain('line1<br>line2');
    expect(html).not.toContain('<script>');
  });

  it('extractedDraftRowHtml escapes and marks committed', () => {
    const html = extractedDraftRowHtml({ id: '1', entity_type: 'npc', name: '<img>', status: 'ready' });
    expect(html).toContain('data-testid="copilot-extracted-draft"');
    expect(html).toContain('npc');
    expect(html).toContain('&lt;img&gt;');
    expect(html).toContain('data-testid="copilot-draft-commit"');
    expect(html).toContain('data-testid="copilot-draft-discard"');
    const committed = extractedDraftRowHtml({ id: '2', entity_type: 'location', name: 'Inn', status: 'committed' });
    expect(committed).toContain('committed');
    expect(committed).toContain('disabled');
  });

  it('getTranscriptId reads window fallback', () => {
    expect(getTranscriptId()).toBe(null);
    exposed._copilotTestSetTranscript(99);
    expect(getTranscriptId()).toBe(99);
    exposed._copilotTestReset();
    (window as any).__copilotTranscriptId = 77;
    expect(getTranscriptId()).toBe(77);
  });
});

describe('copilot panel integration', () => {
  it('renders Recap + extract button disabled without transcript and enabled with one', async () => {
    exposed._copilotTestReset();
    await exposed.refreshCopilotPanel();
    let btn = document.getElementById('copilotExtractBtn') as HTMLButtonElement;
    expect(btn).toBeTruthy();
    expect(btn.getAttribute('data-testid')).toBe('copilot-generate-recap-extract');
    expect(btn.disabled).toBe(true);
    exposed._copilotTestSetTranscript(123);
    await exposed.refreshCopilotPanel();
    btn = document.getElementById('copilotExtractBtn') as HTMLButtonElement;
    expect(btn.disabled).toBe(false);
  });

  it('generates recap + drafts via extract endpoint and renders them escaped', async () => {
    exposed._copilotTestSetTranscript(55);
    await exposed.refreshCopilotPanel();
    mockApi.mockResolvedValueOnce({
      recap: { title: '<b>My Recap</b>', content: 'Hello\nWorld' },
      drafts: [{ id: 'd1', entity_type: 'npc', name: '<Evil>' }, { id: 'd2', entity_type: 'location', name: 'Tavern' }],
      source: { entity_type: 'wiki', entity_id: 1, title: 't' },
    });
    await exposed.copilotGenerateRecapExtract();
    expect(mockApi).toHaveBeenCalledWith('POST', expect.stringContaining('/copilot/transcript/55/extract'), {});
    const recapEl = document.getElementById('copilotRecapPreview')!;
    expect(recapEl.getAttribute('data-testid')).toBe('copilot-recap-preview');
    expect(recapEl.classList.contains('d-none')).toBe(false);
    expect(recapEl.innerHTML).toContain('&lt;b&gt;My Recap&lt;/b&gt;');
    expect(recapEl.innerHTML).toContain('Hello<br>World');
    const draftsWrap = document.getElementById('copilotExtractedDrafts')!;
    expect(draftsWrap.getAttribute('data-testid')).toBe('copilot-extracted-drafts');
    const rows = draftsWrap.querySelectorAll('[data-testid="copilot-extracted-draft"]');
    expect(rows.length).toBe(2);
    expect(draftsWrap.innerHTML).toContain('&lt;Evil&gt;');
    expect(draftsWrap.querySelector('[data-testid="copilot-draft-commit"]')).toBeTruthy();
    expect(draftsWrap.querySelector('[data-testid="copilot-draft-discard"]')).toBeTruthy();
  });

  it('shows inline error when extract fails (503/422) and surfaces message when present', async () => {
    exposed._copilotTestSetTranscript(56);
    await exposed.refreshCopilotPanel();
    const err: any = Object.assign(new Error('AI disabled: enable AI'), {
      status: 422,
      payload: { error: 'AI disabled', message: 'enable AI' },
    });
    mockApi.mockRejectedValueOnce(err);
    await exposed.copilotGenerateRecapExtract();
    const errEl = document.getElementById('copilotExtractError')!;
    expect(errEl.textContent).toContain('AI disabled');
    expect(errEl.textContent).toContain('enable AI');
    expect(errEl.classList.contains('d-none')).toBe(false);
    // no second network call — error payload supplies the message
    expect(globalThis.fetch).not.toHaveBeenCalled();
  });

  it('falls back to e.message when extract error has no payload', async () => {
    exposed._copilotTestSetTranscript(56);
    await exposed.refreshCopilotPanel();
    mockApi.mockRejectedValueOnce(new Error('network down'));
    await exposed.copilotGenerateRecapExtract();
    const errEl = document.getElementById('copilotExtractError')!;
    expect(errEl.textContent).toContain('network down');
    expect(errEl.classList.contains('d-none')).toBe(false);
    expect(globalThis.fetch).not.toHaveBeenCalled();
  });

  it('commits and discards drafts via draft endpoints and updates row badge', async () => {
    exposed._copilotTestSetTranscript(57);
    exposed._copilotTestSetExtract({ title: 'T', content: 'C' }, [
      { id: 'd10', entity_type: 'npc', name: 'Bob', status: 'ready' },
      { id: 'd11', entity_type: 'faction', name: 'Red', status: 'ready' },
    ], '');
    await exposed.refreshCopilotPanel();
    mockApi.mockResolvedValueOnce({});
    await exposed.copilotCommitExtractedDraft('d10');
    expect(mockApi).toHaveBeenCalledWith('POST', '/api/ai/draft/d10/commit', {});
    let rows = document.getElementById('copilotExtractedDrafts')!.innerHTML;
    expect(rows).toContain('committed');

    mockApi.mockResolvedValueOnce({});
    await exposed.copilotDiscardExtractedDraft('d11');
    expect(mockApi).toHaveBeenCalledWith('DELETE', '/api/ai/draft/d11', undefined);
    rows = document.getElementById('copilotExtractedDrafts')!.innerHTML;
    expect(rows).toContain('discarded');
  });
});
