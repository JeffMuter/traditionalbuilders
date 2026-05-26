// Package models defines shared domain types used across the handlers,
// store, and template layers.
package models

// BuilderResult is a single traditional builder returned by a proximity search.
type BuilderResult struct {
	ID            int
	Name          string
	Specialty     string
	ZipCode       string
	City          string
	State         string
	DistanceMiles float64
}
