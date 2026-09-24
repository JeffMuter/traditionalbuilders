//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// chromiumExecutable returns the browser binary the suite should drive.
//
// Resolution order (mirrors PLAN_FRONTEND_TESTING.md § runner decision):
//  1. E2E_CHROMIUM_EXECUTABLE — explicit override, wins everywhere.
//  2. /run/current-system/sw/bin/chromium — the NixOS system Chromium, the only
//     browser known to execute on this machine.
//  3. "" — let go-rod's launcher download a pinned revision (CI runners).
//
// Playwright's own downloaded browser must NOT be used on NixOS: it does not
// execute there. go-rod is used precisely to avoid that constraint.
func chromiumExecutable() string {
	if v := os.Getenv("E2E_CHROMIUM_EXECUTABLE"); v != "" {
		return v
	}
	const nixChromium = "/run/current-system/sw/bin/chromium"
	if _, err := os.Stat(nixChromium); err == nil {
		return nixChromium
	}
	return ""
}

// viewport is a browser window size used for cross-viewport assertions.
type viewport struct {
	Name   string
	Width  int
	Height int
}

var (
	desktopViewport = viewport{Name: "desktop", Width: 1280, Height: 800}
	mobileViewport  = viewport{Name: "mobile", Width: 375, Height: 667}
)

// browserHarness owns the shared browser instance for the whole suite. One
// browser is launched lazily and reused; each test opens its own isolated
// page (via rod's incognito context) so localStorage and cookies do not leak.
type browserHarness struct {
	once    sync.Once
	browser *rod.Browser
	launch  string
	err     error
}

func (h *browserHarness) Browser(t *testing.T) *rod.Browser {
	t.Helper()
	h.once.Do(func() {
		path := chromiumExecutable()
		l := launcher.New().
			Headless(true).
			// A fixed, throwaway user-data dir keeps the suite from touching
			// the developer's real browser profile.
			Set("user-data-dir", filepath.Join(t.TempDir(), "chromium-profile")).
			Leakless(false)
		if path != "" {
			l = l.Bin(path)
		}
		u, err := l.Launch()
		if err != nil {
			h.err = fmt.Errorf("launch chromium (path=%q): %w", path, err)
			return
		}
		h.launch = u
		h.browser = rod.New().ControlURL(u)
		if err := h.browser.Connect(); err != nil {
			h.err = fmt.Errorf("connect to chromium: %w", err)
		}
	})
	if h.err != nil {
		t.Fatalf("browser harness: %v", h.err)
	}
	return h.browser
}

var harness browserHarness

// pageSession is one isolated browser tab plus the diagnostics captured while
// it navigates: console errors and failed same-origin requests.
type pageSession struct {
	t       *testing.T
	page    *rod.Page
	name    string
	mu      sync.Mutex
	console []consoleEntry
	failed  []failedRequest
}

type consoleEntry struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// failedRequest records a request that failed to load or returned ≥400. The
// resource type matters: an error route's *document* legitimately returns 404
// (asserted by the render-contract tests), while a broken stylesheet or image
// is always a defect.
type failedRequest struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// newPage opens a fresh incognito browser context and a single page inside it.
// The context (not just the page) is discarded at test end so storage is clean.
func newPage(t *testing.T, name string) *pageSession {
	t.Helper()

	browser := harness.Browser(t)
	incognito, err := browser.Incognito()
	if err != nil {
		t.Fatalf("open incognito context: %v", err)
	}
	t.Cleanup(func() { _ = incognito.Close() })

	page, err := incognito.Page(proto.TargetCreateTarget{})
	if err != nil {
		t.Fatalf("open page: %v", err)
	}

	s := &pageSession{
		t:    t,
		page: page,
		name: name,
	}
	s.watchDiagnostics()

	t.Cleanup(func() {
		s.captureArtifacts()
	})
	return s
}

