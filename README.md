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
├── db/migrations/       # Database migrations
├── internal/            # Internal application code
├── templates/           # Templ templates
└── static/              # Static assets (CSS, JS, images)
```
