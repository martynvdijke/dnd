import { describe, it, expect, afterEach } from 'vitest';
import { esc, attrEscape, jsEscape, shouldClearModalChrome } from './dom';

describe('esc', () => {
  it('escapes html', () => {
    expect(esc('<b>')).toContain('&lt;');
  });
});

describe('attrEscape', () => {
  it('escapes double and single quotes', () => {
    expect(attrEscape(`a"b'c`)).toBe('a&quot;b&#39;c');
  });
  it('escapes & < >', () => {
    expect(attrEscape('&<>')).toBe('&amp;&lt;&gt;');
  });
});

describe('jsEscape', () => {
  it('escapes single quote', () => {
    expect(jsEscape("a'b")).toBe("a\\'b");
  });
  it('escapes backslash', () => {
    expect(jsEscape('a\\b')).toBe('a\\\\b');
  });
  it('escapes double quote', () => {
    expect(jsEscape('a"b')).toBe('a\\"b');
  });
  it('breaks </script>', () => {
    expect(jsEscape('</script>')).toContain('\\/script');
    expect(jsEscape('</script>')).not.toContain('</script>');
    expect(jsEscape('</SCRIPT>')).not.toContain('</SCRIPT>');
  });
  it('escapes newlines and < > &', () => {
    expect(jsEscape('a\nb\rc<d>e&f')).toBe('a\\nb\\rc\\x3cd\\x3ee\\x26f');
  });
});

describe('shouldClearModalChrome', () => {
  afterEach(() => {
    document.body.innerHTML = '';
    document.body.className = '';
  });

  it('keeps chrome while the target modal is still visible', () => {
    document.body.innerHTML = '<div id="genericModal" class="modal show"></div>';
    const el = document.getElementById('genericModal');
    expect(shouldClearModalChrome(el)).toBe(false);
  });

  it('keeps chrome while another modal is still visible', () => {
    document.body.innerHTML = '<div id="genericModal" class="modal"></div><div id="other" class="modal show"></div>';
    expect(shouldClearModalChrome(document.getElementById('genericModal'))).toBe(false);
  });

  it('allows chrome cleanup when nothing is visible', () => {
    document.body.innerHTML = '<div id="genericModal" class="modal"></div>';
    expect(shouldClearModalChrome(document.getElementById('genericModal'))).toBe(true);
  });

  it('allows chrome cleanup when the modal element is missing', () => {
    expect(shouldClearModalChrome(null)).toBe(true);
  });
});
