# PLAN_FRONTEND_TESTING.md — Front-End Confidence System

Status: **proposed.** Companion to `PLAN_CI_CD.md` and `PLAN_Testing.md`.

This plan covers the two front-end testing systems recommended for adoption. A
third (visual regression) is deliberately deferred — see "Explicitly out of
scope".

## Audit summary (current state)

- **CI tests logic, not the front-end.** `.github/workflows/ci.yml` runs
  `make ci` = templ generate → tailwind → gofmt → vet → `go test -race` →
  build. Every existing test is a handler/store test using `httptest` + an
  in-memory SQLite DB. **No test loads a page in a browser, follows a link,
  clicks a button, checks an image path, or scans accessibility.**
- **Surface is small and fully enumerable.** Routes:

  | Route | Handler | Template |
  |---|---|---|
  | `/` (+ 404 fallthrough) | `handlers.Landing` | `templates.Landing`, `templates.ErrorPage` |
  | `/gallery` | `handlers.Gallery` | `templates.Gallery` |
  | `/builders?zip=` | `h.BuildersPage` | `templates.Builders` (3 states: error / empty / results) |
  | `/builders/{id}` | `h.Profile` | `templates.Profile`, `templates.ProfileNotFound` |
  | `/api/builders?zip=` | `h.SearchBuilders` | JSON |
  | `/healthz` | inline | `ok` |
  | `/static/` | `http.FileServer` | assets |

- **Interactive surface today:** one GET form (zip search), one button (theme
  toggle, localStorage + `dark` class), card → profile links. htmx is loaded on
  every page but **zero `hx-*` attributes exist yet** — the harness must be
  built now so it covers htmx flows when they land.
- **Toolchain available:** Go 1.25, Node 24, system Chromium at
  `/run/current-system/sw/bin/chromium`. No browser-runner dependency installed.
- **NixOS constraint (from `AGENTS.md`):** Chrome/Chromium downloaded by test
  runners does not execute on this machine; the system Chromium does. Any local
  browser runner **must** be pointed at the system binary via an explicit
  executable path. CI (ubuntu runner) may use its own managed browser.
- **Existing test seams to reuse:** `NewTestDB()` and `setupIntegrationTest(t)`
  in `internal/handlers/integration_test.go` already build a seeded in-memory
  DB. The browser suite should reuse the same schema/seed logic, not invent a
  second fixture system.

---

## System 1 — Render contract tests (Go, no browser)

**What it is.** Extend the existing `httptest`-based Go tests to assert the
*structure* of every rendered page, plus crawl links and assets. Runs inside
`go test ./...` with no new runtime dependency and no server process.

**Why first.** It is hours of work, catches the majority of front-end
regressions (dead links, missing CSS, broken image paths, wrong error states),
and never flakes. It is the cheap floor that makes the browser layer optional
per-PR if the browser suite is ever flaky or slow.

### Test inventory

New file: `internal/handlers/render_contract_test.go`.

Per-page structural contracts (table-driven over a `map[string]struct{...}`):

- 200 status and `Content-Type: text/html`.
- Renders a full document: `<!DOCTYPE html>`, `<title>` non-empty, the CSS
  `<link href="/static/css/output.css">`, a `<header>` nav, a `<footer>`.
- Nav links resolve to real routes: `/`, `/gallery` (the `href="#"` "Estimate
  Costs" link is currently dead — assert it is either a real route or explicitly
  allowlisted as a known placeholder).
- Every `<img src>` is a same-origin `/static/...` path.

Zip-search form contract (`/`):

- Form `action="/builders"` `method="GET"`, one `name="zip"` input with
  `maxlength="5"` and `pattern` set, one submit button.

State contracts (`/builders?zip=`):

- Valid zip with results → result cards present, each links to `/builders/{id}`.
- Valid zip, no results (`len(results)==0`) → empty-state heading.
- Unknown zip (`store.ErrZipNotFound`) → "don't recognize" error state.
- Malformed zip (`zipPattern` fail) → validation error state.

Profile + not-found (`/builders/{id}`):

