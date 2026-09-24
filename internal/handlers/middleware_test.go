package handlers

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emerald/traditionbuilders/internal/logging"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestRequestLogger_SetsAndEchoesRequestID(t *testing.T) {
	buf := captureLogs(t)
	var seenID string
	h := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenID = logging.RequestID(r.Context())
		w.WriteHeader(http.StatusTeapot)
	}))

	req, _ := http.NewRequest("GET", "/builders", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seenID == "" {
		t.Fatal("request ID not present in context")
	}
	if got := rec.Header().Get(requestIDHeader); got != seenID {
		t.Errorf("response header %q != context id %q", got, seenID)
	}
	log := buf.String()
	if !strings.Contains(log, "request_id="+seenID) {
		t.Errorf("log missing request_id: %s", log)
	}
	if !strings.Contains(log, "status=418") {
		t.Errorf("log missing status: %s", log)
	}
}

func TestRequestLogger_HonoursInboundRequestID(t *testing.T) {
	captureLogs(t)
	h := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := logging.RequestID(r.Context()); got != "upstream-abc" {
			t.Errorf("request ID = %q, want upstream-abc", got)
		}
	}))
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(requestIDHeader, "upstream-abc")
	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestRequestLogger_DefaultsStatusTo200(t *testing.T) {
	buf := captureLogs(t)
	h := RequestLogger(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(buf.String(), "status=200") {
		t.Errorf("expected status=200, got: %s", buf.String())
	}
}

func TestRequestLogger_StaticLoggedAtDebugOnly(t *testing.T) {
	buf := captureLogs(t)
	h := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/static/css/output.css", nil))
	if strings.Contains(buf.String(), "level=INFO") {
		t.Errorf("static asset should not log at INFO: %s", buf.String())
	}
}

func TestRecover_LogsWithRequestID(t *testing.T) {
	buf := captureLogs(t)
	// RequestLogger installs the scoped logger; Recover must pick it up.
	h := RequestLogger(Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("kaboom")
	})))
	req, _ := http.NewRequest("GET", "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "panic recovered") {
		t.Errorf("missing panic log: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "request_id=") {
		t.Errorf("panic log missing request_id: %s", buf.String())
	}
}

func TestLoggingContext_DefaultsWhenAbsent(t *testing.T) {
	if got := logging.FromContext(context.Background()); got == nil {
		t.Fatal("FromContext returned nil logger")
	}
	if got := logging.RequestID(context.Background()); got != "" {
		t.Errorf("RequestID = %q, want empty", got)
	}
}
