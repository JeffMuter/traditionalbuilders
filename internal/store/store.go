// Package store owns all database queries and returns domain types from
// internal/models. Handlers never write raw SQL; they call store methods.
package store

import (
	"context"
	"database/sql"
	"math"
	"sort"

	"github.com/emerald/traditionbuilders/internal/models"
)

// Store wraps a *sql.DB and exposes typed query methods.
type Store struct {
	db *sql.DB
}

// New returns a Store backed by db.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// maxResults caps the number of builders returned by a proximity search.
const maxResults = 10

// searchRadiusMiles bounds the prefilter used before the exact haversine sort.
// It is generous enough that the nearest maxResults are always included while
// keeping the candidate scan small; widen it if searches legitimately span a
// larger region.
const searchRadiusMiles = 500.0

// FindBuildersNear looks up the lat/lng for zip, then returns the closest
// professionals sorted by haversine distance, capped at maxResults.
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

	// 2. Bounding-box prefilter lets the lat/lng index narrow the candidate
	//    set before the exact haversine pass. Degrees per mile shrink with
	//    latitude, so the longitude span is divided by cos(lat).
	latDelta := searchRadiusMiles / 69.0
	lngDelta := searchRadiusMiles / (69.0 * math.Cos(userLat*math.Pi/180))
	if lngDelta > 180 {
		lngDelta = 180
	}

	// 3. Fetch candidate professionals with coordinates, joining to city/state.
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.specialty, COALESCE(p.zip_code, ''),
		       p.latitude, p.longitude,
		       COALESCE(z.city, ''), COALESCE(z.state, '')
		FROM professionals p
		LEFT JOIN zip_codes z ON p.zip_code = z.zip
		WHERE p.latitude IS NOT NULL AND p.longitude IS NOT NULL
		  AND p.latitude  BETWEEN ? AND ?
		  AND p.longitude BETWEEN ? AND ?
	`, userLat-latDelta, userLat+latDelta,
		userLng-lngDelta, userLng+lngDelta)
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

	// 4. Sort ascending by distance, cap at maxResults.
	sort.Slice(results, func(i, j int) bool {
		return results[i].DistanceMiles < results[j].DistanceMiles
	})
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}
