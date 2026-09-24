package handlers

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/a-h/templ"

	"github.com/emerald/traditionbuilders/templates"
)

// render writes component to w, buffering the response so a mid-render failure
// cannot send a partial page. Render errors are logged with request context
// rather than silently dropped.
func render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	handlerFor(component, 0).ServeHTTP(w, r)
}

// renderStatus is render with an explicit HTTP status code.
func renderStatus(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	handlerFor(component, status).ServeHTTP(w, r)
}

func handlerFor(component templ.Component, status int) *templ.ComponentHandler {
	opts := []func(*templ.ComponentHandler){
		templ.WithErrorHandler(func(r *http.Request, err error) http.Handler {
			slog.Error("render template",
				"method", r.Method,
				"path", r.URL.Path,
				"err", err,
			)
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			})
		}),
	}
	if status != 0 {
		opts = append(opts, templ.WithStatus(status))
	}
	return templ.Handler(component, opts...)
}

// serverError logs the underlying error with request context and renders the
// styled 500 page. The error itself is never shown to the client.
func serverError(w http.ResponseWriter, r *http.Request, op string, err error) {
	slog.Error("request failed",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery,
		"op", op,
		"err", err,
	)
	renderStatus(w, r, http.StatusInternalServerError, templates.ErrorPage(
		http.StatusInternalServerError,
		"Something went wrong",
		"Something went wrong on our end. Please try again in a moment.",
	))
}

// statusWriter records whether a response header was written and with what
// status, so middleware can log the outcome and the panic recoverer can avoid
// writing a second header after the handler already responded.
type statusWriter struct {
	http.ResponseWriter
	status      int
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
	return w.ResponseWriter.Write(b)
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
			slog.Error("panic recovered",
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

// LogRequests records the method, path, status, and duration of every request.
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start),
		)
	})
}
