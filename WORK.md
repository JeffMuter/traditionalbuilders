# WORK.md

## Completed: Zip Code Builder Search

**Shipped.** The landing page zip form, `/builders` results page, `/api/builders`
JSON endpoint, haversine proximity sort, and all supporting migrations are live.

### What was built

| Area | File(s) |
|---|---|
| Domain types | `internal/models/models.go` |
| DB query layer | `internal/store/store.go`, `store/distance.go` |
| JSON API | `internal/handlers/zipcodes.go` → `GET /api/builders?zip=XXXXX` |
| HTML results page | `internal/handlers/builders.go` → `GET /builders?zip=XXXXX` |
| Landing form | `templates/landing.templ` — replaced city-autocomplete with zip form |
| Results template | `templates/builders.templ` |
| Migrations | `002_add_coords.sql`, `003_zip_coords.sql` (113-row seed), `004_seed_professionals.sql` (8 builders) |
| Full zip dataset | `cmd/seed-zips/main.go` — `make seed-zip-codes` imports ~41k GeoNames rows |

### Decisions made

- **Zip data source:** GeoNames public domain CSV, bulk-imported in a single
  transaction by `cmd/seed-zips`. Migration 003 keeps a 113-row bootstrap seed so
  the feature works immediately after `make setup` without a network call.
- **Result limit:** Top 10 closest, no distance cap.
- **Type ownership:** `models.BuilderResult` lives in `internal/models` — not in
  `handlers` or `templates` — to avoid cross-package coupling.
- **Query ownership:** All SQL lives in `internal/store`. Handlers call store
  methods; they never write raw SQL.

---

## Completed: Builder Profile Pages

**Shipped.** `GET /builders/{id}` renders a full profile; every result card on
the search page now links to it.

### What was built

| Area | File(s) |
|---|---|
| Domain types | `internal/models/models.go` — `Professional`, `Project` |
| DB query layer | `internal/store/professional.go` — `GetProfessional`, `ListProjects` |
| Handler | `internal/handlers/profile.go` → `GET /builders/{id}` |
| Profile template | `templates/profile.templ` — profile body, 404 body, project card |
| Shared chrome | `templates/shell.templ` — head/nav/footer/theme-toggle wrapper |
| Template helpers | `templates/format.go` — `formatUSD`, `completedYear`, `displayLocation`, `telHref` |
| Card link | `templates/builders.templ` — dead `href="#"` now points at the profile |

Shows name, specialty, location, bio, mailto/tel contact buttons, and past
projects (title, location, completion year, cost) with an empty state when a
builder has none.

### Decisions made

- **Routing:** Go 1.22+ `ServeMux` wildcards (`/builders/{id}` + `r.PathValue`),
  so no third-party router. `/builders` and `/builders/{id}` coexist cleanly.
- **404 handling:** a non-numeric id and an unknown id render the same
  `ProfileNotFound` page with a 404 status — no distinction leaked to the user.
- **Projects are non-fatal:** if `ListProjects` errors, the profile still renders
  with an empty project list rather than 500-ing the whole page.
- **Shared layout:** `profileShell` extracts the site chrome that
  `landing`/`gallery`/`builders` each duplicate. New pages should use it;
  the three existing templates have **not** been migrated yet.
- **Location display:** canonical `city, state` from the `zip_codes` join wins
  over the free-text `professionals.location` column, which falls back in.

---

## Up Next

Priority order based on what a real user would hit first.

### 1 — Builder registration / onboarding form
Currently all professionals are seed data. Need a way to add real ones.

- `GET /join` — onboarding form (name, email, phone, specialty, bio, zip)
- `POST /join` — validate, look up lat/lng from zip_codes table, INSERT into professionals
- Confirmation page or redirect to their new profile

### 2 — Radius filter
Some zip codes have zero builders; wide searches are currently all-or-nothing.

- Add `radius` query param (default 100 miles, options: 25 / 50 / 100 / 250)
- Filter `FindBuildersNear` results before returning
- UI: radio/select on the landing form or the results page

### 3 — Pagination
10-result cap is fine for MVP. When we have real data this needs pages.

- Add `page` query param (default 1)
- `FindBuildersNear` gains `limit int, offset int` params
- Results page shows "Next / Previous" links

### 4 — Map view
Low priority until profile pages exist and there are enough real builders to make
a map meaningful.

- Leaflet.js (no API key required)
- Pin each result card on a map centered on the searched zip
- Clicking a pin opens the profile or highlights the card

### 5 — Search persistence
Currently the landing page form clears after submit. Could pre-fill the zip from
the URL param when navigating back from the results page (small UX detail).

### 6 — Real project data
The `projects` table has schema but no seed rows, so every profile shows the
empty state. Needs seed data (migration 005) or the registration flow above.

### 7 — Migrate remaining templates to `profileShell`
`landing.templ`, `gallery.templ`, and `builders.templ` each carry their own copy
of the head/nav/footer/theme-toggle. Rename `profileShell` → `shell` and fold
them in; a nav change currently means four edits.
