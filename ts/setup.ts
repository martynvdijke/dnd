import { toggleTheme, initTheme as sharedInitTheme } from './lib/theme';
import { expose } from './lib/expose';
import { fetchAuthStatus, passwordLoginOffered } from './lib/auth-status';

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

expose('toggleTheme', () => {
  toggleTheme();
  updateSetupThemeUI();
});

/** Show or hide an element by id using the page's `d-none` idiom. */
function setVisible(id: string, visible: boolean) {
  document.getElementById(id)?.classList.toggle('d-none', !visible);
}

async function submitSetup(e: Event) {
  e.preventDefault();

  const form = e.currentTarget as HTMLFormElement;
  const errorDiv = document.getElementById('error') as HTMLDivElement | null;
  errorDiv?.classList.add('d-none');

  const submitBtn = form.querySelector('button[type="submit"]') as HTMLButtonElement;
  const origHtml = submitBtn.innerHTML;
  submitBtn.disabled = true;
  submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>Setting up...';

  const username = (document.getElementById('username') as HTMLInputElement).value;
  const password = (document.getElementById('password') as HTMLInputElement).value;
  const confirm = (document.getElementById('confirm') as HTMLInputElement).value;

  if (password !== confirm) {
    if (errorDiv) {
      errorDiv.textContent = 'Passwords do not match';
      errorDiv.classList.remove('d-none');
    }
    submitBtn.disabled = false;
    submitBtn.innerHTML = origHtml;
    return;
  }

  if (password.length < 8) {
    if (errorDiv) {
      errorDiv.textContent = 'Password must be at least 8 characters';
      errorDiv.classList.remove('d-none');
    }
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

    if (import.meta.env.DEV) console.log('Setup POST response:', res.status, res.statusText);
    if (res.ok) {
      window.location.href = '/';
    } else {
      const text = await res.text();
      if (import.meta.env.DEV) console.error('Setup POST failed:', res.status, text);
      if (errorDiv) {
        errorDiv.textContent = text || 'Setup failed';
        errorDiv.classList.remove('d-none');
      }
    }
  } finally {
    submitBtn.disabled = false;
    submitBtn.innerHTML = origHtml;
  }
}

async function init() {
  initTheme();

  // Already set up: the login page is the right place.
  const res = await fetch('/api/check-setup');
  const data = await res.json();
  if (data.setup) {
    window.location.href = '/login';
    return;
  }

  // The first admin is created with a password, so this form only works while
  // password login is allowed. With the break-glass flag off the POST would be
  // answered with 403, so the page explains the way in instead of offering a
  // form that cannot succeed.
  if (!passwordLoginOffered(await fetchAuthStatus())) {
    setVisible('setupForm', false);
    setVisible('setupUnavailable', true);
    return;
  }

  document.getElementById('setupForm')?.addEventListener('submit', submitSetup);
}

/** Resolves once the page has applied the login policy (awaited by tests). */
export const setupReady: Promise<void> = init();
