# Plan: Ship the Full Zip Code Dataset With the Release

Resolves `issues/zip-dataset-not-shipped.md`.

**Status:** proposed
**Owner:** —
**Related issues:** `committed-db-backup.md` (the `.bak` this flow created),
`stale-migration-005.md`, `logo-permissions-outreach.md` (licensing precedent).

---

## 1. Problem (verified)

| Fact | Evidence |
|---|---|
| Prod `zip_codes` gets **113 rows**, not ~41k | `db/migrations/003_zip_coords.sql` seeds 113; `make release` copies only `db/migrations/` |
| Dev DB has **40,999 rows** from an out-of-band step | `make seed-zip-codes` → `db/scripts/load_zipcodes.sh` → `go run ./cmd/seed-zips` (downloads ~3 MB from `download.geonames.org`) |
| Dev-only, network-dependent, not in deploy path | Not referenced by `setup`, `release`, `deploy.sh`, `Dockerfile`, or the systemd unit |
| Silent prod failure | `/api/builders?zip=…` returns `ErrZipNotFound` → "couldn't recognize" for most of the US |

The GeoNames postal dump is now **CC BY 4.0** (not public domain), so a site
attribution link is required. The comment in `003_zip_coords.sql` is wrong and
must be fixed.

### Dataset facts (measured 2026-09-24)

- `US.zip` = 634,329 bytes; contains `US.txt` (41,490 lines) + `readme.txt`.
- Filtered + normalized: 5-digit numeric ZIPs, non-empty city/state, in-range
  coords, deduped → **40,976 rows**, ~1.4 MB TSV, **508 KB gzip**.
- The existing `cmd/seed-zips` integrity bounds give **41,001** (drops ~489
  out-of-range). The vendored file applies stricter normalization (empty
  state, duplicates) for a clean **40,976**; keep that validation at load time
  so the filter can evolve without re-downloading.

## 2. Decision

**Vendor a filtered, gzipped TSV in the repo; `go:embed` it into the server
binary; apply it through one versioned, transactional seed path used by dev,
CI, and prod.**

Why this option over the alternatives in the issue:

- **vs. 41k-row SQL migration** — a 2+ MB SQL file is rewritten by every
  `goose` migration, shipped twice (once in `db/migrations`, once as a data
  file), and awkward to refresh. Embedding one gzip keeps the repo diff tiny
  and lets the loader validate/replace atomically.
- **vs. fetch-at-deploy** — requires network + trust in a third party on every
  deploy; violates the "offline, reproducible" requirement and re-opens the
  `localhost download` failure mode.
- Embedding also fixes the Docker image and systemd deploy for free: the data
  travels inside the binary the tarball already ships.

**Consequence:** `db/migrations` no longer needs to seed 41k rows. Migration
`003` keeps only the schema; the 113 bootstrap rows are removed once the seed
runs everywhere (Phase 5).

## 3. Architecture

```
internal/zipdata/zip_codes.tsv.gz   vendored data (508 KB, checked in)
internal/zipdata/dataset.json       {source_url, fetched_at, license, sha256, row_count}
db/scripts/build-zip-dataset.sh    regenerates the above from GeoNames; manual/periodic
internal/zipdata/                     //go:embed zip_codes.tsv.gz
    zipdata.go                        Rows() iterator + Manifest() + EnsureLoaded()
    zipdata_test.go                   parse/count/bounds/idempotency tests
cmd/seed-zips/main.go               CLI: load embedded data into a DB path
cmd/server/main.go                  -seed-only flag; EnsureLoaded() on startup
db/migrations/007_data_seeds.sql    data_seeds bookkeeping table
```

Seed semantics — `zipdata.EnsureLoaded(db)`:

1. Read `data_seeds` for a row matching `name='zip_codes'` and `sha256`.
2. Up-to-date → no-op (return immediately). This is what makes the systemd
   `Restart=on-failure` cheap and safe.
