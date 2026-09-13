# Design: pre-push-quality-gates

## Context

CI (`.github/workflows/ci.yaml`) enforces gates that have no local equivalent: `go mod tidy` diff check, `gofmt -l .`, `build:vite`, `data-testid` reference lint, Go coverage thresholds (total ≥20%, handlers ≥30%, middleware ≥25%), vitest coverage thresholds (20% all), and a Docker build smoke test. Local invocations also disagree with CI on Go build tags (`fts5` in package.json + pre-commit, `sqlite_fts5` in CI, none in Taskfile), and `task test:e2e` runs Playwright against `./villum-server` without rebuilding it — the binary embeds static assets, so local e2e can pass against yesterday's code. Failures surface via Gotify after a push to `main`, when the authoring session is gone.

The user pushes directly to `main`, accepts ~30 min of local gate time per feature, and wants hooks managed through **prek** (Rust, single-binary, drop-in pre-commit replacement). Agent behavior (PR workflow) is covered separately in AGENTS.md, not in this change.

## Goals / Non-Goals

**Goals:**
- Every CI gate runnable locally with the *same command* CI uses — one definition, two callers.
- A pre-push hook (prek) that runs the full parity suite and blocks the push on failure.
- Zero tag drift across package.json, Taskfile, pre-commit, and CI.
- e2e always runs against a freshly built `./villum-server`.
- CONTRIBUTING.md documents one parity command, not a manual checklist.

**Non-Goals:**
- Changing what CI enforces (thresholds, jobs, projects stay as they are).
- Running the full suite per commit (fast hooks stay at commit stage unchanged).
- Branch protection / GitHub repo settings (out of band, optional hardening).
- The agent PR workflow contract (lives in AGENTS.md, written alongside this change).

## Decisions

### D1: Shared scripts under `scripts/ci/` are the single source of truth
One script per gate: `check-tidy.sh`, `check-fmt.sh`, `check-vet.sh`, `check-ts.sh` (build:vite + tsc), `check-testid.sh`, `test-go.sh` (tests + coverage gates), `test-vitest.sh` (vitest --coverage), `build-server.sh`, `test-e2e.sh`, `test-docker.sh`, plus `run-all.sh` sequencing them. `ci.yaml` steps call these scripts instead of inline `run:` blocks; the pre-push hook calls `run-all.sh` (via `task ci`).
*Alternatives rejected:* status-quo duplication (already drifted); `act` to run the workflow locally (heavy, poor ergonomics, Docker-in-Docker for the smoke test); a single monolithic script (loses per-step structure in GitHub Actions UI).

### D2: Full suite gates at pre-push, not pre-commit
The suite includes e2e (~30 min). Per-commit would be untenable; per-push is the natural boundary — commits stay cheap, nothing leaves the machine ungated.
*Alternative rejected:* gating only in GitHub PR checks (that is the status quo failure mode: feedback after the authoring context is gone).

### D3: prek as the hook runner
User preference; single Rust binary, reads the existing `.pre-commit-config.yaml` unchanged. Hooks are plain `local`/`system` hooks, so falling back to `pre-commit` later costs nothing. Pre-push stage registered via `prek install --hook-type pre-push` (verified during implementation; prek claims full pre-commit compatibility).

### D4: Unify on `-tags sqlite_fts5` unless proven a no-op
CI and CONTRIBUTING already use `sqlite_fts5`; align package.json, Taskfile, and pre-commit to it. Implementation includes a verification step: if modernc.org/sqlite proves to ignore the tag, drop tags everywhere instead — either way, exactly one spelling survives.

### D5: e2e freshness is enforced by the script, not documentation
`test-e2e.sh` runs `build-server.sh` first; Taskfile `test:e2e` gains `deps: [build]`. The stale-binary warning in CONTRIBUTING becomes belt-and-suspenders, not the only defense.

### D6: Docker smoke test adapts to the environment
`test-docker.sh` runs the build when `docker` is available; if not, it prints a loud warning and skips (a missing daemon is an environment fact, not a gate failure). CI always has Docker, so parity is preserved where it matters.

## Risks / Trade-offs

- [Push latency grows to ~30 min including e2e] → Accepted by the user; `git push --no-verify` documented as the escape hatch; fast gates remain at commit time for early feedback.
- [Local e2e flakiness produces false blocks] → Playwright `retries: 2` already configured; local runs use chromium-only, matching CI; flaky list review becomes a habit via CI's `flaky-tests.log` artifact.
- [prek is a young tool] → Config remains valid pre-commit YAML; hooks are standard system hooks; reverting to `pre-commit` is a one-command change.
- [Hook never installed on a fresh clone, gates silently absent] → Install command documented in CONTRIBUTING.md and AGENTS.md; `task ci` works without hooks so the suite is always one command away.
- [Future edits to ci.yaml could re-introduce inline commands and drift] → With all logic in scripts there is nothing inline to drift; CONTRIBUTING gains a rule: CI steps must call `scripts/ci/`.

## Migration Plan

1. Add `scripts/ci/` scripts (extract commands verbatim from ci.yaml, including coverage awk gates).
2. Verify the build-tag question (D4), then unify tags across package.json, Taskfile, pre-commit.
3. Refactor ci.yaml to call the scripts; confirm a CI run stays green.
4. Add `ci` task to Taskfile; add `deps: [build]` to `test:e2e`.
5. Add pre-push hook to `.pre-commit-config.yaml`; `prek install --hook-type pre-push`; validate with `prek run --hook-stage pre-push` and a deliberate-failure test.
6. Update CONTRIBUTING.md (one parity command, prek setup, scripts-rule).

Rollback: hooks are opt-in per clone; scripts are additive; the ci.yaml refactor is a single revert.

## Open Questions

- Does modernc.org/sqlite honor `sqlite_fts5` at all? (Resolved by verification task; outcome decides unify-vs-drop in D4.)

**Verification (2026-09):** Checked `go.mod` (`modernc.org/sqlite v1.56.0` via `sqlite v1.37.1`). Grepped `~/go/pkg/mod/modernc.org/sqlite*` — `grep -r sqlite_fts5` returns no matches and no `//go:build` guard references; FTS5 is enabled unconditionally via `-DSQLITE_ENABLE_FTS5` in the ccgo-generated `lib/sqlite_*.go` files. The `sqlite_fts5` tag is therefore a harmless no-op on this toolchain, but all invocations already unify on it, so the design keeps the tag as the single canonical spelling to avoid drift.
