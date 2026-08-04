import { toggleTheme, initTheme as sharedInitTheme } from './lib/theme';
import { expose } from './lib/expose';

function updateLoginThemeUI() {
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = isDark ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
  const label = document.getElementById('themeLabel');
  if (label) label.textContent = isDark ? 'Light Mode' : 'Dark Mode';
}

function initTheme() {
  sharedInitTheme();
  updateLoginThemeUI();
}

expose('toggleTheme', () => {
  toggleTheme();
  updateLoginThemeUI();
});

async function init() {
  initTheme();

  const form = document.getElementById('loginForm') as HTMLFormElement;
  const errorDiv = document.getElementById('error') as HTMLDivElement;

  // Attach handler immediately (before async checks) to avoid race conditions
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorDiv.classList.add('d-none');

    const submitBtn = form.querySelector('button[type="submit"]') as HTMLButtonElement;
    const origHtml = submitBtn.innerHTML;
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>Signing in...';

    const username = (document.getElementById('username') as HTMLInputElement).value;
    const password = (document.getElementById('password') as HTMLInputElement).value;

    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
        credentials: 'include',
      });

      if (res.ok) {
        window.location.href = '/';
      } else {
        const err = await res.json();
        errorDiv.textContent = err.error || 'Invalid credentials';
        errorDiv.classList.remove('d-none');
      }
    } finally {
      submitBtn.disabled = false;
      submitBtn.innerHTML = origHtml;
    }
  });

  const res = await fetch('/api/check-setup');
  const data = await res.json();
  if (!data.setup) {
    window.location.href = '/setup';
    return;
  }

  const res2 = await fetch('/api/user/me', { credentials: 'include' });
  if (res2.ok) {
    window.location.href = '/';
    return;
  }
}

init();
