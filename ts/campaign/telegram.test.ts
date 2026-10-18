import { describe, it, expect, beforeEach, vi } from 'vitest';
import { toast, hideModal } from '../lib/dom';
import { renderError } from '../lib/errors';

const mockApi = vi.fn();
vi.mock('../lib/api', () => ({
  api: (...args: any[]) => mockApi(...args),
}));

vi.mock('../lib/dom', () => ({
  esc: (s: string | null | undefined) => s || '',
  showModal: vi.fn((title: string, body: string) => {
    document.getElementById('genericModalTitle')!.textContent = title;
    document.getElementById('genericModalBody')!.innerHTML = body;
  }),
  hideModal: vi.fn(),
  toast: vi.fn(),
}));

vi.mock('../lib/errors', () => ({
  renderError: vi.fn(),
}));

vi.mock('../lib/state', () => ({
  currentCampaign: { id: 42, my_role: 'dm' },
  currentUser: { id: 1, role: 'dm' },
}));

const toastMock = vi.mocked(toast);
const hideModalMock = vi.mocked(hideModal);
const renderErrorMock = vi.mocked(renderError);

/** Chat id and the two delivery toggles every save reads. */
const BASE_FORM = `
  <input id="telegramCampaignChatId" value="-100999">
  <input type="checkbox" id="telegramCampaignEnabled" checked>
  <input type="checkbox" id="telegramCampaignAutoPost">
`;

/** The cadence input, as the rendered form always has one. */
const CADENCE_FIELD = '<input id="tgAutopostCadence" value="0">';

/**
 * Read which hour options carry the `selected` attribute. happy-dom reports a
 * wrong <select> value when a non-first option is marked selected, so the
 * markup is the reliable record of what the browser will pre-select.
 */
function selectedHourValues(html: string): string[] {
  return Array.from(html.matchAll(/<option value="(\d+)"\s+selected(?:="")?>/g)).map((m) => m[1]);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockApi.mockReset();
  document.body.innerHTML = `
    <div id="genericModal"><div id="genericModalTitle"></div><div id="genericModalBody"></div></div>
    <div id="toastContainer"></div>
  `;
});

describe('campaign telegram fetch wrappers', () => {
  it('getCampaignTelegram calls GET with campaign id', async () => {
    mockApi.mockResolvedValue({ chat_id: -100123, chat_type: 'group', title_cache: '', is_enabled: true, auto_post_enabled: false, bound_at: null });
    const m = await import('./telegram');
    const res = await m.getCampaignTelegram(42);
    expect(mockApi).toHaveBeenCalledWith('GET', '/api/campaigns/42/telegram');
    expect(res.chat_id).toBe(-100123);
  });

  it('saveCampaignTelegram calls PUT', async () => {
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegram(42, { chat_id: -100123, is_enabled: true, auto_post_enabled: true });
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/42/telegram', { chat_id: -100123, is_enabled: true, auto_post_enabled: true });
  });

  it('showCampaignTelegram renders modal with inputs and data-testids', async () => {
    mockApi.mockResolvedValue({ chat_id: -100123, chat_type: 'group', title_cache: 'My Group', is_enabled: true, auto_post_enabled: false, bound_at: null });
    const m = await import('./telegram');
    await m.showCampaignTelegram(42);
    const body = document.getElementById('genericModalBody')!.innerHTML;
    expect(body).toContain('data-testid="telegram-campaign-chat-id"');
    expect(body).toContain('data-testid="telegram-enable-toggle"');
    expect(body).toContain('data-testid="telegram-auto-post-toggle"');
    expect(body).toContain('data-testid="telegram-auto-save-media-toggle"');
    expect(body).toContain('data-testid="telegram-save-campaign"');
    expect(body).toContain('-100123');
  });
});

