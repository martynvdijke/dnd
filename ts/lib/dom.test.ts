import { describe, it, expect } from 'vitest';
import { esc, attrEscape, jsEscape } from './dom';

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
