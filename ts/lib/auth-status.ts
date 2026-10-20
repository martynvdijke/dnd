/**
 * Client view of the server's login policy.
 *
 * `GET /api/auth/oidc/status` reports two independent things: whether OIDC
 * (Authelia) is configured, and whether password login is allowed at all.
 * Password login is retired by default and only comes back when the operator
 * sets the `ALLOW_PASSWORD_FALLBACK` break-glass flag, so the login and setup
 * pages have to ask before they render a password form.
 */
export interface AuthStatus {
  /** OIDC is configured and usable, so the Authelia button works. */
  enabled?: boolean;
  /** Break-glass state: true only when ALLOW_PASSWORD_FALLBACK is set. */
  password_enabled?: boolean;
}

/**
 * Reads the login policy, or null when the server did not answer.
 *
 * Null means "unknown", not "nothing is enabled": a page that cannot reach the
 * status endpoint keeps whatever it was already showing rather than hiding the
 * only login it might have.
 */
export async function fetchAuthStatus(): Promise<AuthStatus | null> {
  try {
    const res = await fetch('/api/auth/oidc/status', { credentials: 'include' });
    if (!res.ok) return null;
    return (await res.json()) as AuthStatus;
  } catch {
    return null;
  }
}

/**
 * Whether this deployment offers password login.
 *
 * `password_enabled: false` is the break-glass flag being off, so the password
 * form is hidden — the server answers `/api/login` with 403 in that state. A
 * response that predates the field means password login was never retired on
 * that server, and unknown (null) keeps the form, so a status hiccup can never
 * take away a working login path.
 */
export function passwordLoginOffered(status: AuthStatus | null): boolean {
  if (!status) return true;
  return status.password_enabled !== false;
}

/** Whether this deployment can send people to Authelia. */
export function autheliaOffered(status: AuthStatus | null): boolean {
  return status?.enabled === true;
}
