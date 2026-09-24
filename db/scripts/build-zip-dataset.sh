#!/usr/bin/env bash
# Build the vendored zip code dataset from GeoNames.
#
# This is a MANUAL, periodic maintenance step — it is deliberately NOT part of
# any Make target that CI, `setup`, `release`, or `deploy` runs. Run it by hand
# when you want to refresh the data, review the diff, and commit it.
#
# It downloads the GeoNames US postal-code dump, extracts the tab-separated
# US.txt, filters to 5-digit numeric ZIPs, sorts by ZIP, gzips the result, and
# writes internal/zipdata/dataset.json with the provenance/checksum metadata that the
# server embeds and records in the data_seeds table.
#
# Output (both committed), colocated with the Go package that embeds them:
#   internal/zipdata/zip_codes.tsv.gz   zip, lat, lng, city, state  (no header)
#   internal/zipdata/dataset.json       source_url, fetched_at, license, sha256, row_count
#
# Usage:
#   build-zip-dataset.sh [--force]
#
# Without --force the script refuses to overwrite an existing dataset.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="$ROOT/internal/zipdata"
TSV_GZ="$OUT_DIR/zip_codes.tsv.gz"
META="$OUT_DIR/dataset.json"

GEONAMES_URL="https://download.geonames.org/export/zip/US.zip"

force=0
if [ "${1:-}" = "--force" ]; then
    force=1
fi

if [ -f "$TSV_GZ" ] && [ "$force" -ne 1 ]; then
    echo "error: $TSV_GZ already exists; pass --force to rebuild." >&2
    exit 1
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

echo "==> Downloading $GEONAMES_URL"
curl -fsSL -o "$work/US.zip" "$GEONAMES_URL"

echo "==> Extracting US.txt"
unzip -o -q "$work/US.zip" US.txt -d "$work"

echo "==> Filtering, validating, deduping, and sorting"
# GeoNames columns (1-indexed): 2=postal_code 3=place_name 5=state 10=lat 11=lng
# Keep only 5-digit numeric ZIPs with a non-empty city/state and in-range
# coordinates, then sort by ZIP (LC_ALL=C = deterministic across machines) and
# drop duplicate ZIPs, keeping the first. This mirrors internal/zipdata's
# integrity rules so the embedded count is exact.
awk -F'\t' '
    length($2) == 5 && $2 ~ /^[0-9]{5}$/ && $3 != "" && $5 != "" {
        lat = $10 + 0; lng = $11 + 0
        if (lat >= 15 && lat <= 72 && lng >= -180 && lng <= -60) {
            printf "%s\t%s\t%s\t%s\t%s\n", $2, $10, $11, $3, $5
        }
    }
' "$work/US.txt" | LC_ALL=C sort -t "$(printf '\t')" -k1,1 \
    | awk -F'\t' '!seen[$1]++' > "$work/zip_codes.tsv"

row_count="$(wc -l < "$work/zip_codes.tsv" | tr -d ' ')"
if [ "$row_count" -lt 40000 ]; then
    echo "error: only $row_count rows; expected ~41k. Refusing to write." >&2
    exit 1
fi

echo "==> Compressing ($row_count rows)"
gzip -9 -c "$work/zip_codes.tsv" > "$TSV_GZ"

sha="$(sha256sum "$TSV_GZ" | awk '{print $1}')"
fetched_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
# NOT a public-domain dataset: GeoNames postal codes are CC BY 4.0 and require
# attribution. Keep this string in sync with the footer link.
license="CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/)"

cat > "$META" <<JSON
{
  "name": "zip_codes",
  "source": "GeoNames",
  "source_url": "$GEONAMES_URL",
  "license": "$license",
  "attribution": "Postal code data © GeoNames (https://www.geonames.org/), CC BY 4.0",
  "fetched_at": "$fetched_at",
  "row_count": $row_count,
  "columns": ["zip", "lat", "lng", "city", "state"],
  "sha256": "$sha"
}
JSON

echo "==> Wrote $TSV_GZ ($(du -h "$TSV_GZ" | cut -f1), $row_count rows)"
echo "==> Wrote $META (sha256 $sha)"
echo "Review 'git diff' and commit both files."
