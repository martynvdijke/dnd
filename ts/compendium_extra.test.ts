import { describe, it, expect, beforeEach } from 'vitest';
import { humanizeKey, parseCompendiumLinksInto, renderSchemaEntries, entryPreview } from './compendium';

describe('humanizeKey', () => {
  it('replaces underscores and hyphens', () => {
    expect(humanizeKey('spell_level')).toBe('Spell Level');
    expect(humanizeKey('hit-points')).toBe('Hit Points');
    expect(humanizeKey('saving_throw-dc')).toBe('Saving Throw Dc');
  });
  it('capitalizes words', () => {
    expect(humanizeKey('name')).toBe('Name');
    expect(humanizeKey('foo')).toBe('Foo');
  });
  it('handles empty', () => {
    expect(humanizeKey('')).toBe('');
  });
});

describe('renderSchemaEntries edge', () => {
  it('renders multiple entries', () => {
    const html = renderSchemaEntries({
      id: 1, entry_count: 2,
      entries: [{ id: 1, data: { name: 'A', level: 1 } }, { id: 2, data: { name: 'B', level: 2 } }],
    });
    expect(html).toContain('A');
    expect(html).toContain('B');
  });
  it('caps View All at >20', () => {
    const html = renderSchemaEntries({ id: 9, entry_count: 21, entries: [{ id: 1, data: { name: 'X' } }] });
    expect(html).toContain('View All');
  });
  it('no View All at exactly 20', () => {
    const html = renderSchemaEntries({ id: 9, entry_count: 20, entries: [{ id: 1, data: { name: 'X' } }] });
    expect(html).not.toContain('View All');
  });
  it('handles missing entries array', () => {
    expect(renderSchemaEntries({ id: 1, entry_count: 0 })).toBe('');
  });
});

describe('entryPreview cap', () => {
  it('returns empty for null', () => {
    expect(entryPreview(null, 'X')).toBe('');
  });
  it('skips long strings and empty strings', () => {
    const p = entryPreview({ name: 'Foo', long: 'x'.repeat(100), empty: '', good: 'hi' }, 'Foo');
    expect(p).not.toContain('long');
    expect(p).not.toContain('empty');
    expect(p).toContain('good');
  });
  it('caps at 3 parts', () => {
    const p = entryPreview({ name: 'T', a: '1', b: '2', c: '3', d: '4' }, 'T');
    expect(p.split(' · ').length).toBeLessThanOrEqual(3);
  });
  it('includes numbers', () => {
    expect(entryPreview({ name: 'T', level: 5 }, 'T')).toContain('level: 5');
  });
});

describe('parseCompendiumLinksInto', () => {
  beforeEach(() => { document.body.innerHTML = ''; });

  it('replaces [[compendium:type:name]] with anchor', () => {
    const div = document.createElement('div');
    div.textContent = 'See [[compendium:spell:Fireball]] for details';
    document.body.appendChild(div);
    parseCompendiumLinksInto(div);
    const a = div.querySelector('a.compendium-link') as HTMLElement;
    expect(a).not.toBeNull();
    expect((a as any).dataset.schema).toBe('spell');
    expect((a as any).dataset.name).toBe('Fireball');
    expect(a!.textContent).toBe('Fireball');
  });

  it('handles multiple links', () => {
    const div = document.createElement('div');
    div.textContent = '[[compendium:monster:Goblin]] and [[compendium:spell:Fireball]]';
    document.body.appendChild(div);
    parseCompendiumLinksInto(div);
    expect(div.querySelectorAll('a.compendium-link').length).toBe(2);
  });

  it('is no-op when already has compendium-link', () => {
    const div = document.createElement('div');
    div.innerHTML = '<a class="compendium-link" data-schema="spell" data-name="Fireball">Fireball</a> [[compendium:spell:Ice]]';
    parseCompendiumLinksInto(div);
    // should not parse further because early return
    expect(div.querySelectorAll('a.compendium-link').length).toBe(1);
  });

  it('is no-op when no link pattern', () => {
    const div = document.createElement('div');
    div.textContent = 'plain text';
    parseCompendiumLinksInto(div);
    expect(div.innerHTML).toContain('plain text');
    expect(div.querySelector('a')).toBeNull();
  });

  it('handles null-ish root gracefully', () => {
    expect(() => parseCompendiumLinksInto(null as any)).not.toThrow();
  });
});
