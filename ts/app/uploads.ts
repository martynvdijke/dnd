// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { showModal, hideModal, toast } from '../lib/dom';
import { getCsrfToken, getApiToken } from '../lib/api';

// ─── Polymorphic File Uploads ───

expose('showUploadModal', function (ownerType: string, ownerId: number) {
  showModal('Upload File', `
    <div class="mb-3">
      <label class="form-label">Select file</label>
      <input type="file" class="form-control" id="uploadFileInput">
    </div>
    <button class="btn btn-primary w-100" onclick="doUpload('${ownerType}', ${ownerId})"><i class="fa-solid fa-upload me-1"></i>Upload</button>
  `);
});

expose('doUpload', async function (ownerType: string, ownerId: number) {
  const input = document.getElementById('uploadFileInput') as HTMLInputElement;
  if (!input?.files?.length) { toast('Select a file', true); return; }
  const form = new FormData();
  form.append('file', input.files[0]);
  form.append('owner_type', ownerType);
  form.append('owner_id', String(ownerId));
  try {
    const res = await fetch('/api/upload', { method: 'POST', body: form,
      headers: { 'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || getCsrfToken(), ...(getApiToken() ? { 'Authorization': `Bearer ${getApiToken()}` } : {}) }
    });
    if (!res.ok) throw new Error((await res.json()).error || 'Upload failed');
    hideModal();
    toast('File uploaded');
  } catch (e: any) { toast(e.message, true); }
});
