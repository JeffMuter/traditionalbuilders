# Traditional Builders

A platform to help people find traditional architects and builders, and estimate construction costs for traditional architecture projects.

Traditional architecture emphasizes timeless design principles, quality craftsmanship, and materials that stand the test of time. This platform connects clients with skilled professionals who specialize in classical, vernacular, and traditional building methods.

## Features

- Directory of traditional architects and builders
- Project cost estimation tools
- Portfolio browsing and reviews
- Direct contact with verified professionals

## Tech Stack

- **Backend**: Go with standard library HTTP server
- **Templates**: templ for type-safe templating
- **Database**: SQLite with Goose migrations
- **Styling**: Tailwind CSS
- **Dev Environment**: Nix shell

## Getting Started

```bash
# Enter development environment
nix-shell

# Run database migrations
goose -dir db/migrations sqlite3 traditionbuilders.db up

# Load the full US zip code dataset (embedded, offline, ~41k rows).
# `make dev` / `run` / `reset-db` also do this automatically.
go run ./cmd/seed-zips

# Generate templ templates
templ generate

# Run the server
go run cmd/server/main.go
```

Visit `http://localhost:8080`

## Project Structure

```
.
├── cmd/server/          # Main application entry point
├── cmd/seed-zips/       # Loads the embedded zip dataset into the DB
├── db/migrations/       # Database migrations
├── internal/            # Internal application code
│   └── zipdata/         # Embedded GeoNames zip dataset + loader
├── templates/           # Templ templates
└── static/              # Static assets (CSS, JS, images)
```

## Zip code data & attribution

The proximity search uses the GeoNames US postal-code dataset, vendored as a
filtered, gzipped TSV in `internal/zipdata/` and embedded into the server
binary (no network at deploy). It is licensed **CC BY 4.0** and the site footer
carries the required attribution.

Refresh cadence: manual. `./db/scripts/build-zip-dataset.sh --force` regenerates
the vendored file from `download.geonames.org`; review and commit the diff. The
server's `data_seeds` bookkeeping (keyed by checksum) re-applies a changed
dataset automatically on the next start or `server -seed-only`. See
`internal/zipdata/README.md` and `PLAN_ZIP_DATASET.md`.
