/**
 * Tests for backlinks module.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

let renderBacklinksBadge: Function;
let renderLinksContent: Function;

beforeEach(async () => {
  const mod = await import('./backlinks');
  renderBacklinksBadge = mod.renderBacklinksBadge;
  renderLinksContent = mod.renderLinksContent;
});

describe('renderBacklinksBadge', () => {
  it('returns a span with data attributes', () => {
    const html = renderBacklinksBadge('character', 42);
    expect(html).toContain('backlinks-badge');
    expect(html).toContain('data-type="character"');
    expect(html).toContain('data-id="42"');
    expect(html).toContain('fa-link');
  });
});

describe('renderLinksContent', () => {
  const sampleData = {
    outgoing: [
      { id: 1, source_type: 'character', source_id: 1, target_type: 'npc', target_id: 2, context: 'manual', created_at: '', target_title: 'Bob the NPC', target_url: '/npc/2' },
    ],
    backlinks: [
      { id: 2, source_type: 'note', source_id: 3, target_type: 'character', target_id: 1, context: 'mention', created_at: '', source_title: 'Campaign Notes', source_url: '/note/3' },
    ],
  };

  it('renders outgoing section', () => {
    const html = renderLinksContent(sampleData, 'character', 1);
    expect(html).toContain('Links To');
    expect(html).toContain('Bob the NPC');
    expect(html).toContain('backlink-delete-btn');
  });

  it('renders backlinks section with outgoing delete buttons only', () => {
    const html = renderLinksContent(sampleData, 'character', 1);
    expect(html).toContain('Linked From');
    expect(html).toContain('Campaign Notes');
    // There should be one delete button (from outgoing manual link)
    expect((html.match(/backlink-delete-btn/g) || []).length).toBe(1);
  });

  it('shows mention badge for mention context', () => {
    const html = renderLinksContent(sampleData, 'character', 1);
    expect(html).toContain('mention');
  });

  it('shows empty state when no links', () => {
    const html = renderLinksContent({ outgoing: [], backlinks: [] }, 'character', 1);
    expect(html).toContain('No links yet');
    expect(html).toContain('fa-link-slash');
  });

  it('shows delete button only for manual links', () => {
    const data = {
      outgoing: [{ id: 1, source_type: 'character', source_id: 1, target_type: 'npc', target_id: 2, context: 'mention', created_at: '', target_title: 'NPC', target_url: '' }],
      backlinks: [],
    };
    const html = renderLinksContent(data, 'character', 1);
    expect(html).not.toContain('backlink-delete-btn');
  });
});