- Known id → 200, name renders, projects section present.
- Non-numeric id and unknown id → **404** status, `ProfileNotFound` body.

Error routes:

- Unmatched path (`/does-not-exist`) → 404 styled `ErrorPage`.
- `/healthz` → 200 `ok\n`.

Internal link crawler:

- For each page, extract all same-origin `<a href>`, issue a request, assert
  **no 404** (excluding explicitly known placeholders like `href="#"` and
  `javascript:history.back()`).

Static asset crawler:

- Extract every `/static/...` reference from all pages, request it against the
  `FileServer`, assert 200. This catches dead image paths across the
  `static/images/builders/*`, `static/images/gallery/*`, `static/images/logos/*`
  sets, which are numerous and currently unverified.

`/api/builders` JSON contract:

- Valid zip → 200, `application/json`, decodes to a slice, correct length,
  and **never `null`** (the handler guarantees a non-nil slice).
- Malformed/missing zip → **400** with `{"error":"invalid zip code"}`.
- Well-formed but unknown zip → **406** (`store.ToHTTPStatus(ErrZipNotFound)`)
  with `{"error":"zip not found"}`.

### Implementation notes

- Reuse `setupIntegrationTest(t)`/`NewTestDB()`; add only what is missing.
- Parse HTML with `golang.org/x/net/html` (add as a dependency) rather than
  regex, so the crawlers are robust. This is a Go stdlib-adjacent, CI-safe dep.
- For the link/asset crawler, use `httptest.NewServer` wrapping the real `mux`
  from `main` so routing, the `FileServer`, and middleware are exercised
  end-to-end in-process. Factor route registration out of `main()` into a
  `routes(...)` constructor so tests and the browser suite share one source of
  truth.

---

## System 2 — Browser E2E ("acts like a user")

**What it is.** A real Chromium drives a real server over HTTP and behaves like
a user: types into the search, submits, clicks through to a profile, toggles the
theme, hits error pages, and is checked for console errors, failed network
requests, and accessibility violations.

### Runner decision

**Recommended: `github.com/playwright-community/playwright-go` (Go).**

Rationale:
- Keeps the project **single-toolchain** — E2E is literally
  `go test -tags=e2e ./...`. One language, one dependency manager, one thing for
  AI agents to read and maintain. No second Node project drifting from the Go
  one.
- Capability that matters for this project: auto-waiting (kills flake), the
  **trace viewer**, screenshot **on failure**, and easy in-page `evaluate` for
  axe-core. These are exactly the failure artifacts an agent needs.
- Selectors support role/label/text and `data-testid`, satisfying the
  agent-maintainability requirement below.

Local vs CI browser resolution (a small config helper, `e2e/browser.go`):

```
if PLAYWRIGHT_CHROMIUM_EXECUTABLE is set → use it
else if /run/current-system/sw/bin/chromium exists (NixOS) → use it
else → let Playwright use its managed download (CI)
```

So: **local uses the system Chromium via `executablePath`; CI uses Playwright's
own managed browser.** Do not attempt to run Playwright's downloaded browser on
NixOS — it will not execute.

The Playwright Go driver is installed once per environment:
`go run github.com/playwright-community/playwright-go/cmd/playwright install
--with-deps` (CI) or `install chromium` (already have a binary locally; only the
driver is needed). Pin the version in `go.mod`.

**Fallback if the Playwright driver will not install under NixOS:
`github.com/go-rod/rod`.** Rod is pure-Go-over-CDP and launches the system
Chromium directly with no driver step, at the cost of hand-written waits and
assertions and losing the trace viewer. Decide this during Phase 1 spike; the
test bodies are nearly identical either way.

### How it runs

- **Build tag `e2e`** so `go test ./...` (the fast gate) does *not* run it.
  `go test -tags=e2e ./e2e/...` runs it. The normal `make test`/`make ci` stay
  fast; a dedicated `make e2e` target exists.
