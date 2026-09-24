# Stale Migration 005 Targets Non-Existent Professional IDs

## Problem Statement

`db/migrations/005_add_image_path.sql` adds an `image_path` column and then sets
images for professionals with **IDs 9–17**:

```sql
UPDATE professionals SET image_path = 'builders/timber-frame-01.jpg' WHERE id = 9;
...
UPDATE professionals SET image_path = 'builders/colonial-church.jpg' WHERE id = 17;
```

But migration `004_seed_professionals.sql` inserts only **8** professionals
(IDs 1–8). The referenced 9 new timber-frame builders do not exist in any
migration or seed in the repo.

### Verified state

```
$ sqlite3 traditionbuilders.db "SELECT id,name,image_path FROM professionals WHERE image_path IS NOT NULL;"
(no rows)
```

- The column is added successfully.
- **All UPDATEs affect zero rows.** No builder has an image.
- `provider`/`verified_at` (migration 006) are similarly present-but-default.

### Impact

- Image/icon feature is effectively unbuilt, despite appearing done in the
  migration history.
- Any assumption that professionals have images (profile rendering, directory
  icons — see `issues/builder-visual-verification.md`) is currently false.
- Migration history reads as misleading; a future developer may assume images
  are populated.

## Requirements

- Determine the intended source of the 9 builders and their images:
  - Were the builders meant to be seeded (missing `INSERT`s), or
  - Is migration 005 a leftover from an earlier, larger dataset?
- Either:
  - Restore/replace the missing seed so the IDs referenced by 005 exist and
    images are assigned, or
  - Remove the dead UPDATE statements and leave image assignment to the
    builder onboarding path.
- Do not rewrite an already-applied migration's meaning silently; add a new
  migration or clearly comment the intent.

## Acceptance criteria

- No migration references IDs that cannot exist on a fresh database.
- On a fresh `make reset-db`, professionals that should have images actually do
  (or the image path is explicitly deferred).
- `image_path` semantics documented (what stores the files, expected layout
  under `static/images/`).

## Related

- `issues/builder-visual-verification.md` — icon/image feature requirements.
- `issues/logo-permissions-outreach.md` — image/permission constraints.
