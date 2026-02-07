.PHONY: help dev build css templ migrate migrate-down migrate-status clean setup test

help:
	@echo "Traditional Builders - Available Commands"
	@echo "=========================================="
	@echo ""
	@echo "  make setup         - Initialize project (migrations, deps, css, templ)"
	@echo "  make dev           - Run development server"
	@echo "  make build         - Build production binary"
	@echo ""
	@echo "  make css           - Build Tailwind CSS"
	@echo "  make css-watch     - Watch and rebuild CSS"
	@echo "  make templ         - Generate templ templates"
	@echo ""
	@echo "  make migrate       - Run database migrations"
	@echo "  make migrate-down  - Rollback last migration"
	@echo "  make migrate-status- Show migration status"
	@echo ""
	@echo "  make test          - Run tests"
	@echo "  make clean         - Remove generated files"
	@echo ""

setup: clean
	@echo "Setting up Traditional Builders..."
	@mkdir -p bin static/css
	@echo "Running migrations..."
	@goose -dir db/migrations sqlite3 traditionbuilders.db up
	@echo "Generating templates..."
	@templ generate
	@echo "Building CSS..."
	@tailwindcss -i static/css/input.css -o static/css/output.css
	@echo ""
	@echo "✓ Setup complete! Run 'make dev' to start the server."

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

test:
	go test ./... -v

clean:
	rm -f traditionbuilders.db
	rm -f static/css/output.css
	rm -f bin/server
	rm -rf bin
