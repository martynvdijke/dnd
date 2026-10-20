import { test, expect } from './fixtures.js';
import { spawn, spawnSync, type ChildProcess } from 'child_process';
import http from 'http';
import fs from 'fs';

// ─── Password-login retirement cutover + documented rollback ───
//
// The shared worker fixture always spawns with `ALLOW_PASSWORD_FALLBACK=true`
// (the whole suite signs in with a password), so it cannot observe the retired
// default. This spec therefore drives its OWN server process on a dedicated
// port and DB, and kills/restarts it to walk the real cutover:
//
//   phase 1  ALLOW_PASSWORD_FALLBACK=true   → password login works, cookie issued
//   phase 2  ALLOW_PASSWORD_FALLBACK=false  → the cutover: new password logins
//                                             are 403 and the phase-1 session
//                                             stops authenticating (deleted)
//   phase 3  ALLOW_PASSWORD_FALLBACK=true   → rollback: a fresh password login
//                                             works again, while the phase-1
//                                             session stays dead
//
// The same DB file is reused across phases so the phase-1 session row is real
// and its death is caused by the cutover, not by a missing database.

const DB_DIR = '/tmp';
const SERVER_BIN = process.env.SERVER_BIN || './villum-server';
const READY_TIMEOUT = 30_000;

const ADMIN_USERNAME = 'admin';
const ADMIN_PASSWORD = 'testpassword123';

async function waitForPort(port: number): Promise<void> {
  const deadline = Date.now() + READY_TIMEOUT;
  while (Date.now() < deadline) {
    try {
      const ok = await new Promise<boolean>((resolve) => {
        const req = http.get(`http://localhost:${port}/api/check-setup`, (res) => {
          resolve(res.statusCode === 200);
        });
        req.on('error', () => resolve(false));
        req.setTimeout(2000, () => {
          req.destroy();
          resolve(false);
        });
      });
      if (ok) return;
    } catch {
      // retry
    }
    await new Promise((r) => setTimeout(r, 300));
  }
  throw new Error(`Server on port ${port} not ready within ${READY_TIMEOUT}ms`);
}

/** Kill any process listening on `port` (orphans from aborted runs). */
function freePort(port: number): void {
  try {
    spawnSync('fuser', [`${port}/tcp`, '-k'], { timeout: 3000, stdio: 'ignore' });
  } catch {
    // fuser unavailable — a conflict will surface as EADDRINUSE
  }
}

function cleanupDb(dbPath: string): void {
  for (const ext of ['', '-shm', '-wal']) {
    try {
      fs.unlinkSync(dbPath + ext);
    } catch {
      // file doesn't exist
    }
  }
}

/**
 * Start a server on `port`/`dbPath` with `ALLOW_PASSWORD_FALLBACK=fallback`.
 * OIDC_* variables are stripped so the login page deterministically renders the
 * "no way to sign in" notice in phase 2 (Authelia absent) rather than a
 * configured-SSO deployment.
 */
async function startServer(opts: {
  port: number;
  dbPath: string;
  fallback: boolean;
}): Promise<ChildProcess> {
  freePort(opts.port);
  const env: NodeJS.ProcessEnv = { ...process.env };
  for (const key of Object.keys(env)) {
    if (key.startsWith('OIDC_')) delete env[key];
  }
  env.DB_PATH = opts.dbPath;
  env.PORT = String(opts.port);
  env.AUTO_SETUP = 'true';
  env.ALLOW_PASSWORD_FALLBACK = opts.fallback ? 'true' : 'false';
  env.BASE_URL = `http://localhost:${opts.port}`;

  const proc = spawn(SERVER_BIN, [], { env, stdio: 'pipe', detached: false });
  proc.stdout?.on('data', () => {});
  proc.stderr?.on('data', (d: Buffer) => {
    process.stderr.write(`[pw-retire] ${d}`);
  });
  await waitForPort(opts.port);
  return proc;
}

async function stopServer(proc: ChildProcess, port: number): Promise<void> {
  if (proc.exitCode === null) {
    const exited = new Promise<void>((resolve) => proc.once('exit', () => resolve()));
    try {
      proc.kill('SIGKILL');
    } catch {
      // already gone
    }
    await Promise.race([exited, new Promise<void>((r) => setTimeout(r, 5000))]);
  }
  freePort(port);
}

function sessionCookieOf(res: Response): string | undefined {
  return res.headers
    .getSetCookie()
    .map((c) => c.split(';')[0])
    .find((c) => c.startsWith('session='));
}

