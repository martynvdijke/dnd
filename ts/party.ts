// @ts-nocheck — extracted from app.ts, window-level self-registration
import { showView } from './navigation';
import { esc, showModal, hideModal, toast } from './lib/dom';
import { api } from './lib/api';
import { currentUser, currentChar } from './lib/state';
import { expose } from './lib/expose';

// ─── Party View & Campaign Management ───

// DM notes cache, keyed by character id, populated while rendering so notes can
// be passed to the notes modal without fragile inline-string escaping.
const dmNotesCache: Record<number, string> = {};
const dmNotesNames: Record<number, string> = {};

expose('showParty', async function () {
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
      // Admin can always view/edit DM notes, even for uncategorized characters.
      const canNotes = dm || currentUser?.role === 'admin';
      const canOpen = (userId: number) => userId === currentUser?.id || currentUser?.role === 'admin' || dm;
      const partyLabel = g.party_name ? esc(g.party_name) : esc(g.name || 'Unnamed Campaign');
      const subLabel = g.party_name ? `<span class="small text-muted ms-2">Campaign: ${esc(g.name)}</span>` : '';

      // Party overview summary
      const members = g.members || [];
      const totalHp = members.reduce((s:number, m:any) => s + (m.hp_current || 0), 0);
      const totalHpMax = members.reduce((s:number, m:any) => s + (m.hp_max || 0), 0);
      const avgLevel = members.length ? Math.round(members.reduce((s:number, m:any) => s + (m.level || 0), 0) / members.length) : 0;
      const downed = members.filter((m:any) => m.status === 'down').length;
      const injured = members.filter((m:any) => m.status === 'injured').length;
      const summaryChips = members.length ? `
        <div class="d-flex flex-wrap gap-3 small text-muted mt-2 mb-1">
          <span><i class="fa-solid fa-arrow-up me-1" aria-hidden="true"></i>Avg Lv ${avgLevel}</span>
          <span><i class="fa-solid fa-heart-pulse me-1" aria-hidden="true"></i>HP ${totalHp}/${totalHpMax}</span>
          ${injured ? `<span style="color:var(--gold)"><i class="fa-solid fa-bandage me-1" aria-hidden="true"></i>${injured} injured</span>` : ''}
          ${downed ? `<span style="color:var(--danger)"><i class="fa-solid fa-skull me-1" aria-hidden="true"></i>${downed} down</span>` : ''}
        </div>` : '';
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
          ${summaryChips}
          <div class="row g-3">
            ${g.members.map((m:any) => {
              const pct = m.hp_max > 0 ? Math.round((m.hp_current / m.hp_max) * 100) : 0;
              const sc = m.status === 'down' ? 'var(--danger)' : m.status === 'injured' ? 'var(--gold)' : 'var(--success)';
              const isLinked = m.character_type === 'linked';
              const clickable = canOpen(m.user_id) && !isLinked;
              if (canNotes) {
                dmNotesCache[m.id] = m.dm_notes || '';
                dmNotesNames[m.id] = m.name;
              }
              return `<div class="col-md-6 col-lg-4">
                <div class="character-card" ${clickable ? `onclick="openChar(${m.id})"` : ''} style="${clickable ? '' : 'cursor:default;opacity:0.75'}">
                  <div class="d-flex align-items-center gap-2 mb-1">
                    ${m.portrait_url ? `<img src="${esc(m.portrait_url)}" class="character-portrait" style="width:28px;height:28px;object-fit:cover;border-radius:50%" alt="">` : ''}
                    <div class="char-name" style="font-size:0.95rem">${esc(m.name)}</div>
                  </div>
                  <div class="char-detail">
                    ${m.race_color ? `<span class="badge" style="background:${m.race_color};color:#fff">${esc(m.race)}</span>` : esc(m.race)}
                    ${esc(m.class)} · Level ${m.level}
                    ${isLinked ? '<span class="badge bg-secondary ms-1" title="Linked character — view only from party view">linked</span>' : ''}
                  </div>
                  ${m.owner_name && m.owner_name !== currentUser?.username ? `<div class="small text-muted"><i class="fa-solid fa-user me-1"></i>${esc(m.owner_name)}</div>` : ''}
                  <div class="d-flex gap-3 mt-1 small text-muted">
                    <span>AC: ${m.ac}</span><span style="color:${sc}">${esc(m.status)}</span>
                  </div>
                  <div class="hp-bar position-relative mt-2" style="height:12px">
                    <div class="hp-bar-fill" style="width:${pct}%;height:100%"></div>
                    <div class="position-absolute top-0 start-0 end-0 bottom-0 d-flex align-items-center justify-content-center text-white" style="font-size:0.65rem">${m.hp_current}/${m.hp_max}</div>
                  </div>
                  <div class="d-flex gap-1 mt-2">
                    ${isLinked ? `<button class="btn btn-sm btn-outline-primary" onclick="event.stopPropagation();showCharStatsModal(${m.id})"><i class="fa-solid fa-eye me-1"></i>View Stats</button>` : ''}
                    ${canNotes ? `<button class="btn btn-sm btn-outline-secondary" onclick="event.stopPropagation();showCharNotes(${m.id})" title="Private DM notes"><i class="fa-solid fa-note-sticky me-1"></i>DM Notes</button>` : ''}
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
            <button class="btn btn-sm btn-outline-primary" onclick="showRaceColors()"><i class="fa-solid fa-palette me-1"></i>Race Colors</button>
          </div>
        </div>
        ` : ''}
      </div>`;
    }).join('') || '<div class="empty-state"><i class="fa-solid fa-flag fa-2x mb-2 d-block text-muted"></i>No characters yet. Create a campaign and add members to build your party!</div>';

    el.innerHTML = html;
  } catch (e:any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Failed: ${esc(e.message)}</p></div>`;
  }
});

// ─── Roster picker (campaign characters) ───
//
// Reusable grouped multi-select for campaign rosters: player-controlled
// characters (owned by the current user) and external characters (owned by
// other campaign members). Loads candidates from the roster candidates
// endpoint; characters already in the roster render pre-selected.

async function loadRosterCandidates(campaignId: number): Promise<any[]> {
  return api('GET', `/api/campaigns/${campaignId}/character-candidates`);
}

function rosterCandidateRow(ch: any): string {
  const inRoster = !!ch.in_roster;
  return `
    <div class="form-check roster-candidate" data-testid="roster-candidate-${ch.id}">
      <input class="form-check-input roster-cb" type="checkbox" id="rosterCb-${ch.id}" data-id="${ch.id}" ${inRoster ? 'checked' : ''} ${inRoster ? 'disabled' : ''}>
      <label class="form-check-label d-flex align-items-center gap-2 flex-wrap" for="rosterCb-${ch.id}">
        ${ch.portrait_url ? `<img src="${esc(ch.portrait_url)}" class="character-portrait" style="width:24px;height:24px;object-fit:cover;border-radius:50%" alt="">` : ''}
        <span><strong>${esc(ch.name)}</strong></span>
        <span class="text-muted small">${esc(ch.race)} ${esc(ch.class)} · Level ${ch.level}</span>
        <span class="badge ${ch.owned ? 'bg-success' : 'bg-secondary'}">${ch.owned ? 'Player-controlled' : 'External'}</span>
        ${!ch.owned ? `<span class="badge bg-info"><i class="fa-solid fa-user me-1"></i>${esc(ch.owner_username)}</span>` : ''}
        ${inRoster ? '<span class="badge badge-gold">In roster</span>' : ''}
      </label>
    </div>`;
}

function rosterPickerHtml(candidates: any[]): string {
  if (!candidates.length) {
    return '<p class="small text-muted mb-2">No characters available yet. Create characters, then add members so their characters can join the roster.</p>';
  }
  const own = candidates.filter((ch: any) => ch.owned);
  const external = candidates.filter((ch: any) => !ch.owned);
  const section = (title: string, emptyNote: string, items: any[]) => `
    <div class="mb-2">
      <div class="small fw-bold text-muted mb-1">${title}</div>
      ${items.length ? items.map(rosterCandidateRow).join('') : `<p class="small text-muted fst-italic">${emptyNote}</p>`}
    </div>`;
  return section('Your Characters (player-controlled)', 'No characters yet.', own) +
         section("Campaign Members' Characters (external)", "No other members' characters yet.", external);
}

function getSelectedRosterIds(): number[] {
  const ids: number[] = [];
  document.querySelectorAll<HTMLInputElement>('input.roster-cb:checked:not(:disabled)').forEach((cb) => {
    const id = parseInt(cb.dataset.id || '0', 10);
    if (id) ids.push(id);
  });
  return ids;
}

expose('showCreateCampaign', function () {
  showModal('Create Campaign', `
    <div class="mb-3"><label class="form-label">Campaign Name</label><input class="form-control" id="newCampaignName"></div>
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="newPartyName" placeholder="e.g. The Dawnbringers"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newCampaignDesc" rows="2"></textarea></div>
    <div class="mb-3"><label class="form-label">DM Notes</label><textarea class="form-control" id="newCampaignDmNotes" rows="2" placeholder="Private notes for the Dungeon Master"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doCreateCampaign()">Create</button>
  `);
});

expose('doCreateCampaign', async function () {
  try {
    const name = (document.getElementById('newCampaignName') as HTMLInputElement).value;
    if (!name) { toast('Name required', true); return; }
    const partyName = (document.getElementById('newPartyName') as HTMLInputElement).value;
    const description = (document.getElementById('newCampaignDesc') as HTMLTextAreaElement).value;
    const dmNotes = (document.getElementById('newCampaignDmNotes') as HTMLTextAreaElement).value;
    const created = await api('POST', '/api/campaigns', { name, party_name: partyName, description, dm_notes: dmNotes });
    // Step 2: pick characters to attach to the new campaign's roster.
    try {
      const candidates = await loadRosterCandidates(created.id);
      showModal(`Roster: ${esc(created.name)}`, `
        <p class="small text-muted mb-2">Select one or more characters to attach to this campaign. You can change the roster any time from Manage.</p>
        ${rosterPickerHtml(candidates)}
        <button class="btn btn-gold w-100 mt-2" data-testid="roster-picker-confirm" onclick="finishCreateCampaign(${created.id})"><i class="fa-solid fa-check me-1"></i>Done</button>
      `);
    } catch {
      // Roster picker unavailable — the campaign itself was created.
      hideModal();
      toast('Campaign created');
      (window as any).showParty();
    }
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('finishCreateCampaign', async function (campaignId: number) {
  const ids = getSelectedRosterIds();
  try {
    for (const id of ids) {
      await api('POST', `/api/campaigns/${campaignId}/characters`, { character_id: id });
    }
    hideModal();
    toast(ids.length ? `Campaign created with ${ids.length} character(s) in the roster` : 'Campaign created');
    (window as any).showParty();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('showManageCampaign', async function (campaignId: number, name: string, partyName: string = '') {
  const [campaigns, members, roster] = await Promise.all([
    api('GET', '/api/campaigns'),
    api('GET', `/api/campaigns/${campaignId}/members`).catch(() => []),
    api('GET', `/api/campaigns/${campaignId}/characters`).catch(() => []),
  ]);
  const c = campaigns.find((x: any) => x.id === campaignId);
  const curName = (c && c.name) || name;
  const curPartyName = (c && c.party_name) || partyName;
  const curDesc = (c && c.description) || '';
  const curDmNotes = (c && c.dm_notes) || '';
  const ownerNames: Record<number, string> = {};
  (members as any[]).forEach((m: any) => { ownerNames[m.user_id] = m.username; });
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
  const rosterHtml = roster.length
    ? `<ul class="list-group mb-3">${roster.map((ch: any) => {
        // `owned` from the characters endpoint reflects edit rights (true for
        // admins/DM), so derive player-controlled vs external from ownership.
        const isOwned = ch.user_id === currentUser?.id;
        return `
        <li class="list-group-item d-flex justify-content-between align-items-center" data-testid="roster-member-${ch.id}">
          <span class="d-flex align-items-center gap-2 flex-wrap">
            ${ch.portrait_url ? `<img src="${esc(ch.portrait_url)}" class="character-portrait" style="width:24px;height:24px;object-fit:cover;border-radius:50%" alt="">` : ''}
            <strong>${esc(ch.name)}</strong>
            <span class="text-muted small">${esc(ch.race)} ${esc(ch.class)} · Level ${ch.level}</span>
            <span class="badge ${isOwned ? 'bg-success' : 'bg-secondary'}">${isOwned ? 'Player-controlled' : 'External'}</span>
            ${!isOwned ? `<span class="badge bg-info"><i class="fa-solid fa-user me-1"></i>${esc(ownerNames[ch.user_id] || '')}</span>` : ''}
          </span>
          <button class="btn btn-outline-danger btn-sm" data-testid="roster-remove-${ch.id}" onclick="doRemoveRosterCharacter(${campaignId}, ${ch.id})" title="Remove from roster"><i class="fa-solid fa-xmark"></i></button>
        </li>`;
      }).join('')}</ul>`
    : '<p class="text-muted small mb-3">No characters in this campaign yet. Add characters to build the party.</p>';
  showModal(`Manage: ${esc(curName)}`, `
    <div class="mb-2"><label class="form-label small">Campaign Name</label><input class="form-control" id="editCampaignName" value="${esc(curName)}"></div>
    <div class="mb-2"><label class="form-label small">Party Name</label><input class="form-control" id="editPartyName" value="${esc(curPartyName)}" placeholder="e.g. The Dawnbringers"></div>
    <div class="mb-2"><label class="form-label small">Description</label><textarea class="form-control" id="editCampaignDesc" rows="2">${esc(curDesc)}</textarea></div>
    <div class="mb-3"><label class="form-label small">DM Notes</label><textarea class="form-control" id="editCampaignDmNotes" rows="2">${esc(curDmNotes)}</textarea></div>
    <button class="btn btn-gold w-100 mb-3" onclick="doUpdateCampaign(${campaignId})"><i class="fa-solid fa-floppy-disk me-1"></i>Save Settings</button>
    <hr>
    <h6 class="mb-2"><i class="fa-solid fa-users me-1"></i>Roster</h6>
    ${rosterHtml}
    <button class="btn btn-outline-primary w-100 mb-3" data-testid="roster-add-open" onclick="showRosterPicker(${campaignId})"><i class="fa-solid fa-user-plus me-1"></i>Add Characters</button>
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
});

expose('showRosterPicker', async function (campaignId: number) {
  try {
    const candidates = await loadRosterCandidates(campaignId);
    showModal('Add Characters to Roster', `
      <p class="small text-muted mb-2">Characters already in the roster are pre-selected. Check the ones you want to add; remove characters from the roster list.</p>
      ${rosterPickerHtml(candidates)}
      <button class="btn btn-gold w-100 mt-2" data-testid="roster-picker-confirm" onclick="doRosterPickerConfirm(${campaignId})"><i class="fa-solid fa-plus me-1"></i>Add Selected</button>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('doRosterPickerConfirm', async function (campaignId: number) {
  const ids = getSelectedRosterIds();
  try {
    for (const id of ids) {
      await api('POST', `/api/campaigns/${campaignId}/characters`, { character_id: id });
    }
    toast(ids.length ? `Added ${ids.length} character(s) to the roster` : 'No characters selected');
    // Re-render the manage modal in place (hideModal+show would race
    // Bootstrap's transition, leaving the modal closed).
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('doRemoveRosterCharacter', async function (campaignId: number, characterId: number) {
  if (!confirm('Remove this character from the campaign roster?')) return;
  try {
    await api('DELETE', `/api/campaigns/${campaignId}/characters/${characterId}`);
    toast('Character removed from roster');
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('doUpdateCampaign', async function (campaignId: number) {
  try {
    const name = (document.getElementById('editCampaignName') as HTMLInputElement).value;
    if (!name) { toast('Name required', true); return; }
    const partyName = (document.getElementById('editPartyName') as HTMLInputElement).value;
    const description = (document.getElementById('editCampaignDesc') as HTMLTextAreaElement).value;
    const dmNotes = (document.getElementById('editCampaignDmNotes') as HTMLTextAreaElement).value;
    await api('PUT', `/api/campaigns/${campaignId}`, { name, party_name: partyName, description, dm_notes: dmNotes });
    toast('Campaign updated');
    (window as any).showParty();
    hideModal();
  } catch (e: any) {
    toast(e.message, true);
  }
});

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

expose('doAddMember', async function (campaignId: number) {
  const username = (document.getElementById('addMemberUsername') as HTMLInputElement).value.trim();
  if (!username) return;
  try {
    await api('POST', `/api/campaigns/${campaignId}/members`, { username });
    toast('Member added');
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('doToggleDm', async function (campaignId: number, userId: number, newRole: string) {
  try {
    await api('PUT', `/api/campaigns/${campaignId}/members/${userId}`, { role: newRole });
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('doRemoveMember', async function (campaignId: number, userId: number) {
  if (!confirm('Remove this member?')) return;
  try {
    await api('DELETE', `/api/campaigns/${campaignId}/members/${userId}`);
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('deleteCampaign', async function (campaignId: number) {
  if (!confirm('Delete this campaign? Characters will be unlinked.')) return;
  try {
    await api('DELETE', `/api/campaigns/${campaignId}`);
    toast('Campaign deleted');
    (window as any).showParty();
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Party Management ───

expose('showCreateParty', function () {
  showModal('Create Party', `
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="newPartyNameInput"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newPartyDesc" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doCreateParty()">Create</button>
  `);
});

expose('doCreateParty', async function () {
  const name = (document.getElementById('newPartyNameInput') as HTMLInputElement).value;
  if (!name) { toast('Party name required', true); return; }
  const description = (document.getElementById('newPartyDesc') as HTMLTextAreaElement).value;
  try {
    await api('POST', '/api/parties', { name, description });
    hideModal();
    toast('Party created');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
});

expose('renameParty', function (id: number, name: string, description: string) {
  showModal('Rename Party', `
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="editPartyNameInput" value="${esc(name)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="editPartyDesc" rows="2">${esc(description)}</textarea></div>
    <button class="btn btn-primary w-100" onclick="doRenameParty(${id})">Save</button>
  `);
});

expose('doRenameParty', async function (id: number) {
  const name = (document.getElementById('editPartyNameInput') as HTMLInputElement).value;
  if (!name) { toast('Party name required', true); return; }
  const description = (document.getElementById('editPartyDesc') as HTMLTextAreaElement).value;
  try {
    await api('PUT', `/api/parties/${id}`, { name, description });
    hideModal();
    toast('Party updated');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
});

expose('deleteParty', async function (id: number) {
  if (!confirm('Delete this party?')) return;
  try {
    await api('DELETE', `/api/parties/${id}`);
    toast('Party deleted');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
});

// ─── Share & Email ───

expose('shareCharacter', async function () {
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
});

expose('copyShareUrl', function () {
  const input = document.getElementById('shareUrl') as HTMLInputElement;
  if (input) {
    input.select();
    navigator.clipboard.writeText(input.value).then(() => toast('Link copied!')).catch(() => {});
  }
});

expose('shareParty', async function (campaignId: number) {
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
});

// ─── Read-only character stats modal (linked characters) ───

expose('showCharStatsModal', async function (charId: number) {
  try {
    const c: any = await api('GET', `/api/characters/${charId}`);
    const mod = (s: number) => Math.floor((s - 10) / 2);
    const statBox = (label: string, val: number) => `
      <div class="text-center px-3 py-2 border rounded">
        <div class="small text-muted">${label}</div>
        <div class="fs-4 fw-bold">${val}</div>
        <div class="small text-muted">${mod(val) >= 0 ? '+' : ''}${mod(val)}</div>
      </div>`;
    const skills = c.skills
      ? `<div class="mt-1">${esc(c.skills)}</div>`
      : '<div class="mt-1 text-muted small">None</div>';
    showModal(`Stats: ${esc(c.name)}`, `
      <div class="mb-3 small text-muted">
        ${esc(c.race)} ${esc(c.class)}${c.subclass ? ` (${esc(c.subclass)})` : ''} · Level ${c.level} · ${c.hp_current}/${c.hp_max} HP · AC ${c.ac}
      </div>
      <div class="d-flex flex-wrap justify-content-center gap-2 mb-3">
        ${statBox('STR', c.str)}${statBox('DEX', c.dex)}${statBox('CON', c.con)}${statBox('INT', c.int)}${statBox('WIS', c.wis)}${statBox('CHA', c.cha)}
      </div>
      <div class="small"><strong>Skills:</strong>${skills}</div>
      <div class="d-flex justify-content-end mt-3">
        <button class="btn btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── DM notes on characters ───

expose('showCharNotes', function (charId: number) {
  const name = dmNotesNames[charId] || 'Character';
  const notes = dmNotesCache[charId] || '';
  showModal(`DM Notes: ${esc(name)}`, `
    <p class="small text-muted mb-2">Private notes only visible to the DM. Great for plot hooks, secrets, or observations about this character.</p>
    <div class="mb-3">
      <textarea class="form-control" id="dmNotesInput" rows="6">${esc(notes)}</textarea>
    </div>
    <button class="btn btn-gold w-100" onclick="saveCharNotes(${charId})"><i class="fa-solid fa-floppy-disk me-1"></i>Save Notes</button>
  `);
});

expose('saveCharNotes', async function (charId: number) {
  const value = (document.getElementById('dmNotesInput') as HTMLTextAreaElement).value;
  try {
    await api('PUT', `/api/characters/${charId}/dm-notes`, { dm_notes: value });
    dmNotesCache[charId] = value;
    hideModal();
    toast('DM notes saved');
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('sendCampaignHighlights', async function (campaignId: number) {
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
});
