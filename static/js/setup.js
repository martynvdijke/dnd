async function init() {
    const res = await fetch('/api/check-setup');
    const data = await res.json();
    if (data.setup) {
        window.location.href = '/login';
        return;
    }
    const form = document.getElementById('setupForm');
    const errorDiv = document.getElementById('error');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        errorDiv.style.display = 'none';
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const confirm = document.getElementById('confirm').value;
        if (password !== confirm) {
            errorDiv.textContent = 'Passwords do not match';
            errorDiv.style.display = 'block';
            return;
        }
        if (password.length < 8) {
            errorDiv.textContent = 'Password must be at least 8 characters';
            errorDiv.style.display = 'block';
            return;
        }
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password, setup: true }),
            credentials: 'include',
        });
        if (res.ok) {
            window.location.href = '/app';
        }
        else {
            const err = await res.json();
            errorDiv.textContent = err.error || 'Setup failed';
            errorDiv.style.display = 'block';
        }
    });
}
init();
export {};
//# sourceMappingURL=setup.js.map