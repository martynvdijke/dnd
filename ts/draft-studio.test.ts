import { describe, it, expect, beforeEach, vi } from 'vitest';

const { exposed } = vi.hoisted(() => ({ exposed: {} as Record<string, any> }));

vi.mock('./lib/expose', () => ({
  expose: (name: string, value: any) => { exposed[name] = value; return value; },
}));
vi.mock('./lib/dom', () => ({
  esc: (s: any) => String(s ?? ''),
  toast: vi.fn(),
}));
vi.mock('./lib/api', () => ({ api: vi.fn() }));
vi.mock('./lib/state', () => ({ currentCampaign: { id: 7 } }));

import { api } from './lib/api';
import './draft-studio';

const mockApi = api as unknown as ReturnType<typeof vi.fn>;

const MODAL_HTML = `
  <div id="aiDraftModal">
    <h5 id="aiDraftModalTitle"></h5>
    <input id="aiDraftEntityType" value="oneshot" />
    <select id="aiDraftEndpoint"></select>
    <div id="aiDraftMessages"></div>
    <textarea id="aiDraftInput"></textarea>
    <button id="aiDraftSendBtn"></button>
    <span id="aiDraftStatus"></span>
    <div id="aiDraftSections"></div>
    <textarea id="aiDraftJson"></textarea>
    <input id="aiDraftTweakInput" />
    <button id="aiDraftTweakBtn"></button>
    <div id="aiDraftError"></div>
    <div id="aiDraftNotice"></div>
    <button id="aiDraftImportBtn"></button>
    <button id="aiDraftCopyBtn"></button>
    <button id="aiDraftDoneBtn"></button>
  </div>`;

function setupDom() {
  document.body.innerHTML = MODAL_HTML;
  (window as any).bootstrap = {
    Modal: { getOrCreateInstance: () => ({ show: vi.fn(), hide: vi.fn() }) },
  };
  (window as any).showOneShots = vi.fn();
}

function setJson(value: unknown) {
  const ta = document.getElementById('aiDraftJson') as HTMLTextAreaElement;
  ta.value = typeof value === 'string' ? value : JSON.stringify(value);
  ta.dispatchEvent(new Event('input', { bubbles: true }));
}

const DRAFT = { title: 'The Wolves of Welton', acts: [{ title: 'Arrival', scenes: [] }], npcs: [{ name: 'Marla' }] };

beforeEach(() => {
  mockApi.mockReset();
  setupDom();
  localStorage.clear();
});

describe('draft-studio', () => {
  it('imports the JSON in the editor through /api/ai/import', async () => {
    mockApi.mockImplementation((_method: string, path: string) =>
      path === '/api/ai/import'
        ? Promise.resolve({ id: 42, url: '/adventure/42', entity_type: 'oneshot', name: 'The Wolves of Welton', updated: false })
        : Promise.resolve([])
    );
    exposed.openDraftImport('oneshot', { campaignId: 7 });
    setJson(DRAFT);

    await exposed.commitAIDraft();

    const call = mockApi.mock.calls.find((c) => c[1] === '/api/ai/import');
    expect(call).toBeTruthy();
    const [method, path, body] = call!;
    expect(method).toBe('POST');
    expect(path).toBe('/api/ai/import');
    expect(body.entity_type).toBe('oneshot');
    expect(body.campaign_id).toBe(7);
    expect(body.entity_id).toBeUndefined();
    expect(JSON.parse(body.json).title).toBe('The Wolves of Welton');
    const notice = document.getElementById('aiDraftNotice')!;
    expect(notice.classList.contains('d-none')).toBe(false);
    expect(notice.textContent).toContain('The Wolves of Welton');
  });

  it('tweaks only the selected section via /api/ai/revise', async () => {
    exposed.openDraftImport('oneshot', { campaignId: 7 });
    setJson(DRAFT);
    const chip = document.querySelector('[data-section="Act 1: Arrival"]') as HTMLElement;
    expect(chip).toBeTruthy();
    exposed.selectDraftSection(chip);

    mockApi.mockResolvedValueOnce({ status: 'ready', message: 'done', draft: { ...DRAFT, title: 'Scarier Wolves' } });
    (document.getElementById('aiDraftTweakInput') as HTMLInputElement).value = 'make it scarier';

    await exposed.tweakSelectedDraftSection();

    const call = mockApi.mock.calls.find((c) => c[1] === '/api/ai/revise');
    expect(call).toBeTruthy();
    const [method, path, body] = call!;
    expect(method).toBe('POST');
    expect(path).toBe('/api/ai/revise');
    expect(body.section).toBe('Act 1: Arrival');
    expect(body.instruction).toBe('make it scarier');
    expect((document.getElementById('aiDraftJson') as HTMLTextAreaElement).value).toContain('Scarier Wolves');
  });

  it('fills the editor when a chat reply is ready', async () => {
    mockApi.mockResolvedValueOnce([]);
    exposed.openAIDraft('oneshot', { campaignId: 7 });
    (document.getElementById('aiDraftInput') as HTMLTextAreaElement).value = 'draft me wolves';
    mockApi.mockResolvedValueOnce({
      id: 'sess-1', entity_type: 'oneshot', status: 'ready', message: 'Here you go',
      draft: DRAFT, messages: [{ role: 'user', content: 'draft me wolves' }, { role: 'assistant', content: 'Here you go' }],
    });

    await exposed.sendAIDraftMessage();

    const [method, path] = mockApi.mock.calls[1];
    expect(method).toBe('POST');
    expect(path).toBe('/api/ai/draft');
    expect((document.getElementById('aiDraftJson') as HTMLTextAreaElement).value).toContain('The Wolves of Welton');
    expect(document.getElementById('aiDraftImportBtn')!.textContent).toContain('Import');
  });

  it('shows import failures in the modal error area', async () => {
    mockApi.mockRejectedValueOnce(new Error('the pasted text is not a valid JSON object'));
    exposed.openDraftImport('oneshot', { campaignId: 7 });
    setJson('{"title": "broken"');

    await exposed.commitAIDraft();

    const err = document.getElementById('aiDraftError')!;
    expect(err.classList.contains('d-none')).toBe(false);
    expect(err.textContent).toContain('The draft JSON is not valid.');
  });
});
