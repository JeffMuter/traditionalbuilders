#!/usr/bin/env bash
# Load the full GeoNames US postal code dataset into zip_codes.
#
# Wraps cmd/seed-zips, which downloads US.zip (~3 MB) from geonames.org and
# bulk-imports ~41k rows inside a single transaction, replacing the small
# bootstrap seed from migration 003.
#
# Usage:
#   load_zipcodes.sh [path/to/traditionbuilders.db]
set -euo pipefail

DB_PATH="${1:-traditionbuilders.db}"

if [ ! -f "$DB_PATH" ]; then
    echo "error: $DB_PATH not found — run 'make migrate' first." >&2
    exit 1
fi

echo "Backing up $DB_PATH -> ${DB_PATH}.bak"
cp "$DB_PATH" "${DB_PATH}.bak"

echo "Importing GeoNames data (existing zip_codes rows are replaced)…"
go run ./cmd/seed-zips "$DB_PATH"

COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM zip_codes;")
echo "Integrity check: zip_codes now holds $COUNT rows."
if [ "$COUNT" -lt 40000 ]; then
    echo "warning: expected ~41k rows; import may be incomplete." >&2
    echo "restore the previous data with: mv ${DB_PATH}.bak $DB_PATH" >&2
    exit 1
fi

echo "Done."