function passwordLogin(base: string): Promise<Response> {
  return fetch(`${base}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: ADMIN_USERNAME, password: ADMIN_PASSWORD }),
  });
}

function userMe(base: string, cookie: string): Promise<Response> {
  return fetch(`${base}/api/user/me`, { headers: { Cookie: cookie } });
}

test.describe('password login retirement (break-glass cutover + rollback)', () => {
  test('flag on → off → on: cutover kills password sessions, rollback restores login', async ({
    page,
  }) => {
    // slow: spawns its own Go server and restarts it twice (three cold starts,
    // each running migrations + AUTO_SETUP bootstrap).
    test.slow();

    const workerIndex = test.info().workerIndex;
    const port = 6410 + workerIndex;
    const dbPath = `${DB_DIR}/villum-pwretire-${workerIndex}.db`;
    const base = `http://localhost:${port}`;

    let proc: ChildProcess | null = null;

    try {
      // Fresh slate: no orphaned server, no stale DB from a prior run.
      freePort(port);
      cleanupDb(dbPath);

      // ── Phase 1: break-glass ON — password login is a normal path. ──
      proc = await startServer({ port, dbPath, fallback: true });
      const phase1Login = await passwordLogin(base);
      expect(phase1Login.status).toBe(200);
      const phase1Cookie = sessionCookieOf(phase1Login);
      expect(phase1Cookie, 'phase 1 must issue a password session cookie').toBeTruthy();
      const phase1Me = await userMe(base, phase1Cookie!);
      expect(phase1Me.status).toBe(200);
      expect(await phase1Me.json()).toMatchObject({ username: ADMIN_USERNAME });

      // Restart the SAME db with the flag off — the cutover.
      await stopServer(proc, port);
      proc = await startServer({ port, dbPath, fallback: false });

      // ── Phase 2: flag OFF — retired by default. ──
      const statusOff = await fetch(`${base}/api/auth/oidc/status`);
      expect(statusOff.status).toBe(200);
      expect(await statusOff.json()).toMatchObject({ password_enabled: false });

      // Correct credentials are refused and no session cookie is minted.
      const phase2Login = await passwordLogin(base);
      expect(phase2Login.status).toBe(403);
      expect(await phase2Login.json()).toMatchObject({ error: 'password login is disabled' });
      expect(sessionCookieOf(phase2Login), '403 login must not set a session cookie').toBeUndefined();

      // Requirement: a pre-cutover password session stops authenticating.
      const reused = await userMe(base, phase1Cookie!);
      expect(reused.status).toBe(401);
      expect(await reused.json()).toMatchObject({ error: 'session expired' });

      // Password reset is gated the same way.
      const forgot = await fetch(`${base}/api/forgot-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: ADMIN_USERNAME }),
      });
      expect(forgot.status).toBe(403);
      expect(await forgot.json()).toMatchObject({ error: 'password login is disabled' });

      // The login page names the recovery switch and offers no password form.
      const loginPosts: string[] = [];
      page.on('request', (req) => {
        if (req.method() === 'POST' && req.url().endsWith('/api/login')) {
          loginPosts.push(req.url());
        }
      });
      await page.context().clearCookies();
      await page.goto(`${base}/login`, { waitUntil: 'domcontentloaded' });
      const notice = page.locator('#loginUnavailable');
      await expect(notice).toBeVisible();
      await expect(notice).toContainText('ALLOW_PASSWORD_FALLBACK');
      await expect(page.locator('#loginForm')).toBeHidden();
      // The page applied its policy before we assert nothing posted a password.
      expect(loginPosts).toHaveLength(0);

      // ── Phase 3: rollback — flag back ON restores password login. ──
      await stopServer(proc, port);
      proc = await startServer({ port, dbPath, fallback: true });

      const statusOn = await fetch(`${base}/api/auth/oidc/status`);
      expect(statusOn.status).toBe(200);
      expect(await statusOn.json()).toMatchObject({ password_enabled: true });

      const phase3Login = await passwordLogin(base);
      expect(phase3Login.status).toBe(200);
      const phase3Cookie = sessionCookieOf(phase3Login);
      expect(phase3Cookie, 'rollback must issue a new password session').toBeTruthy();
      const phase3Me = await userMe(base, phase3Cookie!);
      expect(phase3Me.status).toBe(200);

      // The cutover deleted the phase-1 session; a rollback does not revive it.
      const stillDead = await userMe(base, phase1Cookie!);
      expect(stillDead.status).toBe(401);
      expect(await stillDead.json()).toMatchObject({ error: 'session expired' });
    } finally {
      if (proc) await stopServer(proc, port);
      cleanupDb(dbPath);
    }
  });
});
