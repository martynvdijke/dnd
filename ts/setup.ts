import { toggleTheme, initTheme as sharedInitTheme } from './lib/theme';

function updateSetupThemeUI() {
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = isDark ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
  const label = document.getElementById('themeLabel');
  if (label) label.textContent = isDark ? 'Light Mode' : 'Dark Mode';
}

function initTheme() {
  sharedInitTheme();
  updateSetupThemeUI();
}

(window as any).toggleTheme = () => {
  toggleTheme();
  updateSetupThemeUI();
};

async function init() {
  initTheme();

  const form = document.getElementById('setupForm') as HTMLFormElement;
  const errorDiv = document.getElementById('error') as HTMLDivElement;

  // Attach handler immediately (before async check) to avoid race condition
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorDiv.classList.add('d-none');

    const submitBtn = form.querySelector('button[type="submit"]') as HTMLButtonElement;
    const origHtml = submitBtn.innerHTML;
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>Setting up...';

    const username = (document.getElementById('username') as HTMLInputElement).value;
    const password = (document.getElementById('password') as HTMLInputElement).value;
    const confirm = (document.getElementById('confirm') as HTMLInputElement).value;

    if (password !== confirm) {
      errorDiv.textContent = 'Passwords do not match';
      errorDiv.classList.remove('d-none');
      submitBtn.disabled = false;
      submitBtn.innerHTML = origHtml;
      return;
    }

    if (password.length < 8) {
      errorDiv.textContent = 'Password must be at least 8 characters';
      errorDiv.classList.remove('d-none');
      submitBtn.disabled = false;
      submitBtn.innerHTML = origHtml;
      return;
    }

    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password, setup: true }),
        credentials: 'include',
      });

      console.log('Setup POST response:', res.status, res.statusText);
      if (res.ok) {
        window.location.href = '/';
      } else {
        const text = await res.text();
        console.error('Setup POST failed:', res.status, text);
        errorDiv.textContent = text || 'Setup failed';
        errorDiv.classList.remove('d-none');
      }
    } finally {
      submitBtn.disabled = false;
      submitBtn.innerHTML = origHtml;
    }
  });

  // Then check if setup already done (redirect if so)
  const res = await fetch('/api/check-setup');
  const data = await res.json();
  if (data.setup) {
    window.location.href = '/login';
    return;
  }
}

init();