3. Missing/stale → single transaction: `DELETE FROM zip_codes`, bulk `INSERT`
   of the streamed embedded rows, then upsert the `data_seeds` row.
4. If `zip_codes` doesn't exist yet (migrations not run), log a warning and
   skip — the server must not crash on an unmigrated DB.

`data_seeds(name TEXT PRIMARY KEY, sha256 TEXT NOT NULL, source TEXT NOT NULL,
row_count INTEGER NOT NULL, loaded_at DATETIME DEFAULT CURRENT_TIMESTAMP)`.

## 4. Phases

### Phase 0 — Attribution & doc corrections
- Fix the false "public domain" comment in `003_zip_coords.sql`; state CC BY 4.0.
- Add GeoNames attribution + link (`https://www.geonames.org/`) to the site
  footer (`templates/shell.templ`) and to `README.md`.
- Record the license/source in `internal/zipdata/dataset.json` and a short
  `internal/zipdata/README.md`.

### Phase 1 — Vendor the dataset (offline from here on)
- Add `db/scripts/build-zip-dataset.sh`:
  download `US.zip`, extract `US.txt`, filter with
  `awk -F'\t' 'length($2)==5 && $2 ~ /^[0-9]{5}$/ {print ...}'` (plus
  non-empty city/state, in-range coords, dedupe), `LC_ALL=C sort` by ZIP,
  `gzip -9`, write `dataset.json` with sha256 + row count.
  Refuse to overwrite unless `--force`; never touch the DB.
- Commit `internal/zipdata/zip_codes.tsv.gz` + `dataset.json`.
- Regeneration is a deliberate, reviewed commit — **not** part of any Make
  target that CI or deploy runs.

### Phase 2 — `internal/zipdata`
- `//go:embed zip_codes.tsv.gz`; no network code.
- `Rows() (iter.Seq2[Row, error])` or a callback/`Each` that streams from the
  gzip reader — never materialize all 41k rows on the startup fast path.
- `Row {Zip string; Lat, Lng float64; City, State string}`; reuse the existing
  integrity bounds (`15..72`, `-180..-60`, 5-digit) so behavior matches
  `cmd/seed-zips`.
- Tests: row count in expected band, ZIP uniqueness, bounds, malformed-row
  skip behavior, embedded sha matches `dataset.json`.

### Phase 3 — CLI + server integration
- Rewrite `cmd/seed-zips/main.go` to call `zipdata.EnsureLoaded` (accept DB
  path arg, default `traditionbuilders.db`). Remove `net/http`, `archive/zip`,
  `bytes`, and the `geonamesURL` constant.
- `cmd/server/main.go`: add `-seed-only` flag → open DB, `EnsureLoaded`, exit.
  On normal startup, call `EnsureLoaded` in a goroutine (or before
  `ListenAndServe`) and log the result; never `log.Fatal` the web server over a
  seed problem.
- `go run ./cmd/seed-zips` becomes fully offline.

### Phase 4 — Migrations
- Add `007_data_seeds.sql` (`+goose Up` / `Down` drop).
- `003_zip_coords.sql`: keep `CREATE TABLE`; delete the 113-row `INSERT`
  block and the stale comment (full dataset is the source of truth).
- Update `internal/store/migration_test.go`: its `TestMigration006_*` apply all
  migrations and currently don't assert the seed; add 007 coverage (table +
  down). It must not assume 113 rows anywhere.

### Phase 5 — Release / deploy / Docker wiring
- `Makefile`
  - `seed-zip-codes`: `go run ./cmd/seed-zips` (offline; drop the shell script
    indirection).
  - `reset-db`: after `goose … up`, run `go run ./cmd/seed-zips` so a clean dev
    DB gets the full set, matching prod.
  - `release`: unchanged for data (embedding handles it); still verify the
    binary starts and reports row_count in a smoke test.
