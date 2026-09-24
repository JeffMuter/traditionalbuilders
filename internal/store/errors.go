package store

import (
	"context"
	"errors"
	"net/http"
)

// ErrZipNotFound is returned when a zip code is not in the zip_codes table.
var ErrZipNotFound = errors.New("zip code not found")

// IsZipNotFound reports whether err indicates a zip code was not found.
func IsZipNotFound(err error) bool {
	return errors.Is(err, ErrZipNotFound)
}

// IsDbError reports whether err is a database-related (transport/context) error
// that should surface as a server error rather than a client error.
func IsDbError(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled)
}

// ToHTTPStatus maps a store error to the HTTP status the handler should return:
//
//	ErrZipNotFound -> 406 Not Acceptable (valid format, no such zip)
//	DB/context err -> 500 Internal Server Error
//	anything else  -> 500 Internal Server Error (unknown = server fault)
func ToHTTPStatus(err error) int {
	switch {
	case IsZipNotFound(err):
		return http.StatusNotAcceptable
	default:
		return http.StatusInternalServerError
	}
}