describe('campaign telegram scheduling controls', () => {
  it('populates the cadence input and the stored quiet window', async () => {
    mockApi.mockResolvedValue({
      chat_id: -100123,
      chat_type: 'group',
      title_cache: '',
      is_enabled: true,
      auto_post_enabled: true,
      bound_at: null,
      auto_post_cadence_minutes: 60,
      quiet_start_hour: 22,
      quiet_end_hour: 7,
      auto_post_last_at: '2026-10-17T21:04:00Z',
    });
    const m = await import('./telegram');
    await m.showCampaignTelegram(42);

    const body = document.getElementById('genericModalBody')!;
    expect(body.innerHTML).toContain('data-testid="tg-autopost-cadence"');
    expect(body.innerHTML).toContain('data-testid="tg-quiet-hours-toggle"');
    expect(body.innerHTML).toContain('data-testid="tg-quiet-start"');
    expect(body.innerHTML).toContain('data-testid="tg-quiet-end"');
    // The helper text has to say what 0 does — a placeholder is not enough.
    expect(body.textContent).toContain('0 posts as soon as there is something to post');
    expect(body.textContent).toContain('Last automatic post: 2026-10-17 21:04 UTC');

    expect((document.getElementById('tgAutopostCadence') as HTMLInputElement).value).toBe('60');
    expect((document.getElementById('tgQuietHoursToggle') as HTMLInputElement).checked).toBe(true);
    // Start and end select are pre-selected to the stored window.
    expect(selectedHourValues(body.innerHTML)).toEqual(['22', '7']);
    // A stored window means the fields are on screen, not hidden.
    expect(document.getElementById('tgQuietHoursFields')!.classList.contains('d-none')).toBe(false);
  });

  it('labels the quiet hours as UTC and lists every hour', async () => {
    mockApi.mockResolvedValue({ chat_id: null, chat_type: '', title_cache: '', is_enabled: false, auto_post_enabled: false, bound_at: null });
    const m = await import('./telegram');
    await m.showCampaignTelegram(42);

    const body = document.getElementById('genericModalBody')!;
    expect(body.textContent).toContain('Quiet hours (UTC)');
    expect(body.textContent).toContain('Quiet from (UTC)');
    expect(body.textContent).toContain('Quiet until (UTC)');
    const start = document.getElementById('tgQuietStart') as HTMLSelectElement;
    const end = document.getElementById('tgQuietEnd') as HTMLSelectElement;
    expect(start.options.length).toBe(24);
    expect(end.options.length).toBe(24);
    expect(start.options[0].textContent).toBe('00:00');
    expect(start.options[23].textContent).toBe('23:00');
    // Both hours are always filled in, so switching quiet hours on needs no typing.
    expect(selectedHourValues(body.innerHTML)).toEqual(['22', '7']);
  });

  it('hides the quiet-hour fields while the toggle is off', async () => {
    mockApi.mockResolvedValue({ chat_id: null, chat_type: '', title_cache: '', is_enabled: false, auto_post_enabled: false, bound_at: null });
    const m = await import('./telegram');
    await m.showCampaignTelegram(42);

    const fields = document.getElementById('tgQuietHoursFields')!;
    expect((document.getElementById('tgQuietHoursToggle') as HTMLInputElement).checked).toBe(false);
    expect(fields.classList.contains('d-none')).toBe(true);

    m.toggleTgQuietHours(true);
    expect(fields.classList.contains('d-none')).toBe(false);

    m.toggleTgQuietHours(false);
    expect(fields.classList.contains('d-none')).toBe(true);
  });

  it('does nothing when the quiet-hour fields are not on screen', async () => {
    const m = await import('./telegram');
    expect(() => m.toggleTgQuietHours(true)).not.toThrow();
  });

  it('says when no automatic post has been sent yet', async () => {
    mockApi.mockResolvedValue({ chat_id: null, chat_type: '', title_cache: '', is_enabled: false, auto_post_enabled: false, bound_at: null, auto_post_last_at: null });
    const m = await import('./telegram');
    await m.showCampaignTelegram(42);
    expect(document.getElementById('genericModalBody')!.textContent).toContain('No automatic post has been sent yet.');
  });
});

