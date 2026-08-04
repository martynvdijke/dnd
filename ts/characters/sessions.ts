import { expose } from '../lib/expose';
import { currentChar } from '../lib/state';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';

async function renderSessions() {
  const el = document.getElementById('sessionsSection')!;
  try {
    const sessions = await api('GET', `/api/characters/${currentChar.id}/sessions`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Session Log</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddSession()"><i class="fa-solid fa-plus me-1"></i>Log Session</button>
      </div>
      <div class="mt-3">
        ${sessions.map((s:any) => `
          <div class="card mb-2">
            <div class="card-body py-2 px-3">
              <div class="d-flex justify-content-between align-items-start">
                <div><span class="fw-bold">${esc(s.title) || 'Session'}</span>
                  <span class="badge badge-gold ms-2">${s.session_date}</span>
                  ${s.xp_earned > 0 ? `<span class="badge badge-blood ms-1">+${s.xp_earned} XP</span>` : ''}
                  ${s.gold_earned > 0 ? `<span class="badge badge-gold ms-1">+${s.gold_earned} GP</span>` : ''}</div>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteSession(${s.id})"><i class="fa-solid fa-trash"></i></button>
              </div>
              <p class="mb-0 mt-1 small text-muted">${esc(s.notes).substring(0, 200)}</p>
              ${s.important_events ? `<p class="mb-0 mt-1 small fst-italic text-muted">${esc(s.important_events).substring(0, 150)}</p>` : ''}
            </div>
          </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-calendar fa-2x mb-2 d-block text-muted"></i>No sessions logged yet.</div>'}
      </div>`;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load sessions. Try again later.</p></div>'; }
}
expose('renderSessions', renderSessions);

expose('showAddSession', function () {
  showModal('Log Session', `
    <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="sessDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="sessTitle" placeholder="Session 1: The Adventure Begins"></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="sessNotes" rows="3" placeholder="What happened?"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">XP Earned</label><input class="form-control" id="sessXP" type="number" value="0"></div>
      <div class="col-6"><label class="form-label">Gold Earned</label><input class="form-control" id="sessGold" type="number" value="0"></div>
    </div>
    <div class="mb-3"><label class="form-label">Important Events</label><textarea class="form-control" id="sessEvents" rows="2" placeholder="Key moments, NPCs met, revelations..."></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveSession()"><i class="fa-solid fa-save me-1"></i>Log Session</button>
  `);
});

expose('saveSession', async function () {
  await api('POST', `/api/characters/${currentChar.id}/sessions`, {
    session_date: (document.getElementById('sessDate') as HTMLInputElement).value,
    title: (document.getElementById('sessTitle') as HTMLInputElement).value,
    notes: (document.getElementById('sessNotes') as HTMLTextAreaElement).value,
    xp_earned: +(document.getElementById('sessXP') as HTMLInputElement).value || 0,
    gold_earned: +(document.getElementById('sessGold') as HTMLInputElement).value || 0,
    important_events: (document.getElementById('sessEvents') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderSessions();
  toast('Session logged');
});

expose('deleteSession', async function (id:number) {
  if (!confirm('Delete this session?')) return;
  await api('DELETE', `/api/sessions/${id}`);
  renderSessions();
  toast('Session deleted');
});
