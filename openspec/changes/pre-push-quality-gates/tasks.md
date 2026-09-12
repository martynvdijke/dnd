# Tasks: pre-push-quality-gates

## 1. Verify and unify build tags

- [x] 1.1 Determine whether modernc.org/sqlite honors the `sqlite_fts5` build tag (check dependency source/docs and confirm FTS5 works with and without the tag); record the outcome in design.md under D4
- [x] 1.2 Based on 1.1, unify all Go invocations to one tag set: package.json (`test`), Taskfile (`test`, `build`), `.pre-commit-config.yaml` (go-vet, go-build hooks), and keep CI as-is if it already matches

## 2. Shared check scripts

- [x] 2.1 Create `scripts/ci/check-tidy.sh` (go mod tidy + `git diff --exit-code go.mod go.sum`) and `scripts/ci/check-fmt.sh` (`gofmt -l .` must be empty), matching ci.yaml
- [x] 2.2 Create `scripts/ci/check-vet.sh` and `scripts/ci/check-ts.sh` (npm ci-safe install check, `npm run build:vite`, `npm run typecheck`)
- [x] 2.3 Extract the data-testid lint loop from ci.yaml into `scripts/ci/check-testid.sh` unchanged (including the `nav-*` exception)
- [x] 2.4 Create `scripts/ci/test-go.sh` with the CI test invocation plus the awk coverage gates (total ≥20%, handlers ≥30%, middleware ≥25%), printing which threshold failed on error
- [x] 2.5 Create `scripts/ci/test-vitest.sh` (`npx vitest run --coverage`, enforcing vitest.config.ts thresholds)
- [x] 2.6 Create `scripts/ci/build-server.sh` (`go build` with the unified tags, output `villum-server`)
- [x] 2.7 Create `scripts/ci/test-e2e.sh` that runs `build-server.sh` first, cleans `villum.db*`, then `npx playwright test --project=chromium`
- [x] 2.8 Create `scripts/ci/test-docker.sh` that runs the Docker build smoke test when a daemon is available, otherwise prints a loud warning and exits 0
- [x] 2.9 Create `scripts/ci/run-all.sh` sequencing 2.1–2.8 with clear per-gate output and fail-fast behavior; make all scripts executable

## 3. CI consumes the scripts

- [x] 3.1 Refactor `.github/workflows/ci.yaml` lint-typecheck and unit-test jobs to call the `scripts/ci/` scripts instead of inline commands
- [x] 3.2 Refactor the e2e job to call `scripts/ci/build-server.sh`, `scripts/ci/test-docker.sh`, and the playwright portion of `test-e2e.sh` (keeping CI-specific steps like `playwright install` in the workflow)
- [ ] 3.3 Push the refactor on a branch and confirm a full CI run stays green with identical gate behavior

## 4. Taskfile and prek pre-push hook

- [x] 4.1 Add `ci` task to Taskfile invoking `scripts/ci/run-all.sh`; add `deps: [build]` to `test:e2e`
- [x] 4.2 Add a pre-push stage local hook to `.pre-commit-config.yaml` running `task ci` (or `scripts/ci/run-all.sh`), `pass_filenames: false`, `always_run: true`
- [ ] 4.3 Install hooks: `prek install` and `prek install --hook-type pre-push`; verify `prek run --hook-stage pre-push` executes the suite
- [ ] 4.4 Negative test: introduce a deliberate failure (e.g. unformatted Go file), confirm the pre-push hook blocks and reports it, then revert

## 5. Documentation and close-out

- [x] 5.1 Update CONTRIBUTING.md: replace the manual CI-parity command list with `task ci`, document prek setup (`prek install --hook-type pre-push`), the `--no-verify` escape hatch, and the rule that CI steps must call `scripts/ci/`
- [x] 5.2 Confirm AGENTS.md gate references match the final commands (`task ci`, prek hook behavior)
- [ ] 5.3 Full local validation: `task ci` green from a clean state, then push and confirm CI green end-to-end
