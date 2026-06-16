import { test as base, expect } from '@playwright/test';
import type { Page, BrowserContext } from '@playwright/test';
import { spawn, spawnSync, type ChildProcess } from 'child_process';
import http from 'http';
import fs from 'fs';

// Re-export runtime values and types that test files need from @playwright/test
export { expect };
export type { Page, BrowserContext };

// ─── Worker-isolated server fixture ───
//
// Each Playwright worker gets its own server instance on a unique port with a
// dedicated SQLite database. This eliminates DB corruption from concurrent
// writes when tests run in parallel across multiple workers.
//
// Server lifecycle:
//   port     = 6270 + workerIndex        (e.g. worker 0 → 6270, worker 1 → 6271)
//   db  path = /tmp/villum-test-${workerIndex}.db
//   env      = AUTO_SETUP=true            (creates admin/testpassword123 automatically)

const BASE_PORT = 6270;
const DB_DIR = '/tmp';
const SERVER_BIN = process.env.SERVER_BIN || './villum-server';
const WORKER_TIMEOUT = 30_000;

async function waitForPort(port: number): Promise<void> {
  const deadline = Date.now() + WORKER_TIMEOUT;
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
  throw new Error(
    `Server on port ${port} not ready within ${WORKER_TIMEOUT}ms`,
  );
}

function cleanupServer(proc: ChildProcess | null, dbPath: string) {
  if (proc) {
    // SIGKILL immediately — test servers have no state worth preserving,
    // and a delay via setTimeout may not fire before Node exits.
    try {
      proc.kill('SIGKILL');
    } catch {
      // already dead
    }
  }
  // Remove DB files
  for (const ext of ['', '-shm', '-wal']) {
    try {
      fs.unlinkSync(dbPath + ext);
    } catch {
      // file doesn't exist
    }
  }
}

/**
 * Kill any process currently listening on {@link port}.
 * Handles orphaned server processes left behind when a previous test worker
 * was killed before its cleanup could run (e.g. on test timeout).
 */
function freePort(port: number): void {
  try {
    spawnSync('fuser', [`${port}/tcp`, '-k'], {
      timeout: 3000,
      stdio: 'ignore',
    });
  } catch {
    // fuser not available — port conflict will surface as EADDRINUSE
  }
}

type WorkerData = {
  port: number;
  proc: ChildProcess;
  dbPath: string;
};

// Worker-scoped fixture: starts a Go server per worker with isolated DB.
export const test = base.extend<{}, { workerData: WorkerData }>({
  workerData: [
    async ({}, use, workerInfo) => {
      const port = BASE_PORT + workerInfo.workerIndex;
      const dbPath = `${DB_DIR}/villum-test-${workerInfo.workerIndex}.db`;
      freePort(port);               // kill orphaned server from a previous aborted run
      cleanupServer(null, dbPath);  // clean any DB leftovers from previous runs

      const proc = spawn(SERVER_BIN, [], {
        env: {
          ...process.env,
          DB_PATH: dbPath,
          PORT: String(port),
          AUTO_SETUP: 'true',
        },
        stdio: 'pipe',
        detached: false,
      });

      proc.stdout?.on('data', (d: Buffer) => {
        process.stdout.write(`[worker ${workerInfo.workerIndex}] ${d}`);
      });
      proc.stderr?.on('data', (d: Buffer) => {
        process.stderr.write(`[worker ${workerInfo.workerIndex}] ${d}`);
      });
      proc.on('exit', (code, signal) => {
        if (code !== 0 && code !== null) {
          process.stderr.write(
            `[worker ${workerInfo.workerIndex}] server exited with code=${code} signal=${signal}\n`,
          );
        }
      });

      try {
        await waitForPort(port);
      } catch (err) {
        cleanupServer(proc, dbPath);
        throw err;
      }

      await use({ port, proc, dbPath });

      cleanupServer(proc, dbPath);
    },
    { scope: 'worker' },
  ],
});

// Override baseURL so every test in this worker uses the correct server port.
test.use({
  baseURL: async ({ workerData }, use) => {
    await use(`http://localhost:${workerData.port}`);
  },
});
