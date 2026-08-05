export function toggleTheme(): void {
  const html = document.documentElement;
  const isDark = html.getAttribute('data-theme') === 'dark';
  const newTheme = isDark ? 'light' : 'dark';
  html.setAttribute('data-theme', newTheme);
  localStorage.setItem('villum-theme', newTheme);
  updateThemeIcon();
}

export function updateThemeIcon(): void {
  const icon = document.getElementById('themeIcon');
  if (!icon) return;
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
  icon.className = isDark ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
}

export function initTheme(): void {
  const saved = localStorage.getItem('villum-theme') || 'light';
  document.documentElement.setAttribute('data-theme', saved);
  updateThemeIcon();
}

export function toggleEink(): void {
  const html = document.documentElement;
  const active = html.classList.toggle('eink');
  if (active) {
    document.cookie = 'eink=1;path=/;max-age=31536000';
  } else {
    document.cookie = 'eink=;path=/;max-age=0';
  }
}
