package templates

import (
	"strconv"
	"strings"

	"github.com/emerald/traditionbuilders/internal/models"
)

// formatUSD renders a whole-dollar amount as "$1,250,000".
func formatUSD(dollars int) string {
	s := strconv.Itoa(dollars)
	if len(s) <= 3 {
		return "$" + s
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return "$" + string(out)
}

// completedYear pulls the year out of a YYYY-MM-DD date string.
// It returns "" when the date is missing or malformed.
func completedYear(date string) string {
	if len(date) < 4 {
		return ""
	}
	return date[:4]
}

// displayLocation prefers the canonical city/state from the zip_codes join and
// falls back to the free-text location on the professional record.
func displayLocation(p models.Professional) string {
	if p.City != "" && p.State != "" {
		if p.ZipCode != "" {
			return p.City + ", " + p.State + " " + p.ZipCode
		}
		return p.City + ", " + p.State
	}
	return p.Location
}

// telHref builds a tel: URI from a display phone number like "(407) 555-0101",
// keeping only digits and a leading "+".
func telHref(phone string) string {
	var b strings.Builder
	b.WriteString("tel:")
	for i, c := range phone {
		switch {
		case c >= '0' && c <= '9':
			b.WriteRune(c)
		case c == '+' && i == 0:
			b.WriteRune(c)
		}
	}
	return b.String()
}