- **Server lifecycle in `TestMain`:** build the binary (or `go run`) into a temp
  path, create a fresh migrated + seeded SQLite DB in a temp dir, start the
  server on a **random free port** with `ADDR=:0`-style discovery (bind :0,
  read the port) or a chosen high port, wait for `/healthz` to return `ok`, run
  the suite, then tear down.
- **Deterministic data:** migrate + seed a fixed fixture (reuse migration seeds
  plus a small known builder set) so assertions reference stable names/zips.
  Never assert against the developer's live `traditionbuilders.db`.
- **Two viewports per key page:** desktop (1280×800) and mobile (375×667).

### Test inventory

`e2e/landing_test.go`
- Load `/`; assert title, hero heading, and nav present.
- Type `32801` into the zip field, submit → URL becomes `/builders?zip=32801`,
  results render.

`e2e/builders_test.go`
- Results page shows ≥1 card; each card's location/specialty renders.
- Click the first "View Profile →" → URL is `/builders/{id}`, and the profile
  name matches the card's name **that was clicked** (assert the specific link,
  not just "a profile loaded").
- Empty state: a seeded zip with no builders shows the empty-state copy.
- Unknown zip → the "don't recognize" error state.
- Malformed zip → validation error state.

`e2e/profile_test.go`
- Known id → name, contact buttons; `mailto:` href is well-formed.
- Unknown/non-numeric id → 404 page with "couldn't find that builder".
- "Search Other Builders" / back link works.

`e2e/theme_test.go`
- Toggle theme → `<html>` gains `dark`; `localStorage.theme === 'dark'`; reload
  persists the choice; toggle again → light.

`e2e/errors_test.go`
- `/does-not-exist` → 404 styled page, "Return Home" link works.

`e2e/a11y_test.go`
- Inject `axe-core` (`evaluate` against the vendored `axe.min.js` from
  `node_modules` or a pinned copy) and run on `/`, `/gallery`, `/builders`,
  `/builders/{id}`, 404, and 500. Fail on **serious/critical** violations;
  report moderate as a warning artifact. Tune to avoid known-acceptable noise.

`e2e/hygiene_test.go`
- Attach console and `requestfailed`/`response` listeners on every navigation.
  Fail the test if the page logs a console **error**, or if any required
  same-origin asset returns ≥400. (The htmx CDN script is external — decide
  whether external asset failures are fatal; recommend warning-only.)

### Failure artifacts (the part that makes this agent-friendly)

On any test failure, capture and write to `e2e/artifacts/<test-name>/`:
- full-page screenshot (`-full-page`),
- Playwright trace (if available),
- the browser console log,
- the failing page's DOM (`page.content()`),
- the server's recent stdout/stderr.

CI uploads `e2e/artifacts/` as a workflow artifact on failure only. On success,
upload a **screenshot of every page** as a separate artifact so a human can
review the front-end from the PR without running anything.

---

## CI integration

Add a job to `.github/workflows/ci.yml` (or a sibling `e2e.yml`):

```
e2e:
  needs: [build]              # or build inline
  steps:
    - checkout
    - setup-go (+ cache)
    - install templ; templ generate
    - install tailwindcss; make css
    - make build
    - install Playwright driver (managed browser) — CI uses its own browser
    - make e2e                  # go test -tags=e2e ./e2e/...
    - on failure: upload e2e/artifacts
    - always:      upload page screenshots (success artifact)
```

- Make the `e2e` job a **required status check** alongside `lint`/`test`/`build`
  (branch protection, same manual step already noted in `PLAN_CI_CD.md`).
- Keep it a **separate job** from `make ci` so a browser outage cannot mask the
  fast unit-test result, and so the two have independent retry/flake history.
- `make e2e` is the single local/CI entry point, mirroring the `make ci`
  philosophy from `PLAN_CI_CD.md`.

---

## Agent-maintainability conventions (build these in from day one)

1. **Selector standard.** Prefer, in order: `data-testid` on interactive
   elements → semantic locators (role/label/placeholder/text) → last resort
   stable CSS. **Never** select by Tailwind class chain
   (`div.bg-surface > .flex ...`) — it is exactly what restyling breaks and what
   makes agent edits dangerous.
   - Add `data-testid` now to: zip input (`zip-search`), submit button, each
     builder card (`builder-card`), the profile link (`view-profile`), theme
     toggle (`theme-toggle`), and the error/empty-state containers.
