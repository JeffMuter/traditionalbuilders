package handlers

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/emerald/traditionbuilders/internal/logging"
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
			logging.FromContext(r.Context()).Error("render template",
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
	logging.FromContext(r.Context()).Error("request failed",
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
