.PHONY: help dev dev-watch build css templ migrate migrate-down migrate-status \
        reset-db seed-zip-codes clean setup test

help:
	@echo "Traditional Builders - Available Commands"
	@echo "=========================================="
	@echo ""
	@echo "  make dev-watch      - Live-reload dev server (recommended)"
	@echo "  make dev            - One-shot dev server (no live reload)"
	@echo "  make build          - Build production binary"
	@echo ""
	@echo "  make css            - Build Tailwind CSS once"
	@echo "  make css-watch      - Watch and rebuild CSS"
	@echo "  make templ          - Generate templ templates once"
	@echo ""
	@echo "  make migrate        - Run pending migrations"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-status - Show migration status"
	@echo "  make reset-db       - Drop database and re-run all migrations"
	@echo "  make seed-zip-codes - Import full GeoNames US zip code dataset (~41k rows)"
	@echo ""
	@echo "  make setup          - Full init: migrate + templ + css"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Remove generated files"
	@echo ""

# Live-reload: runs templ --watch, tailwind --watch, and air in parallel via overmind.
# Requires air + overmind in PATH (provided by nix-shell).
dev-watch:
	overmind start

# One-shot: generate once then run. Useful outside nix-shell or for CI.
dev: css templ
	go run cmd/server/main.go

build: css templ
	go build -o bin/server cmd/server/main.go

css:
	tailwindcss -i static/css/input.css -o static/css/output.css

css-watch:
	tailwindcss -i static/css/input.css -o static/css/output.css --watch

templ:
	templ generate

migrate:
	goose -dir db/migrations sqlite3 traditionbuilders.db up

migrate-down:
	goose -dir db/migrations sqlite3 traditionbuilders.db down

migrate-status:
	goose -dir db/migrations sqlite3 traditionbuilders.db status

# Drop the database file and replay all migrations from scratch.
reset-db:
	rm -f traditionbuilders.db
	goose -dir db/migrations sqlite3 traditionbuilders.db up
	@echo "Database reset complete."

# Download the full GeoNames US zip code CSV and bulk-import it.
# Replaces the small seed set from migration 003 with ~41k real rows.
seed-zip-codes:
	go run cmd/seed-zips/main.go

test:
	go test ./... -v

setup: clean
	@echo "Setting up Traditional Builders..."
	@mkdir -p bin static/css tmp
	@echo "Running migrations..."
	@goose -dir db/migrations sqlite3 traditionbuilders.db up
	@echo "Generating templates..."
	@templ generate
	@echo "Building CSS..."
	@tailwindcss -i static/css/input.css -o static/css/output.css
	@echo ""
	@echo "✓ Setup complete! Run 'make dev-watch' to start the server."

clean:
	rm -f traditionbuilders.db
	rm -f static/css/output.css
	rm -f bin/server
	rm -rf bin tmp
