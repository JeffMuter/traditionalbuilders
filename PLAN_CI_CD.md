# PLAN_CI_CD.md — CI/CD for Traditional Builders

Status: **items 1–4 implemented, verified locally.** Companion to `PLAN_Testing.md`.

## Audit summary (pre-change)

- **No CI/CD existed**: no `.github/workflows`, no Dockerfile, no deploy config,
  no git hooks. Tests only ran if a human typed `make test`.
- Local automation was healthy: `make`/`scripts`, `Procfile` + `.air.toml` +
  overmind (dev only), `shell.nix` (reproducible toolchain), 10 `*_test.go`
  files across 3 packages.
- Verified green locally before starting: `go test ./...`, `go test -race`, `go
  vet ./...`, `gofmt -l .` (clean).
- Key constraint: generated artifacts (`templates/*_templ.go`,
  `static/css/output.css`) are gitignored, so every pipeline must regenerate
  before it builds or tests.

## Scope of this plan (items 1–4)

### 1. GitHub Actions CI on push/PR
`.github/workflows/ci.yml`. Runs on push to `master` and on PRs. Resolves the
toolchain without Nix (runners are Ubuntu):

- `actions/setup-go` with `go-version-file: go.mod`
- `go install github.com/a-h/templ/cmd/templ@<pinned>` — `go install` works even
  though `go run .../cmd/templ` fails on missing `go.sum` entries
- `npm install -g tailwindcss@3.4.17` (matches `shell.nix`'s 3.4.17)
- `make ci` (see item 2)

### 2. `make ci` — one command shared by CI and humans
A single target that is the source of truth for "is this build healthy":

```
templ generate
tailwindcss -i static/css/input.css -o static/css/output.css
gofmt -l . (must be empty)
go vet ./...
go test -race ./...
go build -o bin/server ./cmd/server
```

CI installs the tools then calls `make ci`, so local and CI behaviour cannot
drift. `make test` and `make build` remain for focused use.

### 3. Deploy target + release artifact job
**Decision: self-hosted VPS + systemd, SQLite on a persistent disk.**

Rationale:
- The app is a single Go binary with a **local SQLite file** — a single-writer,
  stateful process. That is the natural shape for one small VPS with a real
  filesystem, not ephemeral/serverless compute.
- Fly.io/Railway can host it but require a mounted volume for the DB and extra
  account setup; a VPS + systemd needs no new vendor and matches the existing
  "build a binary, run it" workflow.
- Container hosting stays possible: a multi-stage `Dockerfile` is included as a
  portable alternative unit.

Deliverables:
- `Dockerfile` + `.dockerignore` — reproducible image (binary + `static/` +
  `db/migrations/`), `ADDR=:8080`, DB on a mounted volume.
- `deploy/traditionalbuilders.service` — hardened systemd unit template.
- `deploy/deploy.sh` — build locally, ship tarball to the host, restart unit.
- `.github/workflows/release.yml` — on `v*` tags, build `linux/amd64` +
  `linux/arm64` release tarballs (binary + `static/` + `db/migrations/` +
  systemd unit) and attach them to a GitHub Release.

### 4. gofmt + vet as required checks
Both already run inside `make ci` and therefore in the CI workflow. The
remaining step is **branch protection**, which is a GitHub repo setting, not a
file — documented here and called out after implementation:

> Settings → Branches → protect `master` → require status checks
> `lint` and `test` and `build` before merge.

## What is explicitly out of scope

- No database migration job in CI (migrations run on the host at deploy).
- No secrets/host credentials are committed; `deploy.sh` reads `DEPLOY_HOST`
  from the environment.
- No coverage threshold gate yet — coverage is uploaded as an artifact only.

## Files added/changed

| File | Change |
|---|---|
| `PLAN_CI_CD.md` | this plan |
| `Makefile` | add `ci`, `release` targets |
| `.github/workflows/ci.yml` | new — lint + test + build on push/PR |
| `.github/workflows/release.yml` | new — tag-driven release artifacts |
| `Dockerfile` | new — portable deploy unit |
| `.dockerignore` | new |
| `deploy/traditionalbuilders.service` | new — systemd unit |
| `deploy/deploy.sh` | new — VPS deploy script |
| `.gitignore` | add `dist/` |

## Verification

- `make ci` passes locally (gofmt + vet + `go test -race` + build). ✅
- Workflows validated with `actionlint` + `yamllint`. ✅
- `deploy/deploy.sh` passes `shellcheck`. ✅
- `systemd-analyze verify deploy/traditionalbuilders.service` parses. ✅
- `docker build` succeeds; container serves `/healthz` → `ok`. ✅
- `make release` produces both `linux/amd64` and `linux/arm64` tarballs (binary +
  `static/` + `migrations/` + unit file). ✅

## Remaining manual step (item 4)

Branch protection is a GitHub repo setting, not a file, and needs authenticated
API access (no `gh` CLI on this machine). Do once in the web UI:

> Settings → Branches → Add branch protection rule for `master` →
> Require status checks to pass → select `lint`, `test`, `build`.

Once enabled, direct pushes to `master` are blocked unless all three jobs pass.
