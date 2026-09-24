# Zip code dataset

`zip_codes.tsv.gz` is the vendored GeoNames US postal-code dataset (filtered to
5-digit numeric ZIPs) that the server embeds into its binary via `go:embed` in
`internal/zipdata`. `dataset.json` records its provenance and checksum.

## License

GeoNames postal-code data is licensed **CC BY 4.0**, not public domain. Use
requires attribution:

> Postal code data © [GeoNames](https://www.geonames.org/), CC BY 4.0

The site footer carries this attribution and the server records it in
`internal/zipdata/dataset.json`.

## Regenerating

The data is generated offline by a manual maintenance script; it is **not**
part of `setup`, `release`, `deploy`, or CI:

```bash
./db/scripts/build-zip-dataset.sh --force
```

This downloads `US.zip` from `download.geonames.org`, extracts `US.txt`, keeps
5-digit numeric ZIPs, sorts by ZIP, gzips the result, and rewrites
`dataset.json`. Review `git diff` and commit both files. The server's
`data_seeds` bookkeeping (keyed by `sha256`) makes a redeploy self-heal: a
changed dataset is re-applied automatically on next start/`-seed-only`.

## Format

Tab-separated, no header, one row per ZIP:

```
zip  lat  lng  city  state
```

Example:

```
00501	40.8154	-73.0451	Holtsville	NY
```
