import { describe, it, expect, beforeEach, vi } from 'vitest';

const { exposed } = vi.hoisted(() => ({ exposed: {} as Record<string, any> }));

vi.mock('./lib/expose', () => ({
  expose: (name: string, value: any) => { exposed[name] = value; return value; },
}));
vi.mock('./lib/dom', () => ({
  esc: (s: any) => String(s ?? ''),
  showModal: (_title: string, body: string) => { document.body.innerHTML = body; },
  hideModal: vi.fn(),
  toast: vi.fn(),
}));
vi.mock('./lib/api', () => ({ api: vi.fn() }));
vi.mock('./lib/state', () => ({ currentCampaign: { id: 7 } }));

import { api } from './lib/api';
import './draft-import';

const mockApi = api as unknown as ReturnType<typeof vi.fn>;

const importReply = (overrides: Record<string, unknown> = {}) => ({
  id: 5,
  url: '#/adventures/5',
  entity_type: 'oneshot',
  name: 'The Wolves of Welton',
  updated: false,
  ...overrides,
});

beforeEach(() => {
  mockApi.mockReset();
  document.body.innerHTML = '';
});

describe('draft-import', () => {
  it('renders the import modal with the requested entity type', () => {
    exposed.openDraftImport('npc', { campaignId: 3 });
    const html = document.body.innerHTML;
    expect(html).toContain('draftImportJson');
    expect(html).toContain('draftImportInstruction');
    const select = document.getElementById('draftImportType') as HTMLSelectElement;
    expect(select.value).toBe('npc');
  });

  it('imports JSON with the campaign scope and then replaces the created entity', async () => {
    exposed.openDraftImport('oneshot');
    (document.getElementById('draftImportJson') as HTMLTextAreaElement).value = '{"title":"The Wolves of Welton"}';
    mockApi.mockResolvedValue(importReply());

    document.getElementById('draftImportApplyBtn')!.click();
    await vi.waitFor(() => expect(mockApi).toHaveBeenCalledTimes(1));
    expect(mockApi.mock.calls[0][0]).toBe('POST');
    expect(mockApi.mock.calls[0][1]).toBe('/api/ai/import');
    expect(mockApi.mock.calls[0][2]).toMatchObject({
      entity_type: 'oneshot',
      campaign_id: 7,
      json: '{"title":"The Wolves of Welton"}',
    });

    await vi.waitFor(() => {
      expect(document.getElementById('draftImportNotice')!.classList.contains('d-none')).toBe(false);
    });
    expect(document.getElementById('draftImportDoneBtn')!.classList.contains('d-none')).toBe(false);
    expect((document.getElementById('draftImportApplyBtn') as HTMLButtonElement).textContent).toBe('Apply changes');

    mockApi.mockResolvedValue(importReply({ updated: true }));
    document.getElementById('draftImportApplyBtn')!.click();
    await vi.waitFor(() => expect(mockApi).toHaveBeenCalledTimes(2));
    expect(mockApi.mock.calls[1][2]).toMatchObject({ entity_id: 5 });
  });

  it('asks the AI to revise the JSON and auto-applies when the entity exists', async () => {
    exposed.openDraftImport('oneshot', { json: '{"title":"Old"}' });
    (document.getElementById('draftImportInstruction') as HTMLInputElement).value = 'rename the title';

    mockApi.mockResolvedValueOnce({ status: 'ready', message: 'Renamed', draft: { title: 'New' } });
    document.getElementById('draftImportReviseBtn')!.click();
    await vi.waitFor(() => {
      expect((document.getElementById('draftImportJson') as HTMLTextAreaElement).value).toContain('"title": "New"');
    });
    expect(mockApi.mock.calls[0][1]).toBe('/api/ai/revise');
    expect(mockApi.mock.calls[0][2]).toMatchObject({ instruction: 'rename the title', json: '{"title":"Old"}' });

    // First import creates the entity.
    mockApi.mockResolvedValueOnce(importReply());
    document.getElementById('draftImportApplyBtn')!.click();
    await vi.waitFor(() => expect(mockApi).toHaveBeenCalledTimes(2));

    // The next revision is auto-applied to the existing entity.
    mockApi
      .mockResolvedValueOnce({ status: 'ready', message: 'Renamed again', draft: { title: 'Newer' } })
      .mockResolvedValueOnce(importReply({ updated: true }));
    document.getElementById('draftImportReviseBtn')!.click();
    await vi.waitFor(() => expect(mockApi).toHaveBeenCalledTimes(4));
    expect(mockApi.mock.calls[3][1]).toBe('/api/ai/import');
    expect(mockApi.mock.calls[3][2]).toMatchObject({ entity_id: 5, json: expect.stringContaining('"title": "Newer"') });
  });

  it('shows backend errors in the modal and keeps the JSON', async () => {
    exposed.openDraftImport('oneshot', { json: 'not json' });
    mockApi.mockRejectedValue(new Error('the pasted text is not a valid JSON object'));

    document.getElementById('draftImportApplyBtn')!.click();
    await vi.waitFor(() => {
      const box = document.getElementById('draftImportError')!;
      expect(box.classList.contains('d-none')).toBe(false);
      expect(box.textContent).toContain('not a valid JSON object');
    });
    expect((document.getElementById('draftImportJson') as HTMLTextAreaElement).value).toBe('not json');
  });
});
