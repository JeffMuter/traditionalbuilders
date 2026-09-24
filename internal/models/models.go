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

// Professional is the full directory record for a single builder or architect,
// as shown on their profile page.
type Professional struct {
	ID        int
	Name      string
	Email     string
	Phone     string
	Specialty string
	Location  string
	Bio       string
	ZipCode   string
	City      string
	State     string
	ImagePath string
}

// Project is a single past project credited to a Professional.
type Project struct {
	ID           int
	Title        string
	Description  string
	Location     string
	CostEstimate int    // dollars; 0 when not disclosed
	CompletedAt  string // YYYY-MM-DD; empty when not recorded
}
