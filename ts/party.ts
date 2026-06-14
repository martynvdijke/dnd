// @ts-nocheck — extracted from app.ts, window-level self-registration
import { showView } from './navigation';
import { esc, showModal, hideModal, toast } from './lib/dom';
import { api } from './lib/api';

// ─── Party View & Campaign Management ───

(window as any).showParty = async function () {
  showView('party');
  const el = document.getElementById('partyContent')!;
  el.innerHTML = '<div class="ornament mb-3">✧ Assembling the party... ✧</div>';
  try {
    const [groups, campaigns] = await Promise.all([
      api('GET', '/api/party'),
      api('GET', '/api/campaigns'),
    ]);

    const getCampaign = (campaignId: number) => campaigns.find((c: any) => c.id === campaignId);
    const isOwner = (campaignId: number) => { const c = getCampaign(campaignId); return c && c.user_id === currentUser?.id; };
    const isDm = (campaignId: number) => { const c = getCampaign(campaignId); return c && (c.my_role === 'dm' || c.user_id === currentUser?.id); };

    let html = `<div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0"><i class="fa-solid fa-flag me-2"></i>Party View</h1>
      <div class="d-flex gap-2">
        <button class="btn btn-gold btn-sm" onclick="showCreateCampaign()"><i class="fa-solid fa-plus me-1"></i>New Campaign</button>
        ${currentUser?.role === 'dm' || currentUser?.role === 'admin' ? `<button class="btn btn-outline-primary btn-sm" onclick="showCreateParty()"><i class="fa-solid fa-flag me-1"></i>New Party</button>` : ''}
      </div>
    </div>`;

    // DM/Admin: Party management section
    if (currentUser?.role === 'dm' || currentUser?.role === 'admin') {
      try {
        const parties = await api('GET', '/api/parties');
        if (parties.length) {
          html += `<h5 class="mb-2"><i class="fa-solid fa-flag me-1"></i>Parties</h5>`;
          for (const p of parties) {
            const factions = await api('GET', `/api/parties/${p.id}/factions`).catch(() => []);
            const uploads = await api('GET', `/api/parties/${p.id}/uploads`).catch(() => []);
            const fileCount = uploads.length;
            html += `<div class="card mb-3">
              <div class="card-header d-flex justify-content-between align-items-center py-2">
                <span><strong>${esc(p.name)}</strong> ${p.description ? `<span class="text-muted small ms-2">${esc(p.description)}</span>` : ''}</span>
                <div class="d-flex gap-1">
                  <span class="badge badge-gold">${factions.length} factions</span>
                  ${fileCount ? `<span class="badge bg-info">${fileCount} files</span>` : ''}
                  <button class="btn btn-sm btn-outline-primary" onclick="renameParty(${p.id},'${esc(p.name)}','${esc(p.description)}')"><i class="fa-solid fa-pen"></i></button>
                  <button class="btn btn-sm btn-outline-danger" onclick="deleteParty(${p.id})"><i class="fa-solid fa-trash"></i></button>
                </div>
              </div>
              ${factions.length ? `<div class="card-body py-2">
                <div class="small"><strong>Factions:</strong></div>
                <div class="d-flex flex-wrap gap-1 mt-1">${factions.map((f: any) =>
                  `<span class="badge bg-light text-dark border">${esc(f.name)}${f.type ? ` <span class="text-muted">(${esc(f.type)})</span>` : ''}</span>`
                ).join('')}</div>
              </div>` : ''}
              ${uploads.length ? `<div class="card-footer py-1">
                <div class="small text-muted">${fileCount} file(s) uploaded</div>
              </div>` : ''}
            </div>`;
          }
        }
      } catch {}
    }

    // Campaign-based party groups
    html += groups.map((g:any) => {
      const own = g.id ? isOwner(g.id) : false;
      const dm = g.id ? isDm(g.id) : false;
      const canOpen = (userId: number) => userId === currentUser?.id || currentUser?.role === 'admin' || dm;
      const partyLabel = g.party_name ? esc(g.party_name) : esc(g.name || 'Unnamed Campaign');
      const subLabel = g.party_name ? `<span class="small text-muted ms-2">Campaign: ${esc(g.name)}</span>` : '';
      return `<div class="card mb-3">
        <div class="card-header d-flex justify-content-between align-items-center">
          <div>
            <strong>${partyLabel}</strong>
            ${subLabel}
            ${g.owner_name ? `<span class="ms-2 small text-muted">DM: ${esc(g.owner_name)}</span>` : ''}
          </div>
          <div class="d-flex align-items-center gap-2">
            <span class="badge badge-gold">${g.members.length} members</span>
            ${g.id && (own || dm) ? `
              <button class="btn btn-outline-gold btn-sm" onclick="showCampaignDashboard(${g.id},'${esc(g.name)}')" title="Dashboard"><i class="fa-solid fa-chart-simple"></i></button>
              <button class="btn btn-outline-primary btn-sm" onclick="showManageCampaign(${g.id},'${esc(g.name)}','${esc(g.party_name || '')}')" title="Manage"><i class="fa-solid fa-users-gear"></i></button>
              <button class="btn btn-outline-info btn-sm" onclick="shareParty(${g.id})" title="Share Party"><i class="fa-solid fa-share-nodes"></i></button>
            ` : ''}
            ${g.id && own ? `<button class="btn btn-outline-danger btn-sm" onclick="deleteCampaign(${g.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>` : ''}
            ${g.id && (own || dm) && currentUser?.role === 'admin' ? `<button class="btn btn-outline-gold btn-sm" onclick="sendCampaignHighlights(${g.id})" title="Email Highlights"><i class="fa-solid fa-envelope"></i></button>` : ''}
          </div>
        </div>
        <div class="card-body">
          <div class="row g-3">
            ${g.members.map((m:any) => {
              const pct = m.hp_max > 0 ? Math.round((m.hp_current / m.hp_max) * 100) : 0;
              const sc = m.status === 'down' ? 'var(--danger)' : m.status === 'injured' ? 'var(--gold)' : 'var(--success)';
              return `<div class="col-md-6 col-lg-4">
                <div class="character-card" ${canOpen(m.user_id) ? `onclick="openChar(${m.id})"` : ''} style="${canOpen(m.user_id) ? '' : 'cursor:default;opacity:0.75'}">
                  <div class="char-name">${esc(m.name)}</div>
                  <div class="char-detail">${esc(m.race)} ${esc(m.class)} · Level ${m.level}</div>
                  ${m.owner_name && m.owner_name !== currentUser?.username ? `<div class="small text-muted"><i class="fa-solid fa-user me-1"></i>${esc(m.owner_name)}</div>` : ''}
                  <div class="d-flex gap-3 mt-1 small text-muted">
                    <span>AC: ${m.ac}</span><span style="color:${sc}">${esc(m.status)}</span>
                  </div>
                  <div class="hp-bar position-relative mt-2" style="height:12px">
                    <div class="hp-bar-fill" style="width:${pct}%;height:100%"></div>
                    <div class="position-absolute top-0 start-0 end-0 bottom-0 d-flex align-items-center justify-content-center text-white" style="font-size:0.65rem">${m.hp_current}/${m.hp_max}</div>
                  </div>
                </div>
              </div>`;
            }).join('')}
          </div>
        </div>
        ${g.id && (own || dm) ? `
        <div class="card-footer py-2">
          <div class="d-flex gap-2 flex-wrap">
            <button class="btn btn-sm btn-outline-gold" onclick="showPartyInventory(${g.id})"><i class="fa-solid fa-box me-1"></i>Party Inventory</button>
            <button class="btn btn-sm btn-outline-primary" onclick="showSessionPlanner(${g.id})"><i class="fa-solid fa-calendar me-1"></i>Session Planner</button>
            <button class="btn btn-sm btn-outline-gold" onclick="showEncounterDifficulty()"><i class="fa-solid fa-crosshairs me-1"></i>Difficulty</button>
            <button class="btn btn-sm btn-outline-gold" onclick="showTreasureGenerator()"><i class="fa-solid fa-coins me-1"></i>Treasure</button>
          </div>
        </div>
        ` : ''}
      </div>`;
    }).join('') || '<div class="empty-state"><i class="fa-solid fa-flag fa-2x mb-2 d-block text-muted"></i>No characters yet. Create a campaign and add members to build your party!</div>';

    el.innerHTML = html;
  } catch (e:any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Failed: ${esc(e.message)}</p></div>`;
  }
};

(window as any).showCreateCampaign = function () {
  showModal('Create Campaign', `
    <div class="mb-3"><label class="form-label">Campaign Name</label><input class="form-control" id="newCampaignName"></div>
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="newPartyName" placeholder="e.g. The Dawnbringers"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newCampaignDesc" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doCreateCampaign()">Create</button>
  `);
};

(window as any).doCreateCampaign = async function () {
  try {
    const name = (document.getElementById('newCampaignName') as HTMLInputElement).value;
    if (!name) { toast('Name required', true); return; }
    const partyName = (document.getElementById('newPartyName') as HTMLInputElement).value;
    await api('POST', '/api/campaigns', { name, party_name: partyName, description: (document.getElementById('newCampaignDesc') as HTMLTextAreaElement).value });
    hideModal();
    toast('Campaign created');
    (window as any).showParty();
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).showManageCampaign = async function (campaignId: number, name: string, partyName: string = '') {
  const [campaigns, members] = await Promise.all([
    api('GET', '/api/campaigns'),
    api('GET', `/api/campaigns/${campaignId}/members`).catch(() => []),
  ]);
  const c = campaigns.find((x: any) => x.id === campaignId);
  const curPartyName = (c && c.party_name) || partyName;
  const curDesc = (c && c.description) || '';
  const membersHtml = members.length
    ? `<ul class="list-group mb-3">${members.map((m: any) => {
        const isDmMember = m.role === 'dm';
        return `<li class="list-group-item d-flex justify-content-between align-items-center">
          <span>
            <i class="fa-solid ${isDmMember ? 'fa-crown text-gold' : 'fa-user'} me-2"></i>
            ${esc(m.username)}
            ${isDmMember ? '<span class="badge badge-gold ms-2">DM</span>' : ''}
          </span>
          <div class="d-flex gap-1">
            ${m.username !== currentUser?.username ? `
              <button class="btn btn-sm ${isDmMember ? 'btn-outline-secondary' : 'btn-outline-gold'}" onclick="doToggleDm(${campaignId}, ${m.user_id}, '${isDmMember ? 'player' : 'dm'}')" title="${isDmMember ? 'Remove DM' : 'Make DM'}">
                <i class="fa-solid ${isDmMember ? 'fa-user' : 'fa-crown'}"></i>
              </button>
              <button class="btn btn-outline-danger btn-sm" onclick="doRemoveMember(${campaignId}, ${m.user_id})"><i class="fa-solid fa-xmark"></i></button>
            ` : '<span class="text-muted small">(you)</span>'}
          </div>
        </li>`;
      }).join('')}</ul>`
    : '<p class="text-muted mb-3">No members yet. Add players by username.</p>';
  showModal(`Manage: ${esc(name)}`, `
    <div class="mb-2"><label class="form-label small">Campaign Name</label><input class="form-control" id="editCampaignName" value="${esc(name)}"></div>
    <div class="mb-2"><label class="form-label small">Party Name</label><input class="form-control" id="editPartyName" value="${esc(curPartyName)}" placeholder="e.g. The Dawnbringers"></div>
    <div class="mb-3"><label class="form-label small">Description</label><textarea class="form-control" id="editCampaignDesc" rows="2">${esc(curDesc)}</textarea></div>
    <button class="btn btn-gold w-100 mb-3" onclick="doUpdateCampaign(${campaignId})"><i class="fa-solid fa-floppy-disk me-1"></i>Save Settings</button>
    <hr>
    ${membersHtml}
    <div class="input-group mb-3">
      <input class="form-control" id="addMemberUsername" placeholder="Username to add">
      <button class="btn btn-gold" onclick="doAddMember(${campaignId})"><i class="fa-solid fa-plus"></i></button>
    </div>
    <div id="userSuggestions" class="mb-2"></div>
    <button class="btn btn-outline-secondary w-100" onclick="(window as any).showParty();hideModal()">Done</button>
  `);
  const input = document.getElementById('addMemberUsername') as HTMLInputElement;
  if (input) {
    input.addEventListener('input', () => searchUsers(input.value));
  }
};

(window as any).doUpdateCampaign = async function (campaignId: number) {
  try {
    const name = (document.getElementById('editCampaignName') as HTMLInputElement).value;
    if (!name) { toast('Name required', true); return; }
    const partyName = (document.getElementById('editPartyName') as HTMLInputElement).value;
    const description = (document.getElementById('editCampaignDesc') as HTMLTextAreaElement).value;
    await api('PUT', `/api/campaigns/${campaignId}`, { name, party_name: partyName, description });
    toast('Campaign updated');
    (window as any).showParty();
    hideModal();
  } catch (e: any) {
    toast(e.message, true);
  }
};

let searchTimeout: any = null;
function searchUsers(q: string) {
  clearTimeout(searchTimeout);
  if (q.length < 2) { document.getElementById('userSuggestions')!.innerHTML = ''; return; }
  searchTimeout = setTimeout(async () => {
    try {
      const users = await api('GET', `/api/users/search?q=${encodeURIComponent(q)}`);
      const el = document.getElementById('userSuggestions')!;
      el.innerHTML = users.map((u: any) =>
        `<div class="d-flex justify-content-between align-items-center p-1 border-bottom" style="cursor:pointer" onclick="document.getElementById('addMemberUsername')!.value='${esc(u.username)}';el.innerHTML=''">
          <span>${esc(u.username)}</span>
        </div>`
      ).join('');
    } catch {}
  }, 300);
}

(window as any).doAddMember = async function (campaignId: number) {
  const username = (document.getElementById('addMemberUsername') as HTMLInputElement).value.trim();
  if (!username) return;
  try {
    await api('POST', `/api/campaigns/${campaignId}/members`, { username });
    toast('Member added');
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).doToggleDm = async function (campaignId: number, userId: number, newRole: string) {
  try {
    await api('PUT', `/api/campaigns/${campaignId}/members/${userId}`, { role: newRole });
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).doRemoveMember = async function (campaignId: number, userId: number) {
  if (!confirm('Remove this member?')) return;
  try {
    await api('DELETE', `/api/campaigns/${campaignId}/members/${userId}`);
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteCampaign = async function (campaignId: number) {
  if (!confirm('Delete this campaign? Characters will be unlinked.')) return;
  try {
    await api('DELETE', `/api/campaigns/${campaignId}`);
    toast('Campaign deleted');
    (window as any).showParty();
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Party Management ───

(window as any).showCreateParty = function () {
  showModal('Create Party', `
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="newPartyNameInput"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newPartyDesc" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doCreateParty()">Create</button>
  `);
};

(window as any).doCreateParty = async function () {
  const name = (document.getElementById('newPartyNameInput') as HTMLInputElement).value;
  if (!name) { toast('Party name required', true); return; }
  const description = (document.getElementById('newPartyDesc') as HTMLTextAreaElement).value;
  try {
    await api('POST', '/api/parties', { name, description });
    hideModal();
    toast('Party created');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).renameParty = function (id: number, name: string, description: string) {
  showModal('Rename Party', `
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="editPartyNameInput" value="${esc(name)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="editPartyDesc" rows="2">${esc(description)}</textarea></div>
    <button class="btn btn-primary w-100" onclick="doRenameParty(${id})">Save</button>
  `);
};

(window as any).doRenameParty = async function (id: number) {
  const name = (document.getElementById('editPartyNameInput') as HTMLInputElement).value;
  if (!name) { toast('Party name required', true); return; }
  const description = (document.getElementById('editPartyDesc') as HTMLTextAreaElement).value;
  try {
    await api('PUT', `/api/parties/${id}`, { name, description });
    hideModal();
    toast('Party updated');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).deleteParty = async function (id: number) {
  if (!confirm('Delete this party?')) return;
  try {
    await api('DELETE', `/api/parties/${id}`);
    toast('Party deleted');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
};

// ─── Share & Email ───

(window as any).shareCharacter = async function () {
  if (!currentChar) return;
  try {
    const result = await api('POST', '/api/share', {
      entity_type: 'character',
      entity_id: currentChar.id,
    });
    showModal('Share Character', `
      <p>Share this link to let others view <strong>${esc(currentChar.name)}</strong>.</p>
      <div class="input-group mb-3">
        <input class="form-control" id="shareUrl" value="${esc(result.url)}" readonly onclick="this.select()">
        <button class="btn btn-gold" onclick="copyShareUrl()"><i class="fa-solid fa-copy"></i></button>
      </div>
      <div class="d-flex gap-2">
        <button class="btn btn-primary flex-grow-1" onclick="window.open('mailto:?subject=Check out my character ${esc(currentChar.name)}&body=${encodeURIComponent(result.url)}','_blank')"><i class="fa-solid fa-envelope me-1"></i>Email</button>
        <button class="btn btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).copyShareUrl = function () {
  const input = document.getElementById('shareUrl') as HTMLInputElement;
  if (input) {
    input.select();
    navigator.clipboard.writeText(input.value).then(() => toast('Link copied!')).catch(() => {});
  }
};

(window as any).shareParty = async function (campaignId: number) {
  try {
    const result = await api('POST', '/api/share', {
      entity_type: 'party',
      entity_id: campaignId,
    });
    showModal('Share Party', `
      <p>Share this link to let others view your party.</p>
      <div class="input-group mb-3">
        <input class="form-control" id="shareUrl" value="${esc(result.url)}" readonly onclick="this.select()">
        <button class="btn btn-gold" onclick="copyShareUrl()"><i class="fa-solid fa-copy"></i></button>
      </div>
      <div class="d-flex gap-2">
        <button class="btn btn-primary flex-grow-1" onclick="window.open('mailto:?subject=Check out our party&body=${encodeURIComponent(result.url)}','_blank')"><i class="fa-solid fa-envelope me-1"></i>Email</button>
        <button class="btn btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).sendCampaignHighlights = async function (campaignId: number) {
  try {
    const result = await api('POST', '/api/admin/campaign-highlights', { campaign_id: campaignId });
    const msg = result.errors && result.errors.length
      ? `Sent to ${result.sent} recipients, but ${result.errors.length} failed.`
      : `Campaign highlights sent to ${result.sent} recipient(s)!`;
    toast(msg);
    if (result.errors) console.warn('Email errors:', result.errors);
  } catch (e: any) {
    toast(e.message, true);
  }
};
