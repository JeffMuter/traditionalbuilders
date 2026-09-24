# Committed Database Backup (`traditionbuilders.db.bak`)

## Problem Statement

`traditionbuilders.db.bak` is **tracked in Git**. It is a copy of the SQLite
database created by `db/scripts/load_zipcodes.sh` before importing GeoNames
data. It contains the full `professionals` table, including names, emails, and
phone numbers.

```
$ git ls-files | grep .bak
traditionbuilders.db.bak
```

`.gitignore` excludes `*.db`, `*.db-shm`, `*.db-wal` — but **not `*.bak`**, so
the backup slipped into version control.

### Impact

- Real-looking contact data (name / email / phone) lives in repo history.
- Repo bloat (~53 KB file, plus history).
- Data-hygiene / privacy concern if the repo is ever shared or made public.
- Not shipped in the release tarball, so **not a prod-runtime risk** — this is a
  source-control hygiene issue, not a serving issue.

## Requirements

1. Add `*.bak` (or `*.db.bak`) to `.gitignore`.
2. Untrack the file: `git rm --cached traditionbuilders.db.bak`.
3. Purge it from history so the data is genuinely gone:
   - `git filter-repo --path traditionbuilders.db.bak --invert-paths`
     (or BFG). Note this rewrites history — coordinate with anyone with clones.
4. Confirm `load_zipcodes.sh` still functions and its backup is ignored going
   forward.

## Acceptance criteria

- `git ls-files` contains no `.bak` files. ✅
- The file is absent from history (`git log --all -- traditionbuilders.db.bak`
  is empty). ❌ still present (see below)
- `.gitignore` covers the backup name pattern. ✅

## Resolution status (partial)

Done in commit `0971a88`:

1. ✅ `.gitignore` now ignores `*.bak` / `*.db.bak`.
2. ✅ `traditionbuilders.db.bak` untracked and deleted from the working tree.
4. ✅ `load_zipcodes.sh` was retired entirely in favour of the offline
   `internal/zipdata` loader, so no future script can recreate this artifact.

Still pending:

3. ❌ **History purge not performed.** The file is still reachable at commit
   `ec48486` (`git log --all -- traditionbuilders.db.bak` is non-empty).
   This requires a destructive history rewrite
   (`git filter-repo --path traditionbuilders.db.bak --invert-paths`)
   and force-push, which must be coordinated with any existing clones.
   Until then the contact data remains in history: **do not make the repo
   public**.

## Notes

- The DB backup is a **local safety artifact**; it should never be committed.
  Consider pointing the script's backup at a gitignored path (e.g.
  `backups/`).
