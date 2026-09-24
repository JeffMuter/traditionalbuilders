package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/emerald/traditionbuilders/internal/logging"
	"github.com/emerald/traditionbuilders/templates"
)

// requestIDHeader is both read (to allow upstream correlation) and written on
// every response so a user-facing error can be matched to server logs.
const requestIDHeader = "X-Request-ID"

// newRequestID returns a random 128-bit hex ID. crypto/rand failure is
// extremely unlikely; a timestamp fallback keeps the request path panic-free.
func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "ts-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b[:])
}

// statusWriter records the status, bytes written, and whether a header was
// written, so middleware can log the outcome and the panic recoverer can avoid
// writing a second header after the handler already responded.
type statusWriter struct {
	http.ResponseWriter
	status      int
	bytes       int64
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

// clientIP extracts the best-effort client address, honouring a single
// X-Forwarded-For hop set by a trusted reverse proxy before falling back to
// RemoteAddr. It is for logging only, never for authorization.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
			return first
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// isHealthOrAsset reports whether the request should be excluded from
// per-request INFO logging (static files and health probes are high-volume
// and low-signal).
func isHealthOrAsset(path string) bool {
	return strings.HasPrefix(path, "/static/") ||
		path == "/healthz" ||
		path == "/favicon.ico"
}

// RequestLogger assigns a request ID, installs a request-scoped logger, and
// emits one log line per request with method, path, status, bytes, client IP,
// and duration. Static assets and health probes are logged at DEBUG to keep
// production output signal-rich.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}

		logger := slog.Default().With("request_id", id)
		ctx := logging.WithLogger(logging.WithRequestID(r.Context(), id), logger)
		r = r.WithContext(ctx)

		w.Header().Set(requestIDHeader, id)

		sw := &statusWriter{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(sw, r)

		status := sw.status
		if !sw.wroteHeader {
			status = http.StatusOK // net/http writes 200 for an empty handler
		}

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", sw.bytes,
			"ip", clientIP(r),
			"user_agent", r.UserAgent(),
			"duration", time.Since(start).String(),
		}
		if r.URL.RawQuery != "" {
			attrs = append(attrs, "query", r.URL.RawQuery)
		}

		if isHealthOrAsset(r.URL.Path) {
			logger.Debug("request", attrs...)
			return
		}
		logger.Info("request", attrs...)
	})
}

// Recover converts a panic in a downstream handler into a logged 500 response
// so one bad request cannot take down the server or leak a stack trace.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			logging.FromContext(r.Context()).Error("panic recovered",
				"method", r.Method,
				"path", r.URL.Path,
				"panic", rec,
				"stack", string(debug.Stack()),
			)
			// If a response has already started, the only safe action is to
			// let the connection close; otherwise replace it with a clean 500.
			if !sw.wroteHeader {
				renderStatus(sw, r, http.StatusInternalServerError, templates.ErrorPage(
					http.StatusInternalServerError,
					"Something went wrong",
					"Something went wrong on our end. Please try again in a moment.",
				))
			}
		}()
		next.ServeHTTP(sw, r)
	})
}
