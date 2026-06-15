import { describe, it, expect } from 'vitest';
import { renderSchemaEntries, entryPreview } from './compendium';

describe('renderSchemaEntries', () => {
  it('returns card markup for an entry with name', () => {
    const schema = {
      id: 1,
      entry_count: 5,
      entries: [{ data: { name: 'Fireball', level: 3, school: 'evocation' } }],
    };
    const html = renderSchemaEntries(schema);
    expect(html).toContain('Fireball');
    expect(html).toContain('card');
    expect(html).toContain('card-body');
  });

  it('includes View All link when entry_count > 20', () => {
    const schema = {
      id: 2,
      entry_count: 25,
      entries: [
        { data: { name: 'Entry 1' } },
        { data: { name: 'Entry 2' } },
      ],
    };
    const html = renderSchemaEntries(schema);
    expect(html).toContain('View All');
    expect(html).toContain('25 entries');
    expect(html).toContain('/api/compendium/schemas/2/entries');
  });

  it('omits View All link when entry_count <= 20', () => {
    const schema = {
      id: 3,
      entry_count: 5,
      entries: [{ data: { name: 'Test' } }],
    };
    const html = renderSchemaEntries(schema);
    expect(html).not.toContain('View All');
  });

  it('returns empty string for empty entries array', () => {
    const schema = { id: 4, entry_count: 0, entries: [] };
    expect(renderSchemaEntries(schema)).toBe('');
  });

  it('handles entries without data gracefully', () => {
    const schema = {
      id: 5,
      entry_count: 1,
      entries: [{ data: null }],
    };
    const html = renderSchemaEntries(schema);
    expect(html).toContain('Unnamed');
    expect(html).toContain('card');
  });

  it('handles entry with missing name fields', () => {
    const schema = {
      id: 6,
      entry_count: 1,
      entries: [{ data: { description: 'no name here' } }],
    };
    const html = renderSchemaEntries(schema);
    expect(html).toContain('Unnamed');
  });
});

describe('entryPreview', () => {
  it('returns up to 3 key-value pairs', () => {
    const data = { name: 'Fireball', level: 3, school: 'evocation', damage: '8d6', range: '150 ft' };
    const preview = entryPreview(data, 'Fireball');
    // Should include 3 items (skips name), joined by ·
    const parts = preview.split(' · ');
    expect(parts.length).toBe(3);
  });

  it('skips the name field', () => {
    const data = { name: 'Fireball', level: 3 };
    const preview = entryPreview(data, 'Fireball');
    expect(preview).not.toContain('name:');
    expect(preview).toContain('level:');
  });

  it('returns empty string for null data', () => {
    expect(entryPreview(null, 'Test')).toBe('');
  });

  it('returns empty string for undefined data', () => {
    expect(entryPreview(undefined, 'Test')).toBe('');
  });

  it('returns empty string for empty data object', () => {
    expect(entryPreview({}, 'Test')).toBe('');
  });

  it('handles string values', () => {
    const data = { name: 'Foo', description: 'A test item' };
    const preview = entryPreview(data, 'Foo');
    expect(preview).toContain('description:');
    expect(preview).toContain('A test item');
  });

  it('handles numeric values', () => {
    const data = { name: 'Foo', level: 5 };
    const preview = entryPreview(data, 'Foo');
    expect(preview).toContain('level: 5');
  });

  it('skips long string values (>80 chars)', () => {
    const data = { name: 'Foo', longField: 'x'.repeat(100) };
    const preview = entryPreview(data, 'Foo');
    expect(preview).not.toContain('longField');
  });

  it('returns at most 3 items even when more than 3 non-name fields exist', () => {
    const data = { name: 'Test', a: '1', b: '2', c: '3', d: '4', e: '5' };
    const preview = entryPreview(data, 'Test');
    const parts = preview.split(' · ');
    expect(parts.length).toBeLessThanOrEqual(3);
    // Should have exactly 3 since we have enough fields
    expect(parts.length).toBe(3);
  });
});
