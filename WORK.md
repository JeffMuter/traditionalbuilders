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

## Up Next

Priority order based on what a real user would hit first.

### 1 — Builder profile pages
Each builder card has a placeholder "View Profile →" link. These need real pages.

- Route: `GET /builders/{id}` (or `/builders/profile?id=X`)
- Template: `templates/profile.templ`
- Handler: `internal/handlers/profile.go`
- Store method: `store.GetProfessional(ctx, id) (models.Professional, error)`
- Show: name, specialty, bio, phone, email (mailto link), location, past projects

### 2 — Builder registration / onboarding form
Currently all professionals are seed data. Need a way to add real ones.

- `GET /join` — onboarding form (name, email, phone, specialty, bio, zip)
- `POST /join` — validate, look up lat/lng from zip_codes table, INSERT into professionals
- Confirmation page or redirect to their new profile

### 3 — Radius filter
Some zip codes have zero builders; wide searches are currently all-or-nothing.

- Add `radius` query param (default 100 miles, options: 25 / 50 / 100 / 250)
- Filter `FindBuildersNear` results before returning
- UI: radio/select on the landing form or the results page

### 4 — Pagination
10-result cap is fine for MVP. When we have real data this needs pages.

- Add `page` query param (default 1)
- `FindBuildersNear` gains `limit int, offset int` params
- Results page shows "Next / Previous" links

### 5 — Map view
Low priority until profile pages exist and there are enough real builders to make
a map meaningful.

- Leaflet.js (no API key required)
- Pin each result card on a map centered on the searched zip
- Clicking a pin opens the profile or highlights the card

### 6 — Search persistence
Currently the landing page form clears after submit. Could pre-fill the zip from
the URL param when navigating back from the results page (small UX detail).
