/**
 * Downtime Activities — render + CRUD UI for a character's between-session
 * activities. Backend: handlers/downtime.go (full CRUD + advance + types).
 */
import { expose } from '../lib/expose';
import { currentChar } from '../lib/state';
import { esc, toast, showModal, hideModal } from '../lib/dom';
import { api } from '../lib/api';

const DOWNTIME_TYPES = ['training', 'crafting', 'research', 'carousing', 'pit_fighting', 'crime', 'religious', 'scribing', 'gambling', 'other'];

const statusBadge = (s: string): string =>
  s === 'complete' ? 'bg-success' : s === 'failed' ? 'bg-danger' : 'bg-info';

export async function renderDowntime(): Promise<void> {
  const el = document.getElementById('downtimeSection');
  if (!el || !currentChar) return;
  try {
    const acts = (await api('GET', `/api/characters/${currentChar.id}/downtime`)) as any[];
    let html = `<div class="d-flex justify-content-between align-items-center mb-2"><h5 class="mb-0">Downtime</h5>
      <button class="btn btn-primary btn-sm" onclick="openDowntimeForm()"><i class="fa-solid fa-plus me-1"></i>New Activity</button></div>`;
    if (!acts.length) {
      html += `<div class="empty-state"><i class="fa-solid fa-hourglass-half fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Downtime Activities</p><p class="small text-muted">Plan training, research or carousing between sessions.</p></div>`;
    } else {
      for (const a of acts) {
        const progress = a.days_required > 0 ? `${a.days_completed}/${a.days_required}d` : `${a.days_completed}d`;
        html += `<div class="card mb-2"><div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between align-items-start flex-wrap gap-2">
            <div>
              <span class="fw-bold">${esc(a.name)}</span>
              <span class="badge badge-muted ms-1">${esc(a.activity_type)}</span>
              <span class="badge ${statusBadge(a.status)} ms-1">${esc(a.status)}</span>
              <div class="small text-muted mt-1">DC ${a.dc} · ${progress}${a.total_cost ? ` · ${a.total_cost} gp` : ''}${a.reward ? ` · reward: ${esc(a.reward)}` : ''}</div>
              ${a.description ? `<div class="small text-muted">${esc(a.description)}</div>` : ''}
            </div>
            <div class="d-flex gap-1">
              ${a.status === 'in-progress' ? `<button class="btn btn-sm btn-outline-primary" onclick="advanceDowntime(${a.id})" title="Advance one day"><i class="fa-solid fa-forward"></i></button>` : ''}
              <button class="btn btn-sm btn-outline-secondary" onclick="openDowntimeForm(${a.id})" title="Edit"><i class="fa-solid fa-pen"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteDowntime(${a.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>
        </div></div>`;
      }
    }
    el.innerHTML = html;
  } catch (e: any) {
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(e.message)}</p></div>`;
  }
}
expose('renderDowntime', renderDowntime);

expose('openDowntimeForm', async function (id?: number) {
  if (!currentChar) return;
  let a: any = {};
  if (id) {
    try {
      const acts = (await api('GET', `/api/characters/${currentChar.id}/downtime`)) as any[];
      a = acts.find((x) => x.id === id) || {};
    } catch { /* fall back to blank form */ }
  }
  const typeOpts = DOWNTIME_TYPES.map((t) => `<option value="${t}" ${a.activity_type === t ? 'selected' : ''}>${t.replace('_', ' ')}</option>`).join('');
  const statusOpts = ['in-progress', 'complete', 'failed'].map((s) => `<option value="${s}" ${a.status === s ? 'selected' : ''}>${s}</option>`).join('');
  showModal(id ? 'Edit Downtime Activity' : 'New Downtime Activity', `
    <div class="row g-2 mb-2">
      <div class="col-6"><label class="form-label small">Type</label><select id="dtType" class="form-select form-select-sm">${typeOpts}</select></div>
      <div class="col-6"><label class="form-label small">Name</label><input id="dtName" class="form-control form-control-sm" value="${a.name ? esc(a.name) : ''}"></div>
    </div>
    <div class="mb-2"><label class="form-label small">Description</label><textarea id="dtDesc" class="form-control form-control-sm" rows="2">${a.description ? esc(a.description) : ''}</textarea></div>
    <div class="row g-2 mb-2">
      <div class="col-3"><label class="form-label small">DC</label><input id="dtDC" type="number" class="form-control form-control-sm" value="${a.dc ?? 10}"></div>
      <div class="col-3"><label class="form-label small">Days req.</label><input id="dtDays" type="number" class="form-control form-control-sm" value="${a.days_required ?? 1}"></div>
      <div class="col-3"><label class="form-label small">Cost/day</label><input id="dtCost" type="number" step="0.1" class="form-control form-control-sm" value="${a.cost_per_day ?? 0}"></div>
      <div class="col-3"><label class="form-label small">Status</label><select id="dtStatus" class="form-select form-select-sm">${statusOpts}</select></div>
    </div>
    <div class="row g-2 mb-2">
      <div class="col-6"><label class="form-label small">Days completed</label><input id="dtDone" type="number" class="form-control form-control-sm" value="${a.days_completed ?? 0}"></div>
      <div class="col-6"><label class="form-label small">Reward</label><input id="dtReward" class="form-control form-control-sm" value="${a.reward ? esc(a.reward) : ''}"></div>
    </div>
    <div class="mb-2"><label class="form-label small">Notes</label><textarea id="dtNotes" class="form-control form-control-sm" rows="2">${a.notes ? esc(a.notes) : ''}</textarea></div>
    <button class="btn btn-primary w-100" onclick="saveDowntime(${id || ''})">${id ? 'Save' : 'Add'}</button>`);
  document.getElementById('genericModal')?.classList.add('show');
});

expose('saveDowntime', async function (id?: number) {
  if (!currentChar) return;
  const val = (el: string) => (document.getElementById(el) as HTMLInputElement)?.value ?? '';
  const body = {
    activity_type: val('dtType'), name: val('dtName'), description: val('dtDesc'),
    dc: +val('dtDC') || 0, days_required: +val('dtDays') || 1, days_completed: +val('dtDone') || 0,
    cost_per_day: +val('dtCost') || 0, reward: val('dtReward'), status: val('dtStatus'), notes: val('dtNotes'),
  };
  try {
    if (id) await api('PUT', `/api/downtime/${id}`, body);
    else await api('POST', `/api/characters/${currentChar.id}/downtime`, body);
    hideModal();
    toast(id ? 'Activity saved' : 'Activity added');
    renderDowntime();
  } catch (e: any) { toast(e.message, true); }
});

expose('advanceDowntime', async function (id: number) {
  try {
    const r = await api('POST', `/api/downtime/${id}/advance`, {});
    toast(`Day advanced — check ${r.skill_check} vs DC (${r.success ? 'success' : 'failure'})`, !r.success);
    renderDowntime();
  } catch (e: any) { toast(e.message, true); }
});

expose('deleteDowntime', async function (id: number) {
  if (!confirm('Delete this downtime activity?')) return;
  try {
    await api('DELETE', `/api/downtime/${id}`);
    toast('Activity deleted');
    renderDowntime();
  } catch (e: any) { toast(e.message, true); }
});
