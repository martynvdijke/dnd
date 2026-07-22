/**
 * Backlink panel — shows outgoing links and backlinks for any entity.
 *
 * Renders a collapsible panel listing links to/from the current entity,
 * with delete capability for manually-created links.
 */
import { esc, toast } from './lib/dom';
import { api } from './lib/api';
import { navigate } from './router';
import type { ViewState } from './types';

export interface LinkItem {
  id: number;
  source_type: string;
  source_id: number;
  target_type: string;
  target_id: number;
  context: string;
  created_at: string;
  source_title?: string;
  target_title?: string;
  source_url?: string;
  target_url?: string;
}

export interface LinksResponse {
  outgoing: LinkItem[];
  backlinks: LinkItem[];
}

// Map entity types to icon classes.
const TYPE_ICONS: Record<string, string> = {
  character: 'fa-users',
  npc: 'fa-user-group',
  note: 'fa-note-sticky',
  quest: 'fa-scroll',
  journal: 'fa-book-open',
  session: 'fa-calendar',
  campaign: 'fa-flag',
  location: 'fa-map-pin',
  faction: 'fa-flag',
  shop: 'fa-store',
  oneshot: 'fa-scroll',
  encounter: 'fa-crosshairs',
  monster: 'fa-dragon',
  spell: 'fa-wand-sparkles',
  equipment: 'fa-backpack',
  race: 'fa-person',
  class: 'fa-graduation-cap',
  feat: 'fa-star',
  background: 'fa-address-card',
  item: 'fa-box',
};

function linkIcon(type: string): string {
  return TYPE_ICONS[type] || 'fa-file-lines';
}

/**
 * Load and render the backlink panel for a given entity.
 * Inserts HTML into the specified container element.
 */
export async function renderBacklinks(
  container: HTMLElement,
  entityType: string,
  entityId: number,
): Promise<void> {
  container.innerHTML = `
    <div class="backlinks-panel">
      <div class="backlinks-header" onclick="this.nextElementSibling.classList.toggle('d-none');this.querySelector('.backlinks-chevron')?.classList.toggle('collapsed')">
        <h6><i class="fa-solid fa-link me-1"></i> Links</h6>
        <span class="backlinks-chevron"><i class="fa-solid fa-chevron-down"></i></span>
      </div>
      <div class="backlinks-body">
        <div class="backlinks-loading text-muted small py-2">
          <i class="fa-solid fa-spinner fa-spin me-1"></i> Loading links...
        </div>
      </div>
    </div>`;

  // Lazy load on first expand
  const header = container.querySelector('.backlinks-header')!;
  const body = container.querySelector('.backlinks-body')!;
  let loaded = false;

  const loadHandler = async () => {
    if (loaded) return;
    loaded = true;
    header.removeEventListener('click', loadHandler);

    try {
      const data: LinksResponse = await api('GET', `/api/links/${entityType}/${entityId}`);
      body.innerHTML = renderLinksContent(data, entityType, entityId);
      // Bind delete buttons
      body.querySelectorAll('.backlink-delete-btn').forEach((btn) => {
        btn.addEventListener('click', async (e) => {
          e.stopPropagation();
          const linkId = parseInt((btn as HTMLElement).dataset.linkId || '0', 10);
          if (!linkId) return;
          if (!confirm('Remove this link?')) return;
          try {
            await api('DELETE', `/api/links/${linkId}`);
            toast('Link removed');
            // Re-render
            loaded = false;
            body.innerHTML = `
              <div class="backlinks-loading text-muted small py-2">
                <i class="fa-solid fa-spinner fa-spin me-1"></i> Loading links...
              </div>`;
            header.addEventListener('click', loadHandler);
            loadHandler();
          } catch (e: any) {
            toast(e.message, true);
          }
        });
      });
      // Bind navigation clicks
      body.querySelectorAll('.backlink-entity-link').forEach((el) => {
        el.addEventListener('click', (e) => {
          e.preventDefault();
          const link = (el as HTMLElement).dataset.href;
          if (link) navigate(link as ViewState);
        });
      });
    } catch (e: any) {
      body.innerHTML = `
        <div class="text-danger small py-2">
          <i class="fa-solid fa-exclamation-triangle me-1"></i>${esc(e.message)}
        </div>`;
    }
  };

  header.addEventListener('click', loadHandler);

  // Auto-expand if it was already open
  if (!body.classList.contains('d-none')) {
    loadHandler();
  }
}

