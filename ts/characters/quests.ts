import { expose } from '../lib/expose';
import { currentChar } from '../lib/state';
import { esc, capitalize, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';

async function renderQuests() {
  const el = document.getElementById('questsSection')!;
  try {
    const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
    const groups: Record<string, any[]> = { active: [], available: [], complete: [], failed: [], abandoned: [] };
    quests.forEach((q:any) => { if (groups[q.status]) groups[q.status].push(q); });
    let html = '<div class="d-flex justify-content-between align-items-center"><h5>Quests</h5><button class="btn btn-primary btn-sm" onclick="showAddQuest()"><i class="fa-solid fa-plus me-1"></i>New Quest</button></div>';
    const labels: Record<string,string> = { active: 'Active', available: 'Available', complete: 'Complete', failed: 'Failed', abandoned: 'Abandoned' };
    for (const st of ['active', 'available', 'complete', 'failed', 'abandoned']) {
      const qs = groups[st] || [];
      if (!qs.length) continue;
      html += `<h6 class="mt-3 text-muted">${labels[st]}</h6>`;
      for (const q of qs) {
        const opts = ['active','available','complete','failed','abandoned'].map(s => `<option value="${s}"${s===q.status?' selected':''}>${capitalize(s)}</option>`).join('');
        html += `<div class="card mb-2">
          <div class="card-body py-2 px-3">
            <div class="d-flex justify-content-between align-items-start">
              <div><span class="fw-bold">${esc(q.name)}</span></div>
              <div class="d-flex gap-1">
                <select class="form-select form-select-sm" style="width:auto" onchange="updateQuestStatus(${q.id},this.value)">${opts}</select>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteQuest(${q.id})"><i class="fa-solid fa-trash"></i></button>
              </div>
            </div>
            <p class="mb-0 mt-1 small text-muted">${esc(q.description).substring(0, 200)}</p>
            ${q.objectives ? `<div class="mt-1 small text-muted"><strong>Objectives:</strong> ${esc(q.objectives).substring(0, 150)}</div>` : ''}
            ${q.rewards ? `<div class="mt-1 small text-success"><strong>Reward:</strong> ${esc(q.rewards).substring(0, 150)}</div>` : ''}
          </div>
        </div>`;
      }
    }
    if (quests.length === 0) html += '<div class="empty-state"><i class="fa-solid fa-scroll fa-2x mb-2 d-block text-muted"></i>No quests yet.</div>';
    el.innerHTML = html;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load quests. Try again later.</p></div>'; }
}
expose('renderQuests', renderQuests);

expose('showAddQuest', function () {
  showModal('New Quest', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="questName" placeholder="e.g. Retrieve the Lost Artifact"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="questDesc" rows="3"></textarea></div>
    <div class="mb-3"><label class="form-label">Objectives</label><textarea class="form-control" id="questObj" rows="2" placeholder="1. Travel to the Temple\n2. Defeat the guardian\n3. Retrieve the artifact"></textarea></div>
    <div class="mb-3"><label class="form-label">Rewards</label><textarea class="form-control" id="questRewards" rows="2" placeholder="500 XP, +1 Longsword, 200 GP"></textarea></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="questNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveQuest()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
});

expose('saveQuest', async function () {
  await api('POST', `/api/characters/${currentChar.id}/quests`, {
    name: (document.getElementById('questName') as HTMLInputElement).value,
    description: (document.getElementById('questDesc') as HTMLTextAreaElement).value,
    objectives: (document.getElementById('questObj') as HTMLTextAreaElement).value,
    rewards: (document.getElementById('questRewards') as HTMLTextAreaElement).value,
    notes: (document.getElementById('questNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderQuests();
  toast('Quest created');
});

expose('updateQuestStatus', async function (id:number, status:string) {
  const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
  const q = quests.find((x:any) => x.id === id);
  if (!q) return;
  q.status = status;
  await api('PUT', `/api/quests/${id}`, q);
  renderQuests();
  toast('Quest status updated');
});

expose('deleteQuest', async function (id:number) {
  if (!confirm('Delete this quest?')) return;
  await api('DELETE', `/api/quests/${id}`);
  renderQuests();
  toast('Quest deleted');
});