2. **Page objects** (`e2e/pages/landing.go`, `builders.go`, `profile.go`):
   one small struct per page exposing `Load()`, `SearchZip(zip)`, etc. Test
   bodies read as user intent; selector changes touch one file.
3. **One assertion vocabulary.** Wrap common checks (`ExpectStatus`,
   `ExpectText`, `ExpectVisible`) so failures read consistently and agents can
   pattern-match.
4. **No sleeps.** Use auto-waiting / `Expect`. A retry helper is allowed only
   around known-async boundaries, never as a fix for a real race.
5. **Every failure names the route and the user step** in its message, e.g.
   `GET /builders?zip=00000: expected empty-state copy`.

---

## Explicitly out of scope (deferred)

- **Visual regression / pixel diffing.** Valuable, but flaky and the branding
  work (leather/cream palette) is still moving. Revisit after the palette
  settles, introduced as a **non-blocking** job with an explicit
  "accept new baseline" workflow.
- **Lighthouse / performance budgets.** Premature at MVP size.
- **htmx interaction tests.** No `hx-*` attributes exist yet. Both systems
  already support them: System 1 can assert fragment responses by sending
  `HX-Request: true`; System 2 needs no change. Add htmx cases when the first
  htmx feature lands.

---

## File inventory (planned)

| File | Change |
|---|---|
| `PLAN_FRONTEND_TESTING.md` | this plan |
| `cmd/server/main.go` | factor route registration into a shared `routes()` constructor |
| `go.mod` | add `golang.org/x/net/html`, `playwright-go` (or `go-rod`) |
| `internal/handlers/render_contract_test.go` | new — System 1 |
| `internal/handlers/crawl_test.go` | new — link + static asset crawler |
| `internal/handlers/api_contract_test.go` | new — `/api/builders` JSON contract |
| `templates/*.templ` | add `data-testid` to interactive elements |
| `e2e/` | new package — `browser.go`, `TestMain`, page objects, suites |
| `Makefile` | add `e2e` target; keep `ci`/`test` browser-free |
| `.github/workflows/ci.yml` (or `e2e.yml`) | new `e2e` job + artifact upload |
| `.gitignore` | add `e2e/artifacts/` |

---

## Phasing

**Phase 1 — System 1 (this week, low risk).**
Route constructor refactor; render-contract + crawler tests; add `data-testid`s;
wire into the existing `go test` gate. Deliverable: `go test ./...` proves every
route renders, every link resolves, every asset loads, every error state shows.

**Phase 2 — System 2 thin slice.**
Spike the runner choice (Playwright-go vs go-rod) on NixOS against system
Chromium. Implement `TestMain` + landing → search → profile → theme-toggle +
a11y on one page. Add the separate CI `e2e` job with failure artifacts.

**Phase 3 — System 2 breadth + hardening.**
Full inventory above; mobile viewport; console/network hygiene; success
screenshot artifacts; make `e2e` a required check.

**Phase 4 — revisit deferred items** once branding and the first htmx feature
land.

---

## Verification

- `go test ./...` passes with System 1 included (no browser needed).
- `make e2e` passes locally against system Chromium on NixOS.
- The `e2e` CI job passes on ubuntu with its managed browser, and its failure
  path uploads artifacts.
- A deliberately broken page (e.g. rename the CSS link, drop an image, change a
  route) is caught by the appropriate system — verified by temporary fault
  injection, then reverted.

## Open decisions

1. Runner: **Playwright-go (recommended)** vs go-rod — resolve in the Phase 2
   spike based on whether the driver installs cleanly on NixOS.
2. Accessibility gate: fail on serious+critical only (recommended) vs report-only
   at first.
3. External asset failures (htmx CDN): fatal vs warning (recommend warning).
4. The dead `href="#"` "Estimate Costs" link: give it a real route, or
   allowlist it as a known placeholder in the link crawler.
