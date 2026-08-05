# Contributing to Villum (D&D)

Thanks for contributing! This document records the standards the CI pipeline
enforces and the conventions e2e and unit tests should follow.

## Development setup

```sh
go mod download
npm install
npm run build:vite   # builds static bundles (go:embed requires them before `go build`)
```

Run the server locally:

```sh
go build -tags sqlite_fts5 -o villum-server .
AUTO_SETUP=true ./villum-server
```

## Commands

| Command              | Purpose                                        |
| -------------------- | ---------------------------------------------- |
| `npm run build:vite` | Build all frontend entry bundles               |
| `npm run typecheck`  | `tsc --noEmit` across `ts/**`                  |
| `npm run build:ts`   | `build:vite` + `typecheck`                     |
| `npm test`           | Go tests with the FTS5 tag (`go test -tags fts5 ./...`) |
| `npm run test:unit`  | Vitest unit tests                              |
| `npm run test:e2e`   | Playwright e2e                                 |
| `npm run typecheck`  | See above                                      |

### CI-parity check before pushing

The CI pipeline runs all of the following. Run them locally first:

```sh
go mod tidy && git diff --exit-code go.mod go.sum
gofmt -l . && test -z "$(gofmt -l .)"
go test ./...
go test -tags sqlite_fts5 ./...
go vet -tags sqlite_fts5 ./...
npx tsc --noEmit
npx vitest run
npm run build:vite
```

Remember to rebuild `./villum-server` after any change to `ts/` or `static/`
before running Playwright locally — the binary embeds the static files, so a
stale binary serves stale assets.

## Test standards

### E2E (Playwright)

- **Selectors**: prefer `data-testid` attributes (`getByTestId`) or `data-nav`
  for navigation. Every `data-testid` in `ts/` and the static HTML must be
  referenced by at least one test — the CI `lint-typecheck` job fails otherwise.
- **Login**: always use the shared `login(page)` helper from `tests/helpers.ts`.
  Never inline a login flow in a spec.
- **Timeouts**: use the shared constants `LOGIN_TIMEOUT` (30s) and `NAV_TIMEOUT`
  (10s) from `tests/helpers.ts` instead of inline numbers.
- **Slow tests**: call `test.slow()` first inside heavy tests, mobile-viewport
  tests, and any test that performs a full SPA init or waits on network flows
  (`waitLoadingDone`, `waitForFunction`).
- **CI-only tests**: guard with `test.skip(!process.env.CI, 'Runs only in CI')`
  (see `tests/visual-ci.spec.ts`) and fix a viewport when layout matters.
- **Fixtures**: the test worker spawns its own `./villum-server` on a per-worker
  port with an isolated database — never depend on a manually started server.

### Unit (Vitest / Go)

- Vitest coverage threshold is enforced at 20% (see `vitest.config.ts`).
- Go per-package coverage thresholds are enforced in CI: `handlers/` ≥ 30%,
  `middleware/` ≥ 25%, total ≥ 20%. Document any intentional change in
  `.github/workflows/ci.yaml` next to the thresholds.

## CI pipeline

`ci.yaml` (reusable via `workflow_call`, also runs on pull requests) has three
jobs:

1. **lint-typecheck** — `go mod tidy` diff check, `gofmt`, vet, TypeScript
   typecheck, and the data-testid reference lint.
2. **unit-test** — Go tests with coverage (thresholds above) and Vitest with
   coverage.
3. **e2e** — frontend build, server binary, a Docker build smoke test, then the
   Playwright chromium suite (CI mode, 30-minute budget).

The Release workflow is gated on the `ci / ci` job. Keep it green.

## Release process

- Semantic-release computes the next version from `package.json` (the only file
  mutated by the release `exec` plugin) and commits `CHANGELOG.md` +
  `package.json`.
- The Docker image receives the released version at build time via
  `--build-arg VERSION` and Go `-ldflags "-X main.Version=..."`. Local
  non-release builds fall back to `main.Version = "0.0.0-dev"`.
- Health endpoint: `GET /healthz` reports `"version"` from the injected value.
