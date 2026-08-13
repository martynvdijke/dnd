# AGENTS.md — Villum

D&D campaign manager. Go backend (`handlers/`, `middleware/`, `crypto/`, `dice/`, `registry/`, `trmnl/`; modernc.org/sqlite + ent), TypeScript frontend (`ts/`, 5 vite bundles), Playwright e2e (`tests/`), vitest unit tests (`ts/**/*.test.ts`). Task runner: Taskfile. Hooks: prek (pre-commit compatible). Specs: OpenSpec (`openspec/`).

## The Workflow Contract (read first)

**Work is done when the PR is merged green — not when the code is written.**

1. **Never push directly to `main`.** All work goes through a PR, no exceptions.
2. Branch: `git checkout -b <type>/<short-name>` (e.g. `fix/login-redirect`).
3. Commit with **Conventional Commits** (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). semantic-release derives the version from these messages on `main` — a bad message means a wrong release.
4. **Before pushing, run the local gates.** The prek pre-push hook runs the full CI-parity suite automatically (~30 min including e2e — this is expected, let it finish). If hooks are not installed, run `task ci` manually first. Never push red.
5. Push and open the PR: `git push -u origin <branch>` then `gh pr create --fill`.
6. **Wait for checks**: `gh pr checks --watch`. Do not consider the task complete while checks are pending.
7. **Green** → merge: `gh pr merge --squash --delete-branch` (squash title must be a valid Conventional Commit).
8. **Red** → get the failing logs first: `gh run view --log-failed`. Fix, commit, push to the same branch, go to step 6. Do not merge red. Do not ask the user to investigate — read the logs yourself.
9. If a **release on `main` fails** (user reports a Gotify failure): your first action is `gh run list --workflow=release.yaml --limit 5` then `gh run view <id> --log-failed`. Diagnose from logs, fix via the PR flow above.

## Local Gates

| When | What | Command |
|---|---|---|
| every commit | hygiene, go vet, go build, tsc, vitest | automatic (prek commit hooks) |
| every push | full CI parity incl. e2e + coverage gates | automatic (prek pre-push hook) or `task ci` |
| on demand | full suite | `task ci` |
| on demand | pieces | `task test`, `task test:e2e`, `npm run test:unit`, `npm run build:vite`, `npm run typecheck` |

Fresh clone setup: `prek install && prek install --hook-type pre-push`.

`task test:e2e` rebuilds `./villum-server` first — Playwright fixtures spawn that binary, and it embeds the frontend assets. Never run `npx playwright test` against a stale binary.

## Testing Standards (details in CONTRIBUTING.md)

- Every `data-testid` used in `ts/` or `static/*.html` MUST be referenced in `tests/` (linted in CI). Dynamic `nav-*` ids are the documented exception.
- E2E uses the `login()` helper from `tests/helpers.ts`; respect `LOGIN_TIMEOUT` (30s) and `NAV_TIMEOUT` (10s); slow tests call `test.slow()`.
- Coverage floors: Go total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%; vitest ≥20% all metrics. Gates run locally and in CI.

## Change Process

Non-trivial work goes through OpenSpec: `/opsx-propose` → artifacts in `openspec/changes/<name>/` → `/opsx-apply` → `/opsx-archive`. Task lists there include verification steps — the local gates above still apply before pushing.
