async function init() {
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
    const form = document.getElementById('loginForm');
    const errorDiv = document.getElementById('error');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        errorDiv.classList.add('d-none');
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
            credentials: 'include',
        });
        if (res.ok) {
            window.location.href = '/';
        }
        else {
            const err = await res.json();
            errorDiv.textContent = err.error || 'Invalid credentials';
            errorDiv.classList.remove('d-none');
        }
    });
}
init();

//# sourceMappingURL=login.js.map