describe('campaign telegram save', () => {
  it('validates chat_id and calls API', async () => {
    document.body.innerHTML += `${BASE_FORM}${CADENCE_FIELD}`;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/42/telegram', expect.objectContaining({ chat_id: -100999, is_enabled: true, auto_post_cadence_minutes: 0 }));
  });

  it('allows empty chat_id as null', async () => {
    document.body.innerHTML += `
      <input id="telegramCampaignChatId" value="">
      <input type="checkbox" id="telegramCampaignEnabled">
      <input type="checkbox" id="telegramCampaignAutoPost">
      ${CADENCE_FIELD}
    `;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    document.getElementById('telegramCampaignChatId')!.setAttribute('value', '');
    (document.getElementById('telegramCampaignChatId') as HTMLInputElement).value = '';
    await m.saveCampaignTelegramForm(99);
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/99/telegram', expect.objectContaining({ chat_id: null }));
  });

  it('sends a cadence and a quiet window', async () => {
    document.body.innerHTML += `
      ${BASE_FORM}
      <input id="tgAutopostCadence" value="120">
      <input type="checkbox" id="telegramCampaignAutoSaveMedia" checked>
      <input type="checkbox" id="tgQuietHoursToggle" checked>
      <select id="tgQuietStart"><option value="22" selected>22:00</option></select>
      <select id="tgQuietEnd"><option value="7" selected>07:00</option></select>
    `;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/42/telegram', {
      chat_id: -100999,
      is_enabled: true,
      auto_post_enabled: false,
      auto_save_media: true,
      auto_post_cadence_minutes: 120,
      quiet_start_hour: 22,
      quiet_end_hour: 7,
    });
    expect(hideModalMock).toHaveBeenCalled();
  });

  it('sends null for both quiet hours when the toggle is off', async () => {
    document.body.innerHTML += `
      ${BASE_FORM}
      <input id="tgAutopostCadence" value="45">
      <input type="checkbox" id="telegramCampaignAutoSaveMedia" checked>
      <input type="checkbox" id="tgQuietHoursToggle">
      <select id="tgQuietStart"><option value="22" selected>22:00</option></select>
      <select id="tgQuietEnd"><option value="7" selected>07:00</option></select>
    `;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).toHaveBeenCalledWith('PUT', '/api/campaigns/42/telegram', expect.objectContaining({
      auto_post_cadence_minutes: 45,
      quiet_start_hour: null,
      quiet_end_hour: null,
    }));
  });

  it('refuses a cadence that is not a whole number of minutes in range', async () => {
    document.body.innerHTML += `${BASE_FORM}<input id="tgAutopostCadence" value="2000">`;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).not.toHaveBeenCalled();
    expect(toastMock).toHaveBeenCalledWith('Cadence must be a whole number of minutes from 0 to 1440', true);

    (document.getElementById('tgAutopostCadence') as HTMLInputElement).value = '60.5';
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).not.toHaveBeenCalled();
    expect(toastMock).toHaveBeenCalledTimes(2);
  });

  it('refuses a cadence left blank', async () => {
    document.body.innerHTML += `${BASE_FORM}<input id="tgAutopostCadence" value="">`;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).not.toHaveBeenCalled();
    expect(toastMock).toHaveBeenCalledWith('Cadence must be a whole number of minutes from 0 to 1440', true);
  });

  it('refuses a quiet window with only one hour set', async () => {
    document.body.innerHTML += `
      ${BASE_FORM}
      ${CADENCE_FIELD}
      <input type="checkbox" id="tgQuietHoursToggle" checked>
      <select id="tgQuietStart"><option value="22" selected>22:00</option></select>
    `;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).not.toHaveBeenCalled();
    expect(toastMock).toHaveBeenCalledWith('Quiet hours need a start hour and an end hour from 0 to 23', true);
  });

  it('refuses a quiet hour outside 0 to 23', async () => {
    document.body.innerHTML += `
      ${BASE_FORM}
      ${CADENCE_FIELD}
      <input type="checkbox" id="tgQuietHoursToggle" checked>
      <select id="tgQuietStart"><option value="24" selected>24:00</option></select>
      <select id="tgQuietEnd"><option value="7" selected>07:00</option></select>
    `;
    mockApi.mockResolvedValue({ ok: true });
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(mockApi).not.toHaveBeenCalled();
    expect(toastMock).toHaveBeenCalledWith('Quiet hours need a start hour and an end hour from 0 to 23', true);
  });

  it('turns a server 400 into a readable message', async () => {
    document.body.innerHTML += `${BASE_FORM}${CADENCE_FIELD}`;
    mockApi.mockRejectedValue(Object.assign(new Error('auto_post_cadence_minutes out of range'), { status: 400 }));
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(toastMock).toHaveBeenCalledWith('Cadence must be a whole number of minutes from 0 to 1440', true);
    expect(renderErrorMock).not.toHaveBeenCalled();
    expect(hideModalMock).not.toHaveBeenCalled();
  });

  it('leaves other failures to the shared error renderer', async () => {
    document.body.innerHTML += `${BASE_FORM}${CADENCE_FIELD}`;
    const err = Object.assign(new Error('forbidden'), { status: 403 });
    mockApi.mockRejectedValue(err);
    const m = await import('./telegram');
    await m.saveCampaignTelegramForm(42);
    expect(renderErrorMock).toHaveBeenCalledWith(err);
  });
});
