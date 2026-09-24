# Rough Project Estimation Page

## Context

The site advertises cost transparency — the landing page sells "Get accurate
estimates based on real traditional construction data", and every page's header
carries an **"Estimate Costs"** link. That link is currently `href="#"` (a dead
placeholder), so the promise has no destination.

`internal/handlers/crawl_test.go` / the E2E link checks treat `/` targets as
valid but `#` as a known placeholder. Until this page exists, that placeholder
stays allowlisted in the crawler. Removing the allowlist is the completion
signal for this issue.

## Requirements

A page that gives a user a **rough** cost range for a traditional building
project, before they contact a builder.

- Reachable from the existing "Estimate Costs" header link (replace `href="#"`).
- Route TBD, e.g. `/estimate`.
- Inputs (rough pass — keep it small):
  - Project type (new build, addition, restoration, etc.)
  - Approximate square footage
  - Zip code / region (reuse the existing zip → coordinates lookup)
  - Optional: finish level / scope
- Output: a cost **range**, clearly labeled as a rough estimate, with a
  "talk to a builder" call to action into the existing directory search.
- Must render server-side first (works without JS); htmx can upgrade the
  submit to an inline fragment later.

## Explicitly not required for the first pass

- Persistence of estimates.
- Per-builder pricing or bids.
- User accounts.

## Acceptance criteria

- Header "Estimate Costs" link navigates to the real page on every template
  (`landing.templ`, `gallery.templ`, `builders.templ`, `shell.templ`).
- The `href="#"` allowlist entry for this link is removed from the link crawler.
- `data-testid` attributes added to the form/controls per the front-end testing
  conventions in `PLAN_FRONTEND_TESTING.md`.
- System 1 render-contract tests cover the new route and its states.
- System 2 E2E covers: land → click "Estimate Costs" → fill → submit → see a
  range and reach the directory search from the CTA.
- Passes the a11y gate (no serious/critical axe violations).

## Open questions

- What cost dataset/model backs the range? (flat $/sqft by project type?)
- Should the estimate be a single range or itemized (design, materials, labor)?
- Does public "real traditional construction data" exist to seed it, or is it
  a curated assumption set at first?
