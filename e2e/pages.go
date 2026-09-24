//go:build e2e

package e2e

import (
	"testing"
	"time"
)

// shortWait is the default timeout for an element to appear. It is generous
// enough for a local server render but short enough that a genuinely missing
// element fails the suite quickly.
const shortWait = 5 * time.Second

// landingPage is the page object for GET / .
type landingPage struct {
	session *pageSession
}

func (p landingPage) Load() landingPage {
	p.session.page.MustNavigate(routeURL("/")).MustWaitLoad()
	return p
}

// SearchZip types zip into the search box and submits. It arms a navigation
// wait *before* clicking so the assertion cannot race the form submit (a fast
// local server can complete the navigation before MustWaitLoad is called).
func (p landingPage) SearchZip(zip string) buildersPage {
	waitNav := p.session.page.MustWaitNavigation()
	input := p.session.page.MustElement(`[data-testid="zip-search"]`)
	input.MustInput(zip)
	p.session.page.MustElement(`[data-testid="zip-submit"]`).MustClick()
	waitNav()
	return buildersPage{session: p.session}
}

// buildersPage is the page object for GET /builders?zip= .
type buildersPage struct {
	session *pageSession
}

func (p buildersPage) Load(zip string) buildersPage {
	p.session.page.MustNavigate(routeURL("/builders?zip=" + zip)).MustWaitLoad()
	return p
}

// CardNames returns the builder names rendered on the results page.
func (p buildersPage) CardNames(t *testing.T) []string {
	t.Helper()
	cards := p.session.page.MustElements(`[data-testid="builder-card"]`)
	var names []string
	for _, c := range cards {
		if name, err := c.Attribute("data-builder-name"); err == nil && name != nil {
			names = append(names, *name)
		}
	}
	return names
}

// ClickFirstProfile follows the first card's "View Profile" link and returns
// the name of the card that was clicked, so the caller can assert the profile
// matches the specific card rather than "some profile".
func (p buildersPage) ClickFirstProfile(t *testing.T) (clickedName string, profile profilePage) {
	t.Helper()
	card := p.session.page.MustElement(`[data-testid="builder-card"]`)
	name, err := card.Attribute("data-builder-name")
	if err != nil || name == nil {
		t.Fatalf("first builder card has no data-builder-name: %v", err)
	}
	card.MustElement(`[data-testid="view-profile"]`).MustClick()
	p.session.page.Timeout(shortWait).MustWaitLoad()
	return *name, profilePage{session: p.session}
}

// profilePage is the page object for GET /builders/{id} .
type profilePage struct {
	session *pageSession
}

func (p profilePage) Load(id string) profilePage {
	p.session.page.MustNavigate(routeURL("/builders/" + id)).MustWaitLoad()
	return p
}
