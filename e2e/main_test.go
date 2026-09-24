//go:build e2e

// Package e2e drives the real application in a real Chromium. It is the
// browser layer of the front-end confidence system described in
// PLAN_FRONTEND_TESTING.md.
//
// Every file in this package is behind the `e2e` build tag so the fast
// `go test ./...` gate never launches a browser. Run it with:
//
//	make e2e        # or: go test -tags=e2e ./e2e/...
//
// The runner is go-rod (pure Go over the Chrome DevTools Protocol). The plan
// recommended Playwright-go, but its driver CDN returned 404 on this machine
// and its downloaded browsers do not execute on NixOS; go-rod drives the
// system Chromium directly with no driver step. See the report.
package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"

	"github.com/emerald/traditionbuilders/internal/handlers"
	"github.com/emerald/traditionbuilders/internal/store"
	"github.com/emerald/traditionbuilders/internal/zipdata"
	_ "github.com/mattn/go-sqlite3"
)

// baseURL is the address of the server under test. It is set once in TestMain
// and read by every page-object Load method.
var baseURL string

// repoRoot is the absolute path to the repository root, resolved once.
var repoRoot string

// TestMain owns the full lifecycle: migrate a throwaway SQLite database, start
// the real HTTP server on a free port, run the suite, then tear everything
// down. This keeps tests independent of the developer's live database.
func TestMain(m *testing.M) {
	code := runMain(m)
	os.Exit(code)
}

func runMain(m *testing.M) int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	slog.SetDefault(logger)

	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: cannot find repo root:", err)
		return 1
	}
	repoRoot = root

	tmpDir, err := os.MkdirTemp("", "tb-e2e-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: temp dir:", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	db, err := migrateAndSeed(filepath.Join(tmpDir, "e2e.db"), root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: database setup:", err)
		return 1
	}
	defer db.Close()

	server, listener, err := startServer(db, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: start server:", err)
		return 1
	}
	defer server.Shutdown(context.Background())

	baseURL = "http://" + listener.Addr().String()
	fmt.Fprintf(os.Stderr, "e2e: server listening at %s\n", baseURL)

	if err := waitForHealth(baseURL, 10*time.Second); err != nil {
		fmt.Fprintln(os.Stderr, "e2e: server never became healthy:", err)
		return 1
	}

	return m.Run()
}

// findRepoRoot walks up from the test's working directory until it finds go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// migrateAndSeed creates a fresh SQLite database at path and applies every real
// migration from db/migrations via the goose library, so the browser suite
// asserts against the same schema and seed data production uses.
func migrateAndSeed(path, root string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err := goose.Up(db, filepath.Join(root, "db", "migrations")); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	// Apply the embedded full zip dataset, exactly as production does, so the
	// browser suite exercises the real ~41k-row search surface.
	if _, err := zipdata.EnsureLoaded(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("seed zip codes: %w", err)
	}
	return db, nil
}

// startServer binds a free port and serves the real application handler.
func startServer(db *sql.DB, root string) (*http.Server, net.Listener, error) {
	h := &handlers.Handler{Store: store.New(db)}
	handler := h.Routes(db, filepath.Join(root, "static"))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}

	server := &http.Server{Handler: handler}
	go func() {
		_ = server.Serve(listener)
	}()
	return server, listener, nil
}

// waitForHealth polls /healthz until it returns 200 or the deadline passes.
func waitForHealth(base string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/healthz")
		if err != nil {
			lastErr = err
			time.Sleep(50 * time.Millisecond)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		lastErr = fmt.Errorf("status %d", resp.StatusCode)
		time.Sleep(50 * time.Millisecond)
	}
	return lastErr
}

// routeURL joins baseURL with a path, tolerating both "/x" and "x".
func routeURL(path string) string {
	return baseURL + "/" + strings.TrimPrefix(path, "/")
}
