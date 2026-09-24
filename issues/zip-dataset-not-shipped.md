# Zip Code Dataset Is Not Part of the Release (Prod Search Will Be Broken)

> **RESOLVED** — see `PLAN_ZIP_DATASET.md` and `internal/zipdata/`. The full
> ~40,976-row US dataset is now vendored in `internal/zipdata/` (gzipped TSV),
> embedded into the server binary via `go:embed`, and applied by
> `internal/zipdata.EnsureLoaded` (`make seed-zip-codes`, `server -seed-only`,
> and best-effort on startup). Bookkeeping lives in `data_seeds` (migration
> `007_data_seeds.sql`), keyed by dataset SHA-256 for idempotent self-healing
> redeploys. Migration `003` no longer hard-codes 113 rows. `deploy.sh` runs
> `server -seed-only` after migrations. GeoNames is CC BY 4.0 (attribution in
> the site footer and `internal/zipdata/README.md`). Verified end-to-end on a
> clean DB: `/api/builders?zip=90210` resolves, `99999` → 406.

## Problem Statement

Proximity/directory search depends on the `zip_codes` table. The development
database holds **40,999 rows**, but that data is **not produced by the
migrations that ship in the release**. It comes from a manual, out-of-band step:

```bash
make seed-zip-codes      # -> db/scripts/load_zipcodes.sh
                         # -> go run ./cmd/seed-zips  (downloads GeoNames US.zip)
```

The release target (`make release`) copies only `db/migrations/` into the
tarball. It does **not** copy a database and does **not** run the zip import.

Migration `003_zip_coords.sql` seeds only **113 hard-coded zip codes** (the
original bootstrap set for the demo professionals).

### Consequence

A production deploy that runs migrations will get **113 zip codes, not
~41,000**. Directory/proximity search will fail for the vast majority of the
country, while the demo data looks complete. This is a silent, hard-to-catch
launch failure.

## Cause

`cmd/seed-zips` downloads dataset (~3 MB) from `download.geonames.org` at
runtime. That makes it unsuitable to run during a normal deploy (network
dependency, no offline reproducibility) and it was never wired into `setup`,
`release`, or the service unit.

## Requirements

- Decide and document how production gets the full zip dataset. Options:
  1. **Bake the dataset into the release** — ship a pre-generated migration or
     a data file inside the tarball, applied at deploy. (Offline, reproducible.)
  2. **Run the import during deploy** — wire `seed-zip-codes` into the deploy
     runbook/service, accepting the network dependency.
  3. **Generate a migration from the CSV** — commit a `007_zip_codes_full.sql`
     (large, but deterministic).
- Whichever path is chosen, it must be **tested end-to-end on a clean database**
  (not the current dev DB, which already has the rows).
- Remove/retire the misleading 113-row bootstrap seed once the full dataset is
  the source of truth.
- Document the monthly refresh workflow (see `scripts/seed-zip-codes`) for prod.

## Acceptance criteria

- A fresh `make reset-db` + deploy yields `zip_codes` with ~40k rows in prod.
- A documented, repeatable prod procedure exists (not a dev-only command).
- Search for a zip outside the original 113 works on the deployed instance.

## Open questions

- Do we want the GeoNames data vendored into the repo, or fetched at deploy?
- Licensing/attribution for GeoNames (public domain, but confirm attribution
  expectations) — see `issues/logo-permissions-outreach.md` for precedent.
