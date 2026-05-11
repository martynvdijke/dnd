(() => {

function toggleTheme() {
  const html = document.documentElement;
  const isDark = html.getAttribute('data-theme') === 'dark';
  const newTheme = isDark ? 'light' : 'dark';
  html.setAttribute('data-theme', newTheme);
  localStorage.setItem('villum-theme', newTheme);
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = isDark ? 'fa-solid fa-moon' : 'fa-solid fa-sun';
  const label = document.getElementById('themeLabel');
  if (label) label.textContent = isDark ? 'Dark Mode' : 'Light Mode';
}
(window as any).toggleTheme = toggleTheme;

function initTheme() {
  const saved = localStorage.getItem('villum-theme') || 'light';
  document.documentElement.setAttribute('data-theme', saved);
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = saved === 'dark' ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
  const label = document.getElementById('themeLabel');
  if (label) label.textContent = saved === 'dark' ? 'Light Mode' : 'Dark Mode';
}

async function init() {
  initTheme();
  const res = await fetch('/api/check-setup');
  const data = await res.json();
  if (data.setup) {
    window.location.href = '/login';
    return;
  }

  const form = document.getElementById('setupForm') as HTMLFormElement;
  const errorDiv = document.getElementById('error') as HTMLDivElement;

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorDiv.classList.add('d-none');

    const username = (document.getElementById('username') as HTMLInputElement).value;
    const password = (document.getElementById('password') as HTMLInputElement).value;
    const confirm = (document.getElementById('confirm') as HTMLInputElement).value;

    if (password !== confirm) {
      errorDiv.textContent = 'Passwords do not match';
      errorDiv.classList.remove('d-none');
      return;
    }

    if (password.length < 8) {
      errorDiv.textContent = 'Password must be at least 8 characters';
      errorDiv.classList.remove('d-none');
      return;
    }

    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password, setup: true }),
      credentials: 'include',
    });

    if (res.ok) {
      window.location.href = '/';
    } else {
      const err = await res.json();
      errorDiv.textContent = err.error || 'Setup failed';
      errorDiv.classList.remove('d-none');
    }
  });
}

init();

})();
