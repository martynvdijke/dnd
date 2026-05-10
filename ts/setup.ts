export {};

async function init() {
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
