export function toggleTheme() {
    const html = document.documentElement;
    const isDark = html.getAttribute('data-theme') === 'dark';
    const newTheme = isDark ? 'light' : 'dark';
    html.setAttribute('data-theme', newTheme);
    localStorage.setItem('villum-theme', newTheme);
    updateThemeIcon();
}
export function updateThemeIcon() {
    const icon = document.getElementById('themeIcon');
    if (!icon)
        return;
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    icon.className = isDark ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
}
export function initTheme() {
    const saved = localStorage.getItem('villum-theme') || 'light';
    document.documentElement.setAttribute('data-theme', saved);
    updateThemeIcon();
}