export function renderLinksContent(
  data: LinksResponse,
  entityType: string,
  entityId: number,
): string {
  const outgoing = data.outgoing || [];
  const backlinks = data.backlinks || [];
  const total = outgoing.length + backlinks.length;

  if (total === 0) {
    return `<div class="backlinks-empty small text-muted py-2">
      <i class="fa-solid fa-link-slash me-1"></i> No links yet
    </div>`;
  }

  let html = '';

  if (outgoing.length > 0) {
    html += `<div class="backlinks-section">
      <div class="backlinks-section-label"><i class="fa-solid fa-arrow-right me-1"></i> Links To (${outgoing.length})</div>`;
    html += outgoing
      .map(
        (l) => `
      <div class="backlink-row" data-context="${esc(l.context)}">
        <i class="fa-solid ${linkIcon(l.target_type)} backlink-type-icon"></i>
        <span class="backlink-title">${esc(l.target_title || l.target_type + ' #' + l.target_id)}</span>
        ${l.context === 'mention' ? '<span class="badge badge-muted">mention</span>' : ''}
        ${l.target_url ? `<a href="#" class="backlink-entity-link" data-href="${esc(l.target_url)}"><i class="fa-solid fa-up-right-from-square"></i></a>` : ''}
        ${l.context === 'manual' ? `<button class="btn btn-sm btn-outline-danger backlink-delete-btn" data-link-id="${l.id}" title="Remove link">&times;</button>` : ''}
      </div>`,
      )
      .join('');
    html += '</div>';
  }

  if (backlinks.length > 0) {
    html += `<div class="backlinks-section">
      <div class="backlinks-section-label"><i class="fa-solid fa-arrow-left me-1"></i> Linked From (${backlinks.length})</div>`;
    html += backlinks
      .map(
        (l) => `
      <div class="backlink-row" data-context="${esc(l.context)}">
        <i class="fa-solid ${linkIcon(l.source_type)} backlink-type-icon"></i>
        <span class="backlink-title">${esc(l.source_title || l.source_type + ' #' + l.source_id)}</span>
        ${l.context === 'mention' ? '<span class="badge badge-muted">mention</span>' : ''}
        ${l.source_url ? `<a href="#" class="backlink-entity-link" data-href="${esc(l.source_url)}"><i class="fa-solid fa-up-right-from-square"></i></a>` : ''}
      </div>`,
      )
      .join('');
    html += '</div>';
  }

  return html;
}

/**
 * Render a minimal inline backlinks indicator for list views.
 * Shows a count badge if links exist.
 */
export function renderBacklinksBadge(
  entityType: string,
  entityId: number,
): string {
  return `<span class="backlinks-badge" data-type="${esc(entityType)}" data-id="${entityId}">
    <i class="fa-solid fa-link text-muted"></i>
  </span>`;
}

/**
 * Load and update backlinks badges for list view items.
 * Queries link counts in bulk (or individually) and appends counts.
 */
export async function hydrateBacklinksBadges(): Promise<void> {
  document.querySelectorAll('.backlinks-badge').forEach(async (el) => {
    const type = (el as HTMLElement).dataset.type;
    const id = parseInt((el as HTMLElement).dataset.id || '0', 10);
    if (!type || !id) return;
    try {
      const data: LinksResponse = await api('GET', `/api/links/${type}/${id}`);
      const total = (data.outgoing?.length || 0) + (data.backlinks?.length || 0);
      if (total > 0) {
        (el as HTMLElement).innerHTML = `<span class="badge badge-gold">${total}</span>`;
      } else {
        (el as HTMLElement).innerHTML = '';
      }
    } catch {
      // silently degrade
    }
  });
}
