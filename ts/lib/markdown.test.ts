import { describe, it, expect } from 'vitest';
import { renderMarkdown } from './markdown';

describe('renderMarkdown', () => {
  it('strips script tags', () => {
    const out = renderMarkdown('<script>alert(1)</script>');
    expect(out).not.toContain('<script');
    expect(out).not.toContain('alert(1)');
  });

  it('strips onerror attribute from img', () => {
    const out = renderMarkdown('<img src=x onerror=alert(1)>');
    expect(out.toLowerCase()).not.toContain('onerror');
    expect(out).not.toContain('alert(1)');
  });

  it('strips javascript: href', () => {
    const out = renderMarkdown('[click](javascript:alert(1))');
    expect(out.toLowerCase()).not.toContain('javascript:');
  });

  it('strips vbscript and data urls', () => {
    expect(renderMarkdown('[x](vbscript:alert(1))').toLowerCase()).not.toContain('vbscript:');
    expect(renderMarkdown('<img src="data:text/html,<script>alert(1)</script>">').toLowerCase()).not.toContain('data:');
  });

  it('strips iframe/object/form tags', () => {
    expect(renderMarkdown('<iframe src="https://evil.com"></iframe>')).not.toContain('<iframe');
    expect(renderMarkdown('<object data="x"></object>')).not.toContain('<object');
    expect(renderMarkdown('<form action="/x"><input></form>')).not.toContain('<form');
  });

  it('strips on* attributes generically', () => {
    const out = renderMarkdown('<a href="https://example.com" onclick="alert(1)" onmouseover="alert(2)">hi</a>');
    expect(out.toLowerCase()).not.toContain('onclick');
    expect(out.toLowerCase()).not.toContain('onmouseover');
  });

  it('preserves headings', () => {
    const out = renderMarkdown('# Hello');
    expect(out).toContain('<h1');
    expect(out).toContain('Hello');
  });

  it('preserves bold', () => {
    const out = renderMarkdown('**bold**');
    expect(out).toContain('<strong>bold</strong>');
  });

  it('preserves https links', () => {
    const out = renderMarkdown('[example](https://example.com)');
    expect(out).toContain('href="https://example.com"');
  });

  it('preserves lists', () => {
    const out = renderMarkdown('- a\n- b');
    expect(out).toContain('<ul>');
    expect(out).toContain('<li>');
  });

  it('preserves code blocks', () => {
    const out = renderMarkdown('`code`');
    expect(out).toContain('<code>code</code>');
  });

  it('preserves tables', () => {
    const out = renderMarkdown('| a | b |\n|---|---|\n| 1 | 2 |');
    expect(out).toContain('<table>');
  });

  it('preserves images with https and relative uploads', () => {
    const https = renderMarkdown('![alt](https://example.com/img.png)');
    expect(https).toContain('src="https://example.com/img.png"');
    const rel = renderMarkdown('![alt](/uploads/img.png)');
    expect(rel).toContain('src="/uploads/img.png"');
  });
});
