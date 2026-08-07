import { describe, it, expect } from 'vitest';
import { renderLinksContent, renderBacklinksBadge } from './backlinks';
import type { LinksResponse, LinkItem } from './backlinks';

function link(overrides: Partial<LinkItem>): LinkItem {
  return {
    id: 1,
    source_type: 'character',
    source_id: 10,
    target_type: 'note',
    target_id: 20,
    context: 'manual',
    created_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('renderLinksContent', () => {
  it('renders empty state when there are no links', () => {
    const html = renderLinksContent({ outgoing: [], backlinks: [] }, 'character', 1);
    expect(html).toContain('No links yet');
  });

  it('tolerates missing outgoing/backlinks arrays', () => {
    const html = renderLinksContent({} as LinksResponse, 'character', 1);
    expect(html).toContain('No links yet');
  });

  it('renders outgoing links section with mention badge', () => {
    const data: LinksResponse = {
      outgoing: [link({ context: 'mention', target_title: 'The Quest', target_type: 'quest', target_url: '/notes/5' })],
      backlinks: [],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).toContain('Links To');
    expect(html).toContain('The Quest');
    expect(html).toContain('mention');
    expect(html).toContain('backlink-entity-link');
  });

  it('falls back to type + id when target_title is missing', () => {
    const data: LinksResponse = {
      outgoing: [link({ target_title: undefined, target_type: 'spell' })],
      backlinks: [],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).toContain('spell #20');
  });

  it('renders a delete button for manual links', () => {
    const data: LinksResponse = {
      outgoing: [link({ context: 'manual' })],
      backlinks: [],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).toContain('backlink-delete-btn');
    expect(html).toContain('data-link-id="1"');
  });

  it('omits delete button and link for mention links without url', () => {
    const data: LinksResponse = {
      outgoing: [link({ context: 'mention', target_url: undefined, target_title: undefined })],
      backlinks: [],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).not.toContain('backlink-delete-btn');
    expect(html).not.toContain('backlink-entity-link');
  });

  it('renders backlinks section with source icon', () => {
    const data: LinksResponse = {
      outgoing: [],
      backlinks: [link({ source_type: 'npc', source_title: 'Goblin', context: 'mention', source_url: '/npcs/3' })],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).toContain('Linked From');
    expect(html).toContain('Goblin');
    expect(html).toContain('mention');
    expect(html).toContain('backlink-entity-link');
  });

  it('falls back to source type + id when source_title missing', () => {
    const data: LinksResponse = {
      outgoing: [],
      backlinks: [link({ source_title: undefined, source_type: 'faction' })],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).toContain('faction #10');
  });

  it('renders both outgoing and backlinks sections together', () => {
    const data: LinksResponse = {
      outgoing: [link({ context: 'manual', target_title: 'Out' })],
      backlinks: [link({ context: 'mention', source_title: 'In' })],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).toContain('Links To');
    expect(html).toContain('Linked From');
    expect(html).toContain('Out');
    expect(html).toContain('In');
  });
});

describe('renderBacklinksBadge', () => {
  it('renders badge with entity type and id', () => {
    const html = renderBacklinksBadge('character', 42);
    expect(html).toContain('backlinks-badge');
    expect(html).toContain('data-type="character"');
    expect(html).toContain('data-id="42"');
  });

  it('escapes entity type in badge', () => {
    const html = renderBacklinksBadge('<script>', 1);
    expect(html).not.toContain('<script>');
  });
});
