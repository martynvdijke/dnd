import { toggleTheme, initTheme as sharedInitTheme } from './lib/theme';
import { expose } from './lib/expose';
import { fetchAuthStatus, passwordLoginOffered, autheliaOffered, type AuthStatus } from './lib/auth-status';

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

export function oidcErrorText(code: string, passwordOffered = true): string {
  switch (code) {
    case 'oidc_expired': return 'SSO session expired — please try again.';
    case 'oidc_state': return 'SSO verification failed — please try again.';
    case 'oidc_email': return 'SSO login requires a verified email.';
    default:
      // Never point someone at a password form this deployment does not have.
      return passwordOffered
        ? 'SSO login failed — try password login or contact an admin.'
        : 'SSO login failed — try again, or ask an admin to check Authelia.';
  }
}

// Only allow same-origin relative redirects (guards open-redirect).
export function safeNext(): string {
  const n = new URLSearchParams(window.location.search).get('next') || '';
  return n.startsWith('/') && !n.startsWith('//') ? n : '/';
}

/** Show or hide an element by id using the page's `d-none` idiom. */
function setVisible(id: string, visible: boolean) {
  document.getElementById(id)?.classList.toggle('d-none', !visible);
}

/**
 * Renders this deployment's login policy on the existing page and reports
 * whether a password form is on offer.
 *
 * One code path, no separate "password disabled" page: Authelia is the default,
 * so the password form and its "Forgot password?" link disappear unless the
 * break-glass flag is set. When neither option is available the page says why
 * and names the flag to set, instead of leaving a form that always fails.
 */
export function applyAuthPolicy(status: AuthStatus | null): boolean {
  const password = passwordLoginOffered(status);
  const authelia = autheliaOffered(status);

  setVisible('loginForm', password);
  setVisible('forgotBlock', password);
  setVisible('oidcLogin', authelia);
  // The "or" rule only separates two options; offered alone it dangles.
  setVisible('oidcDivider', password && authelia);
  setVisible('oidcOnlyNote', authelia && !password);
  setVisible('loginUnavailable', !password && !authelia);

  return password;
}

/**
 * Applies the login policy and the SSO error from the redirect, and reports
 * whether password login is available (so callers only wire the form when it
 * can actually post).
 */
export async function initOIDC(): Promise<boolean> {
  const status = await fetchAuthStatus();
  const passwordOffered = applyAuthPolicy(status);

  const errorDiv = document.getElementById('error') as HTMLDivElement | null;
  const code = new URLSearchParams(window.location.search).get('error');
  if (code && errorDiv) {
    errorDiv.textContent = oidcErrorText(code, passwordOffered);
    errorDiv.classList.remove('d-none');
  }

  return passwordOffered;
}

async function submitLogin(e: Event) {
  e.preventDefault();

  const form = e.currentTarget as HTMLFormElement;
  const errorDiv = document.getElementById('error') as HTMLDivElement | null;
  errorDiv?.classList.add('d-none');

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
      window.location.href = safeNext();
    } else {
      const err = await res.json();
      if (errorDiv) {
        errorDiv.textContent = err.error || 'Invalid credentials';
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

  const res = await fetch('/api/check-setup');
  const data = await res.json();
  if (!data.setup) {
    window.location.href = '/setup';
    return;
  }

  // Ask which login options exist before wiring anything. With password login
  // retired there is no submit handler at all, so the page cannot post a
  // password the server would answer with 403.
  const passwordOffered = await initOIDC();
  if (passwordOffered) {
    document.getElementById('loginForm')?.addEventListener('submit', submitLogin);
  }

  const res2 = await fetch('/api/user/me', { credentials: 'include' });
  if (res2.ok) {
    window.location.href = safeNext();
    return;
  }
}

/** Resolves once the page has applied the login policy (awaited by tests). */
export const loginReady: Promise<void> = init();
