// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, attrEscape, toast, showModal, hideModal } from '../lib/dom';
import { api } from './state';
import { renderError } from '../lib/errors';
async function loadUsers() {
  try {
    const users = await api('GET', '/api/admin/users');
    const tbody = document.querySelector('#userTable tbody')!;
    tbody.innerHTML = users.map((u: any) => `
      <tr>
        <td>${u.id}</td>
        <td>${esc(u.username)}</td>
        <td>${esc(u.display_name)}</td>
        <td>${esc(u.email)}</td>
        <td><span class="badge ${u.role === 'admin' ? 'badge-blood' : 'badge-gold'}">${u.role}</span></td>
        <td>${u.created_at}</td>
        <td>
          <button class="btn btn-outline-primary btn-sm js-edit-user" data-id="${u.id}" data-username="${attrEscape(u.username)}" data-display="${attrEscape(u.display_name)}" data-email="${attrEscape(u.email)}" data-role="${attrEscape(u.role)}"><i class="fa-solid fa-pen"></i></button>
          <button class="btn btn-outline-danger btn-sm" onclick="deleteUser(${u.id})"><i class="fa-solid fa-trash"></i></button>
          <button class="btn btn-outline-secondary btn-sm" onclick="resetPass(${u.id})"><i class="fa-solid fa-key"></i></button>
        </td>
      </tr>
    `).join('');
    tbody.querySelectorAll<HTMLButtonElement>('.js-edit-user').forEach(btn => {
      btn.addEventListener('click', () => (window as any).editUser(Number(btn.dataset.id), btn.dataset.username || '', btn.dataset.display || '', btn.dataset.email || '', btn.dataset.role || ''));
    });
  } catch (e: any) {
    renderError(e);
  }
}
expose('loadUsers', loadUsers);
expose('showAddUser', function () {
  showModal('Add User', `
    <div class="mb-3"><label class="form-label">Username</label><input class="form-control" id="addUsername"></div>
    <div class="mb-3"><label class="form-label">Password</label><input class="form-control" type="password" id="addPassword"></div>
    <div class="mb-3"><label class="form-label">Display Name</label><input class="form-control" id="addDisplay"></div>
    <div class="mb-3"><label class="form-label">Email</label><input class="form-control" type="email" id="addEmail"></div>
    <div class="mb-3">
      <label class="form-label">Role</label>
      <select class="form-select" id="addRole"><option value="user">User</option><option value="dm">DM</option><option value="admin">Admin</option></select>
    </div>
    <button class="btn btn-primary w-100" onclick="saveNewUser()">Create</button>
  `);
});
expose('saveNewUser', async function () {
  try {
    await api('POST', '/api/admin/users', {
      username: (document.getElementById('addUsername') as HTMLInputElement).value,
      password: (document.getElementById('addPassword') as HTMLInputElement).value,
      display_name: (document.getElementById('addDisplay') as HTMLInputElement).value,
      email: (document.getElementById('addEmail') as HTMLInputElement).value,
      role: (document.getElementById('addRole') as HTMLSelectElement).value,
    });
    hideModal();
    loadUsers();
    toast('User created');
  } catch (e: any) {
    renderError(e);
  }
});
expose('editUser', function (id: number, username: string, display: string, email: string, role: string) {
  showModal('Edit User', `
    <div class="mb-3"><label class="form-label">Username</label><input class="form-control" id="editUsername" value="${esc(username)}"></div>
    <div class="mb-3"><label class="form-label">Display Name</label><input class="form-control" id="editDisplay" value="${esc(display)}"></div>
    <div class="mb-3"><label class="form-label">Email</label><input class="form-control" type="email" id="editEmail" value="${esc(email)}"></div>
    <div class="mb-3">
      <label class="form-label">Role</label>
      <select class="form-select" id="editRole"><option value="user" ${role === 'user' ? 'selected' : ''}>User</option><option value="dm" ${role === 'dm' ? 'selected' : ''}>DM</option><option value="admin" ${role === 'admin' ? 'selected' : ''}>Admin</option></select>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEditUser(${id})">Save</button>
  `);
});
expose('saveEditUser', async function (id: number) {
  try {
    await api('PUT', `/api/admin/users/${id}`, {
      username: (document.getElementById('editUsername') as HTMLInputElement).value,
      display_name: (document.getElementById('editDisplay') as HTMLInputElement).value,
      email: (document.getElementById('editEmail') as HTMLInputElement).value,
      role: (document.getElementById('editRole') as HTMLSelectElement).value,
    });
    hideModal();
    loadUsers();
    toast('User updated');
  } catch (e: any) {
    renderError(e);
  }
});
expose('deleteUser', async function (id: number) {
  if (!confirm('Delete this user?')) return;
  try {
    await api('DELETE', `/api/admin/users/${id}`);
    loadUsers();
    toast('User deleted');
  } catch (e: any) {
    renderError(e);
  }
});
expose('resetPass', function (id: number) {
  showModal('Reset Password', `
    <div class="mb-3"><label class="form-label">New Password</label><input class="form-control" type="password" id="resetPass"></div>
    <button class="btn btn-primary w-100" onclick="doResetPass(${id})">Reset</button>
  `);
});
expose('doResetPass', async function (id: number) {
  try {
    await api('PUT', `/api/admin/users/${id}/password`, {
      password: (document.getElementById('resetPass') as HTMLInputElement).value,
    });
    hideModal();
    toast('Password reset');
  } catch (e: any) {
    renderError(e);
  }
});
