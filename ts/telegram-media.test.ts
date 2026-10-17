import { describe, it, expect, beforeEach, vi } from 'vitest';

const mockApi = vi.fn();
vi.mock('./lib/api', () => ({ api: (...args: any[]) => mockApi(...args) }));
vi.mock('./lib/dom', () => ({
  esc: (s: string | null | undefined) => (s ? String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;') : ''),
  toast: vi.fn(),
}));
vi.mock('./lib/expose', () => ({ expose: vi.fn() }));

import { validateSelection, batchAlbumIds, isAlbumKind, sizeLimitSimple, renderTargetOptions, renderResults, fetchTargets } from './telegram-media';

beforeEach(() => {
  mockApi.mockReset();
  document.body.innerHTML = '';
});

describe('telegram-media helpers', () => {
  it('isAlbumKind true for image and video', () => {
    expect(isAlbumKind('image')).toBe(true);
    expect(isAlbumKind('video')).toBe(true);
    expect(isAlbumKind('audio')).toBe(false);
    expect(isAlbumKind('pdf')).toBe(false);
    expect(isAlbumKind('')).toBe(false);
  });

  it('sizeLimitSimple returns 10MB for image else 50MB', () => {
    expect(sizeLimitSimple('image')).toBe(10 * 1024 * 1024);
    expect(sizeLimitSimple('audio')).toBe(50 * 1024 * 1024);
    expect(sizeLimitSimple('video')).toBe(50 * 1024 * 1024);
    expect(sizeLimitSimple('pdf')).toBe(50 * 1024 * 1024);
  });

  it('validateSelection rejects empty and >10', () => {
    expect(validateSelection([])).toMatch(/Select at least/);
    const many = Array.from({ length: 11 }, (_, i) => ({ id: i, kind: 'image' }));
    expect(validateSelection(many)).toMatch(/up to 10/);
  });

  it('validateSelection rejects oversized photo', () => {
    expect(validateSelection([{ id: 1, kind: 'image', size: 11 * 1024 * 1024 }])).toMatch(/exceeds 10 MB/);
  });

  it('validateSelection rejects oversized video at 51MB', () => {
    expect(validateSelection([{ id: 2, kind: 'video', size: 51 * 1024 * 1024 }])).toMatch(/exceeds 50 MB/);
  });

  it('validateSelection passes within limits', () => {
    expect(validateSelection([{ id: 1, kind: 'image', size: 5 * 1024 * 1024 }])).toBeNull();
    expect(validateSelection([{ id: 1, kind: 'audio', size: 20 * 1024 * 1024 }])).toBeNull();
  });

  it('batchAlbumIds groups consecutive photo/video up to 10', () => {
    const items = [
      { id: 1, kind: 'image' },
      { id: 2, kind: 'video' },
      { id: 3, kind: 'image' },
      { id: 4, kind: 'audio' },
      { id: 5, kind: 'image' },
      { id: 6, kind: 'video' },
    ];
    const batches = batchAlbumIds(items);
    expect(batches.length).toBe(3);
    expect(batches[0].map(i => i.id)).toEqual([1, 2, 3]);
    expect(batches[1].map(i => i.id)).toEqual([4]);
    expect(batches[2].map(i => i.id)).toEqual([5, 6]);
  });

  it('batchAlbumIds splits at 10', () => {
    const items = Array.from({ length: 12 }, (_, i) => ({ id: i + 1, kind: 'image' }));
    const batches = batchAlbumIds(items);
    expect(batches.length).toBe(2);
    expect(batches[0].length).toBe(10);
    expect(batches[1].length).toBe(2);
  });

  it('batchAlbumIds handles mixed non-album interruptions', () => {
    const items = [
      { id: 1, kind: 'image' },
      { id: 2, kind: 'pdf' },
      { id: 3, kind: 'image' },
    ];
    const batches = batchAlbumIds(items);
    expect(batches).toEqual([[{ id: 1, kind: 'image' }], [{ id: 2, kind: 'pdf' }], [{ id: 3, kind: 'image' }]]);
  });

  it('fetchTargets calls GET /api/telegram/targets', async () => {
    mockApi.mockResolvedValue({ campaigns: [], dm: { linked: true } });
    const t = await fetchTargets();
    expect(mockApi).toHaveBeenCalledWith('GET', '/api/telegram/targets');
    expect(t.dm.linked).toBe(true);
  });

  it('renderTargetOptions includes campaigns and dm', () => {
    const html = renderTargetOptions({ campaigns: [{ id: 1, title: 'Curse', chat_title: 'Strahd Table' }], dm: { linked: true } });
    expect(html).toContain('Curse');
    expect(html).toContain('Strahd Table');
    expect(html).toContain('campaign:1');
    expect(html).toContain('My DM chat');
  });

  it('renderTargetOptions omits DM when not linked', () => {
    const html = renderTargetOptions({ campaigns: [{ id: 2, title: 'Tomb', chat_title: '' }], dm: { linked: false } });
    expect(html).toContain('Tomb');
    expect(html).not.toContain('My DM chat');
  });

  it('renderResults shows sent and failed with escaped error', () => {
    const html = renderResults([
      { upload_id: 1, status: 'sent', telegram_message_id: 123 },
      { upload_id: 2, status: 'failed', error: 'File too large <script>' },
    ]);
    expect(html).toContain('File 1 sent');
    expect(html).toContain('#123');
    expect(html).toContain('File 2 failed');
    expect(html).toContain('&lt;script');
  });

  it('renderResults handles empty', () => {
    expect(renderResults([])).toContain('No results');
  });
});
