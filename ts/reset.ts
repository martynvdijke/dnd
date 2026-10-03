const form = document.getElementById("resetForm") as HTMLFormElement | null;
const msg = document.getElementById("msg") as HTMLDivElement | null;

function showMsg(text: string, cls: string) {
  if (!msg) return;
  msg.textContent = text;
  msg.className = "alert " + cls;
  msg.classList.remove("d-none");
}

form?.addEventListener("submit", async (e) => {
  e.preventDefault();
  const token = new URLSearchParams(window.location.search).get("token") || "";
  if (!token) {
    showMsg("Missing reset token.", "alert-danger");
    return;
  }
  const np = (document.getElementById("newPassword") as HTMLInputElement).value;
  const cp = (document.getElementById("confirmPassword") as HTMLInputElement).value;
  if (np.length < 8) {
    showMsg("Password must be at least 8 characters.", "alert-warning");
    return;
  }
  if (np !== cp) {
    showMsg("Passwords do not match.", "alert-warning");
    return;
  }
  try {
    const res = await fetch("/api/reset-password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token, new_password: np }),
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) {
      showMsg("Password updated. Redirecting to login…", "alert-success");
      setTimeout(() => { window.location.href = "/login"; }, 1500);
    } else {
      showMsg(data.error || "Reset failed.", "alert-danger");
    }
  } catch {
    showMsg("Network error.", "alert-danger");
  }
});
