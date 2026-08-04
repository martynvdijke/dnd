import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import { expose } from '../lib/expose';
import { currentChar } from '../lib/state';
import { esc, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';

let journalEditor: Editor | null = null;

function destroyJournalEditor() {
  if (journalEditor) { journalEditor.destroy(); journalEditor = null; }
}

function initJournalEditor(content?: string) {
  setTimeout(() => {
    const el = document.getElementById('journalEditor');
    if (!el) return;
    journalEditor = new Editor({
      element: el,
      extensions: [
        StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
        Placeholder.configure({ placeholder: 'Write your character\'s thoughts...' }),
      ],
      content: content || '<p></p>',
    });
    const toolbar = document.getElementById('journalToolbar');
    if (toolbar) {
      const btns = [
        { icon: 'fa-bold', action: () => journalEditor?.chain().focus().toggleBold().run(), test: () => journalEditor?.isActive('bold') },
        { icon: 'fa-italic', action: () => journalEditor?.chain().focus().toggleItalic().run(), test: () => journalEditor?.isActive('italic') },
        { icon: 'fa-heading', action: () => journalEditor?.chain().focus().toggleHeading({ level: 2 }).run(), test: () => journalEditor?.isActive('heading', { level: 2 }) },
        { icon: 'fa-list-ul', action: () => journalEditor?.chain().focus().toggleBulletList().run(), test: () => journalEditor?.isActive('bulletList') },
        { icon: 'fa-list-ol', action: () => journalEditor?.chain().focus().toggleOrderedList().run(), test: () => journalEditor?.isActive('orderedList') },
        { icon: 'fa-quote-right', action: () => journalEditor?.chain().focus().toggleBlockquote().run(), test: () => journalEditor?.isActive('blockquote') },
      ];
      btns.forEach(b => {
        const btn = document.createElement('button');
        btn.type = 'button'; btn.className = 'editor-btn';
        btn.innerHTML = `<i class="fa-solid ${b.icon}"></i>`;
        btn.onclick = (e: MouseEvent) => { e.preventDefault(); b.action(); };
        toolbar.appendChild(btn);
      });
      journalEditor.on('selectionUpdate', () => {
        toolbar.querySelectorAll('.editor-btn').forEach((el: Element, i: number) => {
          el.classList.toggle('active', btns[i]?.test() || false);
        });
      });
    }
    const modal = document.getElementById('genericModal');
    modal?.addEventListener('hidden.bs.modal', destroyJournalEditor, { once: true });
  }, 50);
}

async function renderJournal() {
  const el = document.getElementById('journalSection')!;
  try {
    const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
    const months = ['January','February','March','April','May','June','July','August','September','October','November','December'];
    const groups: Record<string, any[]> = {};
    entries.forEach((j: any) => {
      const d = new Date(j.entry_date + 'T00:00:00');
      const key = months[d.getMonth()] + ' ' + d.getFullYear();
      if (!groups[key]) groups[key] = [];
      groups[key].push(j);
    });
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center flex-wrap gap-2 mb-3">
        <h5 class="mb-0"><i class="fa-solid fa-book-journal-whills me-2"></i>Character Journal</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddJournal()"><i class="fa-solid fa-plus me-1"></i>Write Entry</button>
      </div>
      <div class="journal-timeline">
        ${Object.keys(groups).length ? Object.entries(groups).map(([month, monthEntries]) => `
          <div class="journal-month-group">
            <div class="journal-month-header">${month} <small class="text-muted">(${(monthEntries as any[]).length} entries)</small></div>
            ${(monthEntries as any[]).reverse().map((j: any) => `
              <div class="journal-entry-card">
                <div class="journal-entry-header" onclick="this.closest('.journal-entry-card').classList.toggle('expanded')">
                  <div class="d-flex justify-content-between align-items-start w-100">
                    <div class="min-w-0">
                      <span class="fw-bold">${esc(j.title) || 'Untitled'}</span>
                      <span class="badge badge-gold ms-2">${j.entry_date}</span>
                    </div>
                    <div class="d-flex gap-1 flex-shrink-0" onclick="event.stopPropagation()">
                      <button class="btn btn-sm btn-outline-primary" onclick="showEditJournal(${j.id})"><i class="fa-solid fa-pen"></i></button>
                      <button class="btn btn-sm btn-outline-danger" onclick="deleteJournal(${j.id})"><i class="fa-solid fa-trash"></i></button>
                    </div>
                  </div>
                  <i class="fa-solid fa-chevron-down journal-expand-icon"></i>
                </div>
                <div class="journal-entry-body">${j.entry}</div>
              </div>
            `).join('')}
          </div>
        `).join('') : '<div class="empty-state"><i class="fa-solid fa-book-open fa-2x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Journal</p><p class="small text-muted">Record your character\'s thoughts and experiences.</p></div>'}
      </div>`;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load journal. Try again later.</p></div>'; }
}
expose('renderJournal', renderJournal);

expose('showAddJournal', function () {
  showModal('Journal Entry', `
    <div class="journal-editor-modal">
      <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="journalDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
      <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="journalTitle" placeholder="Day 1: Arrival in Waterdeep"></div>
      <div class="mb-3"><label class="form-label">Entry</label><div class="editor-toolbar" id="journalToolbar"></div><div id="journalEditor" class="journal-editor"></div></div>
      <button class="btn btn-primary w-100" onclick="saveJournal()"><i class="fa-solid fa-save me-1"></i>Save</button>
    </div>
  `);
  initJournalEditor();
});

expose('showEditJournal', async function (id: number) {
  const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
  const j = entries.find((e: any) => e.id === id);
  if (!j) return;
  showModal('Edit Journal Entry', `
    <div class="journal-editor-modal">
      <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="journalDate" type="date" value="${esc(j.entry_date)}"></div>
      <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="journalTitle" value="${esc(j.title)}"></div>
      <div class="mb-3"><label class="form-label">Entry</label><div class="editor-toolbar" id="journalToolbar"></div><div id="journalEditor" class="journal-editor"></div></div>
      <button class="btn btn-primary w-100" onclick="saveJournal(${id})"><i class="fa-solid fa-save me-1"></i>Update</button>
    </div>
  `);
  initJournalEditor(j.entry);
});

expose('saveJournal', async function (editId?: number) {
  const entry = journalEditor?.getHTML() || '';
  const title = (document.getElementById('journalTitle') as HTMLInputElement)?.value || '';
  const entry_date = (document.getElementById('journalDate') as HTMLInputElement)?.value || new Date().toISOString().split('T')[0];
  if (editId) {
    await api('PUT', `/api/journal/${editId}`, { entry_date, title, entry });
  } else {
    await api('POST', `/api/characters/${currentChar.id}/journal`, { entry_date, title, entry });
  }
  destroyJournalEditor();
  hideModal();
  renderJournal();
  toast(editId ? 'Journal entry updated' : 'Journal entry saved');
});

expose('deleteJournal', async function (id: number) {
  if (!confirm('Delete this journal entry?')) return;
  await api('DELETE', `/api/journal/${id}`);
  renderJournal();
  toast('Journal entry deleted');
});
