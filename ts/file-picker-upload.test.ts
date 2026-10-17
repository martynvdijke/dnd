import { describe, it, expect } from 'vitest';
import { FilePicker } from './file-picker';
import { UPLOAD_FIELD } from './app/uploads';

describe('file-picker accepted types', () => {
  it('allows image, audio, video and pdf', () => {
    expect(FilePicker.isAllowedFile({ type: 'image/png', name: 'a.png' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: 'image/jpeg', name: 'b.jpg' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: 'audio/mpeg', name: 'c.mp3' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: 'audio/ogg', name: 'd.ogg' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: 'video/mp4', name: 'e.mp4' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: 'video/webm', name: 'f.webm' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: 'application/pdf', name: 'g.pdf' })).toBe(true);
  });

  it('allows via extension fallback when mime is generic', () => {
    expect(FilePicker.isAllowedFile({ type: 'application/octet-stream', name: 'song.mp3' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: 'application/octet-stream', name: 'clip.m4a' })).toBe(true);
    expect(FilePicker.isAllowedFile({ type: '', name: 'video.mov' })).toBe(true);
  });

  it('rejects disallowed types', () => {
    expect(FilePicker.isAllowedFile({ type: 'text/plain', name: 'notes.txt' })).toBe(false);
    expect(FilePicker.isAllowedFile({ type: 'application/zip', name: 'archive.zip' })).toBe(false);
    expect(FilePicker.isAllowedFile({ type: '', name: 'file.exe' })).toBe(false);
  });
});

describe('modal upload field name', () => {
  it('uses image field so /api/upload accepts it', () => {
    expect(UPLOAD_FIELD).toBe('image');
  });
});
