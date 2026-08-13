import { esc, showModal, toast } from './lib/dom';
import { api } from './lib/api';
import { expose } from './lib/expose';

interface ShareLink {
  token: string;
  url: string;
  entity_type: string;
  entity_id: number;
  label?: string;
  created_at: string;
  expires_at?: string;
}

const TYPE_LABELS: Record<string, string> = {
  character: 'Character',
  party: 'Party',
  note: 'Note',
  journal: 'Journal',
  map: 'Map',
  upload: 'File',
};

function typeLabel(t: string): string {
  return TYPE_LABELS[t] || t;
}

// shareEntity creates a share link for any supported entity and shows the URL
// modal so the user can copy or email it.
export async function shareEntity(entityType: string, entityId: number, entityName?: string): Promise<void> {
  try {
    const result = await api('POST', '/api/share', {
      entity_type: entityType,
      entity_id: entityId,
    });
    const name = entityName || result.label || '';
    showModal(`Share ${typeLabel(entityType)}`, `
      <p>Anyone with this link can view it${name ? `: <strong>${esc(name)}</strong>` : ''}.</p>
      <div class="input-group mb-3">
        <input class="form-control" id="shareUrl" value="${esc(result.url)}" readonly onclick="this.select()">
        <button class="btn btn-gold" onclick="copyShareUrl()"><i class="fa-solid fa-copy"></i></button>
      </div>
      <div class="d-flex gap-2">
        <button class="btn btn-primary flex-grow-1" onclick="window.open('mailto:?subject=${encodeURIComponent('Check this out')}&body=${encodeURIComponent(result.url)}','_blank')"><i class="fa-solid fa-envelope me-1"></i>Email</button>
        <button class="btn btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
}

function renderLinkList(links: ShareLink[]): string {
  if (!links.length) {
    return '<p class="text-muted mb-0">No shared links yet. Use the share button on a character, party, note, journal, map, or file to create one.</p>';
  }
  return links.map((l) => `
    <div class="d-flex justify-content-between align-items-center border-bottom py-2" data-testid="share-link-row">
      <div class="min-w-0">
        <div class="text-truncate"><strong>${esc(l.label || l.url)}</strong> <span class="badge badge-muted ms-1">${typeLabel(l.entity_type)}</span></div>
        <a class="small text-break" href="${esc(l.url)}" target="_blank" rel="noopener">${esc(l.url)}</a>
      </div>
      <button class="btn btn-sm btn-outline-danger ms-2 flex-shrink-0" onclick="revokeShareLink('${l.token}')" title="Revoke link"><i class="fa-solid fa-trash"></i></button>
    </div>
  `).join('');
}

// openSharedLinks shows the management dialog for all links the user created.
export async function openSharedLinks(): Promise<void> {
  try {
    const links = await api('GET', '/api/share') as ShareLink[];
    showModal('Shared Links', `
      <div id="shareLinkList" data-testid="shared-links-list">${renderLinkList(links)}</div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
}

// revokeShareLink deletes a link and re-renders the management dialog.
export async function revokeShareLink(token: string): Promise<void> {
  try {
    await api('DELETE', '/api/share/' + encodeURIComponent(token));
    await openSharedLinks();
  } catch (e: any) {
    toast(e.message, true);
  }
}

expose('shareEntity', shareEntity);
expose('openSharedLinks', openSharedLinks);
expose('revokeShareLink', revokeShareLink);
