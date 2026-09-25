import { expose } from '../lib/expose';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import { showView } from '../navigation';
import { playAmbience, stopAmbience, listAmbienceTracks } from '../lib/ambience';

// Campaign Completeness Enhancements
// ═══════════════════════════════════════════

// ─── Campaign Dashboard ───

expose('showCampaignDashboard', async function (campaignId: number, campaignName: string) {
  showModal(`${esc(campaignName)} Dashboard`, `<div id="campaignDashContent"><div class="ornament">✧ Loading dashboard... ✧</div></div>`);
  try {
    const d = await api('GET', `/api/campaigns/${campaignId}/dashboard`);
    const hpPct = (h: number, m: number) => m > 0 ? Math.round((h / m) * 100) : 0;
    const avatarLetter = (n: string) => (n || '?').charAt(0).toUpperCase();

    const content = `
      <div class="dash-grid">
        <div class="dash-card">
          <h6>Characters</h6>
          ${(d.characters || []).map((ch: any) => `
            <div class="dash-char-card" onclick="openChar(${ch.id})" style="cursor:pointer">
              <div class="char-avatar">${avatarLetter(ch.name)}</div>
              <div class="char-info">
                <div class="char-name">${esc(ch.name)}</div>
                <div class="char-detail">${esc(ch.race)} ${esc(ch.class)} · Lvl ${ch.level}</div>
                <div class="dash-hp-bar"><div class="dash-hp-bar-fill${hpPct(ch.hp_current, ch.hp_max) < 30 ? ' low-hp' : ''}" style="width:${hpPct(ch.hp_current, ch.hp_max)}%"></div></div>
              </div>
              <span class="fw-bold" style="font-size:0.85rem">${ch.hp_current}/${ch.hp_max}</span>
            </div>
          `).join('') || '<div class="text-muted small">No characters yet.</div>'}
        </div>
        <div class="dash-card">
          <h6>Overview</h6>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px">
            <div><div class="dash-value">${d.active_quests}</div><div class="dash-label">Active Quests</div></div>
            <div><div class="dash-value">${d.upcoming_sessions}</div><div class="dash-label">Upcoming Sessions</div></div>
            <div><div class="dash-value">${d.active_conditions}</div><div class="dash-label">Conditions</div></div>
            <div><div class="dash-value">${d.downtime_count}</div><div class="dash-label">Downtime Acts</div></div>
            <div><div class="dash-value">${d.recent_journal}</div><div class="dash-label">Journal (7d)</div></div>
            <div><div class="dash-value">${d.total_members}</div><div class="dash-label">Members</div></div>
          </div>
        </div>
        <div class="dash-card">
          <h6>Upcoming Events</h6>
          ${(d.upcoming_events || []).map((ev: any) => `
            <div class="dash-list-item">
              <span>${esc(ev.title)}</span>
              <span class="text-muted small">${ev.event_date || ''}</span>
            </div>
          `).join('') || '<div class="text-muted small">No upcoming events.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Timeline</h6>
          ${(d.recent_timeline || []).map((tl: any) => `
            <div class="dash-list-item">
              <span>${esc(tl.title)}</span>
              <span class="text-muted small">${tl.event_date || ''}</span>
            </div>
          `).join('') || '<div class="text-muted small">No timeline events.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Recaps</h6>
          ${(d.recent_recaps || []).map((r: any) => `
            <div class="dash-list-item">
              <span>${esc(r.title)}</span>
              <span class="text-muted small">${r.created_at || ''}</span>
            </div>
          `).join('') || '<div class="text-muted small">No recaps yet.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Combats</h6>
          ${(d.recent_combats || []).map((cbt: any) => `
            <div class="dash-list-item">
              <span>${esc(cbt.name)}</span>
              <span class="text-muted small">${esc(cbt.created_at)}</span>
            </div>
          `).join('') || '<div class="text-muted small">No combats yet.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Dice Rolls</h6>
          ${(d.recent_dice_rolls || []).map((dr: any) => `
            <div class="dice-roll-mini">
              <span class="roll-expr">${esc(dr.expression)}</span>
              <span class="roll-total">${dr.total}</span>
            </div>
          `).join('') || '<div class="text-muted small">No dice rolls yet.</div>'}
        </div>
      </div>
      <div class="text-center mt-3 d-flex gap-2 justify-content-center">
        <button class="btn btn-sm btn-outline-warning" onclick="showTableScreen(${campaignId})"><i class="fa-solid fa-tv me-1"></i>Table Screen</button>
        <button class="btn btn-sm btn-outline-light" onclick="showAmbienceControls(${campaignId})"><i class="fa-solid fa-music me-1"></i>Ambience</button>
        <button class="btn btn-sm btn-outline-success" onclick="showBattlemap(${campaignId})"><i class="fa-solid fa-border-all me-1"></i>Battlemap</button>
        <button class="btn btn-sm btn-outline-info" onclick="showTransferExport(undefined, ${campaignId})"><i class="fa-solid fa-download me-1"></i>Export Campaign</button>
        <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>`;
    document.getElementById('campaignDashContent')!.innerHTML = content;
  } catch (e: any) {
    document.getElementById('campaignDashContent')!.innerHTML = `<div class="empty-state"><p class="text-danger">${esc(e.message)}</p></div>`;
  }
});

