import { describe, it, expect, vi, beforeEach } from 'vitest';

const mocks = vi.hoisted(() => ({
  api: vi.fn(),
  showModal: vi.fn(),
  toast: vi.fn(),
}));

vi.mock('./lib/api', () => ({ api: mocks.api }));
vi.mock('./lib/dom', () => ({
  esc: (s: string | null | undefined) => s ?? '',
  showModal: mocks.showModal,
  toast: mocks.toast,
}));
vi.mock('./lib/expose', () => ({ expose: () => {} }));

import { shareEntity, openSharedLinks, revokeShareLink } from './share';

describe('shareEntity', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('creates a link and shows the URL modal', async () => {
    mocks.api.mockResolvedValue({ url: 'https://x/api/share/abc', label: 'My Note' });
    await shareEntity('note', 42);
    expect(mocks.api).toHaveBeenCalledWith('POST', '/api/share', {
      entity_type: 'note',
      entity_id: 42,
    });
    expect(mocks.showModal).toHaveBeenCalledWith('Share Note', expect.stringContaining('https://x/api/share/abc'));
    expect(mocks.showModal).toHaveBeenCalledWith('Share Note', expect.stringContaining('My Note'));
  });

  it('falls back to the label from the response when no name is given', async () => {
    mocks.api.mockResolvedValue({ url: 'https://x/api/share/def', label: 'Session 12' });
    await shareEntity('journal', 7);
    expect(mocks.showModal).toHaveBeenCalledWith('Share Journal', expect.stringContaining('Session 12'));
  });

  it('shows a toast on failure', async () => {
    mocks.api.mockRejectedValue({ message: 'access denied' });
    await shareEntity('map', 1);
    expect(mocks.toast).toHaveBeenCalledWith('access denied', true);
    expect(mocks.showModal).not.toHaveBeenCalled();
  });
});

describe('openSharedLinks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders rows with labels and revoke buttons', async () => {
    mocks.api.mockResolvedValue([
      { token: 't1', url: 'https://x/api/share/t1', entity_type: 'note', entity_id: 1, label: 'Secret Plan' },
      { token: 't2', url: 'https://x/api/share/t2', entity_type: 'upload', entity_id: 2, label: 'map.png' },
    ]);
    await openSharedLinks();
    const html = mocks.showModal.mock.calls[0][1] as string;
    expect(html).toContain('Secret Plan');
    expect(html).toContain('map.png');
    expect(html).toContain("revokeShareLink('t1')");
    expect(html).toContain('share-link-row');
  });

  it('renders the empty state', async () => {
    mocks.api.mockResolvedValue([]);
    await openSharedLinks();
    const html = mocks.showModal.mock.calls[0][1] as string;
    expect(html).toContain('No shared links yet');
  });

  it('shows a toast on failure', async () => {
    mocks.api.mockRejectedValue({ message: 'boom' });
    await openSharedLinks();
    expect(mocks.toast).toHaveBeenCalledWith('boom', true);
  });
});

describe('revokeShareLink', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('deletes the link then refreshes the dialog', async () => {
    mocks.api.mockResolvedValueOnce({}).mockResolvedValueOnce([]);
    await revokeShareLink('t1');
    expect(mocks.api).toHaveBeenNthCalledWith(1, 'DELETE', '/api/share/t1');
    expect(mocks.api).toHaveBeenNthCalledWith(2, 'GET', '/api/share');
    expect(mocks.showModal).toHaveBeenCalledWith('Shared Links', expect.stringContaining('No shared links yet'));
  });
});
