package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/emerald/traditionbuilders/internal/models"
)

// ErrProfessionalNotFound is returned when no professional exists with the
// requested id.
var ErrProfessionalNotFound = errors.New("professional not found")

// GetProfessional returns the full directory record for a single professional,
// joined to zip_codes for a canonical city/state.
//
// ErrProfessionalNotFound is returned when id matches no row.
func (s *Store) GetProfessional(ctx context.Context, id int) (models.Professional, error) {
	var (
		p                                                    models.Professional
		phone, location, bio, zipCode, city, state, imagePath sql.NullString
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT p.id, p.name, p.email, p.phone, p.specialty, p.location, p.bio,
		       p.zip_code, z.city, z.state, p.image_path
		FROM professionals p
		LEFT JOIN zip_codes z ON p.zip_code = z.zip
		WHERE p.id = ?
	`, id).Scan(
		&p.ID, &p.Name, &p.Email, &phone, &p.Specialty, &location, &bio,
		&zipCode, &city, &state, &imagePath,
	)
	if err == sql.ErrNoRows {
		return models.Professional{}, ErrProfessionalNotFound
	}
	if err != nil {
		return models.Professional{}, err
	}

	p.Phone = phone.String
	p.Location = location.String
	p.Bio = bio.String
	p.ZipCode = zipCode.String
	p.City = city.String
	p.State = state.String
	p.ImagePath = imagePath.String

	return p, nil
}

// ListProjects returns the projects credited to a professional, most recently
// completed first. The returned slice is always non-nil.
func (s *Store) ListProjects(ctx context.Context, professionalID int) ([]models.Project, error) {
	projects := make([]models.Project, 0)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, description, location, cost_estimate, completed_at
		FROM projects
		WHERE professional_id = ?
		ORDER BY completed_at DESC, id DESC
	`, professionalID)
	if err != nil {
		return projects, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			pr                          models.Project
			description, location, done sql.NullString
			cost                        sql.NullInt64
		)
		if err := rows.Scan(&pr.ID, &pr.Title, &description, &location, &cost, &done); err != nil {
			continue
		}
		pr.Description = description.String
		pr.Location = location.String
		pr.CostEstimate = int(cost.Int64)
		pr.CompletedAt = done.String
		projects = append(projects, pr)
	}
	if err := rows.Err(); err != nil {
		return projects, err
	}

	return projects, nil
}