// ─── Party Inventory & Treasury ───

expose('showPartyInventory', async function (campaignId: number) {
  showModal('Party Inventory', `<div id="partyInvContent"><div class="ornament">✧ Loading... ✧</div></div>`);
  try {
    const items = await api('GET', `/api/campaigns/${campaignId}/party-items`);
    const content = `
      <button class="btn btn-gold btn-sm mb-2" onclick="addPartyItem(${campaignId})"><i class="fa-solid fa-plus me-1"></i>Add Item</button>
      ${items.length ? items.map((i: any) => `
        <div class="inv-item">
          <div>
            <strong>${esc(i.name)}</strong>
            <span class="badge badge-muted ms-1">×${i.quantity}</span>
            ${i.notes ? `<div class="small text-muted">${esc(i.notes)}</div>` : ''}
          </div>
          <button class="btn btn-sm btn-outline-secondary" onclick="editPartyItem(${campaignId}, ${i.id})"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-sm btn-outline-danger" onclick="deletePartyItem(${campaignId}, ${i.id})"><i class="fa-solid fa-trash"></i></button>
        </div>
      `).join('') : '<div class="text-muted small fst-italic">No party items yet. Add some loot!</div>'}
      <div class="text-center mt-3">
        <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>`;
    document.getElementById('partyInvContent')!.innerHTML = content;
  } catch (e: any) {
    document.getElementById('partyInvContent')!.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

expose('addPartyItem', async function (campaignId: number) {
  showModal('Add Party Item', `
    <div class="mb-2"><label class="form-label">Item Name</label><input class="form-control" id="piName"></div>
    <div class="mb-2"><label class="form-label">Quantity</label><input class="form-control" id="piQty" type="number" value="1"></div>
    <div class="mb-2"><label class="form-label">Notes</label><textarea class="form-control" id="piNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="savePartyItem(${campaignId})">Add</button>
  `);
});

expose('editPartyItem', async function (campaignId: number, itemId: number) {
  const items = await api('GET', `/api/campaigns/${campaignId}/party-items`);
  const it = items.find((x: any) => x.id === itemId);
  if (!it) { toast('Item not found', true); return; }
  showModal('Edit Party Item', `
    <div class="mb-2"><label class="form-label">Item Name</label><input class="form-control" id="piName" value="${esc(it.name)}"></div>
    <div class="mb-2"><label class="form-label">Quantity</label><input class="form-control" id="piQty" type="number" value="${it.quantity}"></div>
    <div class="mb-2"><label class="form-label">Notes</label><textarea class="form-control" id="piNotes" rows="2">${esc(it.notes)}</textarea></div>
    <button class="btn btn-primary w-100" onclick="savePartyItem(${campaignId}, ${itemId})">Save</button>
  `);
});

expose('savePartyItem', async function (campaignId: number, itemId?: number) {
  const name = (document.getElementById('piName') as HTMLInputElement).value.trim();
  if (!name) { toast('Name required', true); return; }
  const body = {
    name,
    quantity: parseInt((document.getElementById('piQty') as HTMLInputElement).value) || 1,
    notes: (document.getElementById('piNotes') as HTMLTextAreaElement).value,
  };
  if (itemId) {
    await api('PUT', `/api/party-items/${itemId}`, body);
  } else {
    await api('POST', `/api/campaigns/${campaignId}/party-items`, body);
  }
  hideModal();
  toast(itemId ? 'Item updated' : 'Item added to party inventory');
  (window as any).showPartyInventory(campaignId);
});

expose('deletePartyItem', async function (campaignId: number, itemId: number) {
  if (!confirm('Remove this item?')) return;
  await api('DELETE', `/api/party-items/${itemId}`);
  toast('Item removed');
  (window as any).showPartyInventory(campaignId);
});

// ─── Session Planner ───

expose('showSessionPlanner', async function (campaignId: number) {
  showModal('Session Planner', `<div id="sessionPlanContent"><div class="ornament">✧ Loading sessions... ✧</div></div>`);
  try {
    const plans = await api('GET', `/api/campaigns/${campaignId}/session-plans`);
    const rsvpData = await Promise.all(plans.map((p: any) =>
      api('GET', `/api/session-plans/${p.id}/rsvps`).catch(() => ({ rsvps: [], counts: { yes: 0, maybe: 0, no: 0, total: 0 }, my_status: '' }))
    ));
    const statusBadge = (s: string) => {
      const cls = s === 'planned' ? 'status-badge-planned' : s === 'ready' ? 'status-badge-ready' : s === 'in-progress' ? 'status-badge-in-progress' : 'status-badge-completed';
      return `<span class="${cls}">${esc(s)}</span>`;
    };
    const rsvpControls = (p: any, data: any) => {
      const c = data?.counts || {};
      const mine = data?.my_status || '';
      const btn = (status: string, label: string, cls: string) =>
        `<button class="btn btn-sm ${mine === status ? cls : 'btn-outline-secondary'}" onclick="setSessionRsvp(${campaignId}, ${p.id}, '${status}')">${label}</button>`;
      return `
        <div class="d-flex align-items-center gap-2 mt-2 flex-wrap" id="rsvp-${p.id}">
          ${btn('yes', 'Going', 'btn-success')}
          ${btn('maybe', 'Maybe', 'btn-warning')}
          ${btn('no', "Can't", 'btn-danger')}
          <span class="small text-muted">
            <i class="fa-solid fa-check text-success"></i> ${c.yes || 0}
            <i class="fa-solid fa-question text-warning ms-2"></i> ${c.maybe || 0}
            <i class="fa-solid fa-xmark text-danger ms-2"></i> ${c.no || 0}
          </span>
          <button class="btn btn-sm btn-outline-info ms-auto" onclick="showSessionAttendance(${campaignId}, ${p.id})" title="Mark attendance"><i class="fa-solid fa-clipboard-check"></i></button>
        </div>`;
    };
    const content = `
      <button class="btn btn-gold btn-sm mb-2" onclick="showSessionPlanForm(${campaignId})"><i class="fa-solid fa-plus me-1"></i>New Session Plan</button>
      ${plans.length ? plans.map((p: any, i: number) => `
        <div class="session-plan-card">
          <div class="d-flex justify-content-between align-items-start">
            <div>
              <div class="plan-title">${esc(p.title)}</div>
              <div class="plan-meta">
                ${p.session_date ? `<span><i class="fa-regular fa-calendar me-1"></i>${esc(p.session_date)}</span>` : ''}
                ${p.expected_duration ? `<span class="ms-2"><i class="fa-regular fa-clock me-1"></i>${p.expected_duration} min</span>` : ''}
              </div>
            </div>
            <div class="d-flex gap-1 align-items-center">
              ${statusBadge(p.status)}
              <button class="btn btn-sm btn-outline-primary" onclick="showSessionPlanForm(${campaignId}, ${JSON.stringify(p).replace(/"/g, "'")})"><i class="fa-solid fa-pen"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteSessionPlan(${p.id}, ${campaignId})"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>
          ${p.dm_notes ? `<div class="small text-muted mt-1">${esc(p.dm_notes.substring(0, 200))}${p.dm_notes.length > 200 ? '...' : ''}</div>` : ''}
          ${rsvpControls(p, rsvpData[i])}
        </div>
      `).join('') : '<div class="text-muted small fst-italic">No session plans yet. Create one to get started!</div>'}
      <div class="text-center mt-3">
        <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>`;
    document.getElementById('sessionPlanContent')!.innerHTML = content;
  } catch (e: any) {
    document.getElementById('sessionPlanContent')!.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

// Set the current user's RSVP for a session plan and refresh the planner.
expose('setSessionRsvp', async function (campaignId: number, planId: number, status: string) {
  try {
    await api('PUT', `/api/session-plans/${planId}/rsvp`, { status, note: '' });
    toast(status === 'yes' ? "You're in!" : status === 'maybe' ? 'Marked as maybe' : "Marked as can't make it");
    (window as any).showSessionPlanner(campaignId);
  } catch (e: any) {
    toast(e.message || 'Failed to save RSVP', true);
  }
});

// DM view: toggle attendance for each member of a session.
expose('showSessionAttendance', async function (campaignId: number, planId: number) {
  showModal('Attendance', `<div id="attendanceContent"><div class="ornament">✧ Loading... ✧</div></div>`);
  try {
    const data = await api('GET', `/api/session-plans/${planId}/rsvps`);
    const rows = (data.rsvps || []).map((r: any) => `
      <div class="d-flex align-items-center justify-content-between border-bottom py-2">
        <div>
          <span class="fw-bold">${esc(r.username)}</span>
          ${r.status ? `<span class="badge bg-secondary ms-2">${esc(r.status)}</span>` : '<span class="badge bg-light text-dark ms-2">no reply</span>'}
        </div>
        <div class="form-check form-switch">
          <input class="form-check-input" type="checkbox" ${r.attended ? 'checked' : ''} onchange="setSessionAttendance(${planId}, ${r.user_id}, this.checked)">
        </div>
      </div>`).join('');
    document.getElementById('attendanceContent')!.innerHTML = `
      ${rows || '<div class="text-muted small fst-italic">No members.</div>'}
      <div class="text-center mt-3"><button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Done</button></div>`;
  } catch (e: any) {
    document.getElementById('attendanceContent')!.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

expose('setSessionAttendance', async function (planId: number, userId: number, attended: boolean) {
  try {
    await api('PUT', `/api/session-plans/${planId}/attendance`, { user_id: userId, attended });
  } catch (e: any) {
    toast(e.message || 'Failed to update attendance', true);
  }
});

expose('showSessionPlanForm', function (campaignId: number, plan?: any) {
  const isEdit = !!plan;
  const title = isEdit ? 'Edit Session Plan' : 'New Session Plan';
  showModal(title, `
    <div class="mb-2"><label class="form-label">Title</label><input class="form-control" id="spTitle" value="${isEdit ? esc(plan.title) : ''}"></div>
    <div class="row g-2 mb-2">
      <div class="col-6"><label class="form-label">Session Date</label><input class="form-control" id="spDate" type="date" value="${isEdit && plan.session_date ? plan.session_date : ''}"></div>
      <div class="col-6"><label class="form-label">Expected Duration (minutes)</label><input class="form-control" id="spDuration" type="number" min="0" step="15" placeholder="e.g. 180" value="${isEdit && plan.expected_duration ? Number(plan.expected_duration) : ''}"></div>
    </div>
    <div class="mb-2"><label class="form-label">Status</label>
      <select class="form-select" id="spStatus">
        <option value="planned" ${isEdit && plan.status === 'planned' ? 'selected' : ''}>Planned</option>
        <option value="ready" ${isEdit && plan.status === 'ready' ? 'selected' : ''}>Ready</option>
        <option value="in-progress" ${isEdit && plan.status === 'in-progress' ? 'selected' : ''}>In Progress</option>
        <option value="completed" ${isEdit && plan.status === 'completed' ? 'selected' : ''}>Completed</option>
      </select>
    </div>
    <div class="mb-2"><label class="form-label">DM Notes</label><textarea class="form-control" id="spNotes" rows="3">${isEdit ? esc(plan.dm_notes || '') : ''}</textarea></div>
    <div class="mb-2"><label class="form-label">Planned Encounters (one per line)</label><textarea class="form-control" id="spEncounters" rows="2" placeholder="Goblin ambush&#10;Bugbear leader">${isEdit && plan.planned_encounters ? (Array.isArray(plan.planned_encounters) ? plan.planned_encounters.join('\n') : plan.planned_encounters) : ''}</textarea></div>
    <div class="mb-2"><label class="form-label">Player Goals (one per line)</label><textarea class="form-control" id="spGoals" rows="2" placeholder="Rescue the prisoners&#10;Find the hidden passage">${isEdit && plan.player_goals ? (Array.isArray(plan.player_goals) ? plan.player_goals.join('\n') : plan.player_goals) : ''}</textarea></div>
    <button class="btn btn-primary w-100" onclick="saveSessionPlan(${campaignId}${isEdit ? `, ${plan.id}` : ''})"><i class="fa-solid fa-save me-1"></i>${isEdit ? 'Update' : 'Create'}</button>
  `);
});

expose('saveSessionPlan', async function (campaignId: number, planId?: number) {
  const title = (document.getElementById('spTitle') as HTMLInputElement).value.trim();
  if (!title) { toast('Title required', true); return; }
  const encounters = (document.getElementById('spEncounters') as HTMLTextAreaElement).value.split('\n').filter((l: string) => l.trim());
  const goals = (document.getElementById('spGoals') as HTMLTextAreaElement).value.split('\n').filter((l: string) => l.trim());
  const body = {
    title,
    session_date: (document.getElementById('spDate') as HTMLInputElement).value || '',
    status: (document.getElementById('spStatus') as HTMLSelectElement).value,
    dm_notes: (document.getElementById('spNotes') as HTMLTextAreaElement).value,
    planned_encounters: JSON.stringify(encounters),
    npc_ids: '[]',
    player_goals: JSON.stringify(goals),
    expected_duration: Number((document.getElementById('spDuration') as HTMLInputElement).value) || 0,
  };
  if (planId) {
    await api('PUT', `/api/session-plans/${planId}`, body);
  } else {
    await api('POST', `/api/campaigns/${campaignId}/session-plans`, body);
  }
  hideModal();
  toast(planId ? 'Session plan updated' : 'Session plan created');
  (window as any).showSessionPlanner(campaignId);
});

expose('deleteSessionPlan', async function (planId: number, campaignId: number) {
  if (!confirm('Delete this session plan?')) return;
  await api('DELETE', `/api/session-plans/${planId}`);
  toast('Session plan deleted');
  (window as any).showSessionPlanner(campaignId);
});

// ─── Table Screen (second display) ───
// Mints (or reuses) a read-only table share and opens the TV-friendly page.
expose('showTableScreen', async function (campaignId: number) {
  try {
    const res = await api('POST', '/api/share', { entity_type: 'table', entity_id: campaignId });
    const url = window.location.origin + '/share/' + res.token;
    window.open(url, '_blank', 'noopener');
    toast('Table screen opened in a new tab');
  } catch {
    toast('Could not open table screen', true);
  }
});

// ─── Ambience / soundboard ───
// Broadcasts a synthesized soundscape to every connected campaign screen.
expose('showAmbienceControls', function (campaignId: number) {
  const active = (document.body.dataset.ambience || '') as string;
  const buttons = listAmbienceTracks().map((t) => `
    <button class="btn ${active === t.key ? 'btn-gold' : 'btn-outline-light'} mb-2"
            onclick="ambiencePlay(${campaignId},'${t.key}')">
      <i class="fa-solid ${t.icon} me-1"></i>${t.label}
    </button>`).join('');
  showModal('Table Ambience', `
    <div class="d-flex flex-wrap gap-2">${buttons}</div>
    <hr>
    <button class="btn btn-outline-danger w-100" onclick="ambienceStop(${campaignId})">
      <i class="fa-solid fa-volume-xmark me-1"></i>Stop Ambience
    </button>
    <p class="small text-muted mt-2 mb-0">Sound is synthesized in the browser — every connected
    screen and the table display hear the same soundscape.</p>`);
});

expose('ambiencePlay', async function (campaignId: number, track: string) {
  try {
    await api('POST', `/api/campaigns/${campaignId}/ambience`, { track, action: 'play', volume: 0.5 });
    playAmbience(track, 0.5);
    toast(`${track} ambience playing`);
    (window as any).showAmbienceControls?.(campaignId);
  } catch {
    toast('Could not set ambience', true);
  }
});

expose('ambienceStop', async function (campaignId: number) {
  try {
    await api('POST', `/api/campaigns/${campaignId}/ambience`, { action: 'stop' });
    stopAmbience();
    toast('Ambience stopped');
    hideModal();
  } catch {
    toast('Could not stop ambience', true);
  }
});
