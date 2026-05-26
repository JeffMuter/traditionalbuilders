// Package store owns all database queries and returns domain types from
// internal/models. Handlers never write raw SQL; they call store methods.
package store

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"sort"

	"github.com/emerald/traditionbuilders/internal/models"
)

// ErrZipNotFound is returned when the searched zip code is not in the
// zip_codes lookup table.
var ErrZipNotFound = errors.New("zip code not found")

// Store wraps a *sql.DB and exposes typed query methods.
type Store struct {
	db *sql.DB
}

// New returns a Store backed by db.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// FindBuildersNear looks up the lat/lng for zip, then returns the 10 closest
// professionals sorted by haversine distance.
//
// The returned slice is always non-nil. ErrZipNotFound is returned (with an
// empty slice) when zip is not in the zip_codes table.
func (s *Store) FindBuildersNear(ctx context.Context, zip string) ([]models.BuilderResult, error) {
	results := make([]models.BuilderResult, 0)

	// 1. Resolve user zip to coordinates
	var userLat, userLng float64
	err := s.db.QueryRowContext(ctx,
		`SELECT lat, lng FROM zip_codes WHERE zip = ?`, zip,
	).Scan(&userLat, &userLng)
	if err == sql.ErrNoRows {
		return results, ErrZipNotFound
	}
	if err != nil {
		return results, err
	}

	// 2. Fetch all professionals with coordinates, joining to city/state
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.specialty, p.zip_code,
		       p.latitude, p.longitude,
		       COALESCE(z.city, ''), COALESCE(z.state, '')
		FROM professionals p
		LEFT JOIN zip_codes z ON p.zip_code = z.zip
		WHERE p.latitude IS NOT NULL AND p.longitude IS NOT NULL
	`)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        int
			name      string
			specialty string
			zipCode   string
			lat, lng  float64
			city      string
			state     string
		)
		if err := rows.Scan(&id, &name, &specialty, &zipCode, &lat, &lng, &city, &state); err != nil {
			continue
		}
		dist := haversine(userLat, userLng, lat, lng)
		results = append(results, models.BuilderResult{
			ID:            id,
			Name:          name,
			Specialty:     specialty,
			ZipCode:       zipCode,
			City:          city,
			State:         state,
			DistanceMiles: math.Round(dist*10) / 10,
		})
	}
	if err := rows.Err(); err != nil {
		return results, err
	}

	// 3. Sort ascending by distance, cap at 10
	sort.Slice(results, func(i, j int) bool {
		return results[i].DistanceMiles < results[j].DistanceMiles
	})
	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}