// Page returns the underlying rod page for page objects that need direct access.
func (s *pageSession) Page() *rod.Page { return s.page }

// watchDiagnostics subscribes to console and network events for the lifetime of
// the page. The goroutine exits when the page or context closes, signalled via
// the returned stop function.
func (s *pageSession) watchDiagnostics() {
	// Runtime.enable and Network.enable make the browser emit these events.
	_ = proto.RuntimeEnable{}.Call(s.page)
	_ = proto.NetworkEnable{}.Call(s.page)

	go s.page.EachEvent(
		func(e *proto.RuntimeConsoleAPICalled) {
			if e.Type != proto.RuntimeConsoleAPICalledTypeError {
				return
			}
			text := ""
			for _, a := range e.Args {
				text += a.Description + a.Value.Str()
			}
			s.mu.Lock()
			s.console = append(s.console, consoleEntry{Type: string(e.Type), Text: text})
			s.mu.Unlock()
		},
		func(e *proto.NetworkLoadingFailed) {
			s.mu.Lock()
			s.failed = append(s.failed, failedRequest{
				Type:   string(e.Type),
				Detail: fmt.Sprintf("loading failed: %s (%s)", e.ErrorText, e.Type),
			})
			s.mu.Unlock()
		},
		func(e *proto.NetworkResponseReceived) {
			if e.Response != nil && e.Response.Status >= 400 {
				s.mu.Lock()
				s.failed = append(s.failed, failedRequest{
					Type:   string(e.Type),
					Detail: fmt.Sprintf("HTTP %d %s", e.Response.Status, e.Response.URL),
				})
				s.mu.Unlock()
			}
		},
	)()
}

// ConsoleErrors returns the console errors captured so far.
func (s *pageSession) ConsoleErrors() []consoleEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]consoleEntry(nil), s.console...)
}

// FailedRequests returns the failed/≥400 requests captured so far.
func (s *pageSession) FailedRequests() []failedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]failedRequest(nil), s.failed...)
}

// captureArtifacts writes a full-page screenshot and the page DOM to
// e2e/artifacts/<test-name>/ when the test has failed. On success it writes
// nothing (success screenshots are handled separately by the screenshot test).
func (s *pageSession) captureArtifacts() {
	if !s.t.Failed() {
		return
	}
	dir := filepath.Join(repoRoot, "e2e", "artifacts", sanitise(s.name))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		s.t.Logf("could not create artifact dir: %v", err)
		return
	}

	if png, err := s.page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	}); err != nil {
		s.t.Logf("could not capture screenshot: %v", err)
	} else {
		_ = os.WriteFile(filepath.Join(dir, "full-page.png"), png, 0o644)
	}

	if html, err := s.page.HTML(); err == nil {
		_ = os.WriteFile(filepath.Join(dir, "dom.html"), []byte(html), 0o644)
	}

	diag := map[string]any{
		"url":             s.page.MustInfo().URL,
		"console_errors":  s.ConsoleErrors(),
		"failed_requests": s.FailedRequests(),
	}
	if b, err := json.MarshalIndent(diag, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(dir, "diagnostics.json"), b, 0o644)
	}
	s.t.Logf("artifacts written to %s", dir)
}

// sanitise makes a test name safe to use as a directory name.
func sanitise(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

// withTimeout returns a copy of page whose operations time out after d. It is
// the single place the suite sets a deadline, so no test can hang forever.
func withTimeout(page *rod.Page, d time.Duration) *rod.Page {
	return page.Timeout(d)
}

// setViewport pins the page's device metrics so responsive layout can be
// asserted deterministically.
func setViewport(t *testing.T, s *pageSession, v viewport) {
	t.Helper()
	err := s.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             v.Width,
		Height:            v.Height,
		DeviceScaleFactor: 1,
		Mobile:            v.Width < 600,
	})
	if err != nil {
		t.Fatalf("set %s viewport: %v", v.Name, err)
	}
}
