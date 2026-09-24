.PHONY: help dev dev-watch build css templ migrate migrate-down migrate-status \
        reset-db seed-zip-codes clean setup test ci release

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
	@echo "  make ci             - Full CI gate: generate + fmt + vet + test -race + build"
	@echo "  make release        - Build linux/amd64 + linux/arm64 release tarballs into dist/"
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

# Drop the database file, replay all migrations from scratch, then load the
# full embedded zip dataset so a clean dev DB matches prod (~41k rows).
reset-db:
	rm -f traditionbuilders.db
	goose -dir db/migrations sqlite3 traditionbuilders.db up
	go run ./cmd/seed-zips
	@echo "Database reset complete."

# Load the full GeoNames US zip code dataset (~41k rows) into the DB.
# The dataset is vendored in internal/zipdata and embedded in the binary, so
# this is offline and idempotent (no download).
seed-zip-codes:
	go run ./cmd/seed-zips

test:
	go test ./... -v

# ci is the single source of truth for "is this build healthy".
# GitHub Actions installs the toolchain (templ, tailwindcss) and calls this,
# so local and CI behaviour cannot drift. See PLAN_CI_CD.md.
ci: templ css
	@echo "==> gofmt"
	@test -z "$$(gofmt -l . | grep -v '_templ.go')" || { echo "gofmt needed:"; gofmt -l . | grep -v '_templ.go'; exit 1; }
	@echo "==> go vet"
	go vet ./...
	@echo "==> go test -race"
	go test -race ./...
	@echo "==> go build"
	go build -o bin/server ./cmd/server
	@echo "✓ CI gate passed"

# Build portable release tarballs (binary + static + migrations + unit file).
# Used by .github/workflows/release.yml on v* tags; runnable locally too.
release: css templ
	@command -v templ >/dev/null || { echo "templ not in PATH"; exit 1; }
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/server-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build -o dist/server-arm64 ./cmd/server
	@for arch in amd64 arm64; do \
		stage="dist/traditionalbuilders-linux-$$arch"; \
		rm -rf "$$stage"; mkdir -p "$$stage"; \
		cp "dist/server-$$arch" "$$stage/server"; \
		cp -r static "$$stage/static"; \
		cp -r db/migrations "$$stage/migrations"; \
		cp deploy/traditionalbuilders.service "$$stage/"; \
		tar -C dist -czf "$$stage.tar.gz" "traditionalbuilders-linux-$$arch"; \
	done
	@rm -f dist/server-amd64 dist/server-arm64
	@echo "✓ release tarballs in dist/"

setup: clean
	@echo "Setting up Traditional Builders..."
	@mkdir -p bin static/css tmp
	@echo "Running migrations..."
	@goose -dir db/migrations sqlite3 traditionbuilders.db up
	@echo "Loading zip code dataset..."
	@go run ./cmd/seed-zips
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