- `deploy/deploy.sh`: after installing the new binary and running
  `goose -dir migrations … up`, run `./server -seed-only` as the
  `traditionalbuilders` user, and fail the deploy if it exits non-zero.
  Document the manual runbook in the script header/README.
- `Dockerfile`: nothing to copy (embedded); optionally add a `docker-entrypoint`
  note that `-seed-only` should run after migrations.
- Retire `db/scripts/load_zipcodes.sh` (its `.bak` copy is the cause of
  `committed-db-backup.md`). New flow must not create `*.bak` files.
- `scripts/seed-zip-codes`, `scripts/reset-db`: keep as thin wrappers or update
  to the new command.

### Phase 6 — Tests (the acceptance gate)
- **Unit** (`internal/zipdata`, `cmd/seed-zips`): parse, count, idempotency
  (second `EnsureLoaded` is a no-op and leaves `data_seeds` untouched).
- **Fresh-DB integration** (new `-tags=e2e`-free test, or extend `e2e`):
  create a temp DB, run all migrations via `goose.Up`, call `EnsureLoaded`,
  assert `COUNT(*) BETWEEN 41000 AND 41490` and that a non-bootstrap ZIP
  (e.g. `90210`) resolves through `FindBuildersNear` (empty result, not
  `ErrZipNotFound`).
- **E2E**: existing `90001` empty-state / `99999` unknown-zip assertions still
  hold against the full dataset (`90001` exists; `99999` does not).
- Add a startup smoke assertion that `-seed-only` on an already-seeded DB
  exits 0 quickly.

### Phase 7 — Cleanup & docs
- Update `AGENTS.md` schema/commands (replace the "manual GeoNames" note),
  `README.md`, and the deploy runbook.
- Remove/download references to `download.geonames.org` from shipped code.
- Cross-link resolution in `issues/zip-dataset-not-shipped.md`; delete the
  tracked `traditionbuilders.db.bak` per `committed-db-backup.md` (same flow).

## 5. Acceptance mapping

| Issue criterion | How this plan satisfies it |
|---|---|
| Fresh `reset-db` + deploy → ~40k rows | Phase 5 `reset-db` + `deploy.sh -seed-only`; Phase 6 asserts 40k–41.5k, actual 40,976 |
| Documented, repeatable prod procedure | `deploy.sh` runs `server -seed-only` after migrations; runbook in README |
| Zip outside the original 113 works on deployed instance | Embedded full set; Phase 6 test uses `90210` |
| Offline / reproducible | Data vendored + `go:embed`; no network in seed path |
| Remove misleading 113-row seed | Phase 4 strips `003`'s INSERT |
| Document monthly refresh | `build-zip-dataset.sh` + `internal/zipdata/README.md` |

## 6. Open questions resolved

- **Vendor vs. fetch at deploy** → vendored filtered TSV.gz, embedded.
- **Licensing** → CC BY 4.0: footer link + `dataset.json` + `README`
  attribution (Phase 0). No permission request needed (unlike logos).
- **Refresh cadence** → manual, on demand (`build-zip-dataset.sh --force`);
  `sha256` in `data_seeds` makes a stale-DB deploy self-heal.

## 7. Risks

- **Binary size** +~523 KB; negligible.
- **Startup latency** on first run after deploy: ~41k inserts in one
  transaction (~1–2 s observed for the existing importer). With `Type=simple`
  and no `TimeoutStartSec` this is safe; `Restart=on-failure` only re-runs the
  seed if the first attempt fails. The `data_seeds` short-circuit means every
  later restart is a single `SELECT`.
- **ZIP+4 / non-numeric ZIPs** are still excluded (unchanged behavior);
  `zip_codes.zip` is the 5-digit base and callers normalize.
- **`003` INSERT removal** requires the seed to run before anyone queries;
  `reset-db` and the deploy runbook both enforce this ordering.

## 8. Out of scope

- Changing proximity search, the 500-mile prefilter, or `maxResults`.
- The `projects` table, estimation page, builder icons, and logo permissions
  (separate issues).
