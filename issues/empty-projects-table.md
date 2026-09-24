# `projects` Table Is Empty — Portfolio Section Renders Nothing

## Problem Statement

The `projects` table exists (migration `001_init.sql`), is read by the app, and
a "projects" section is rendered on every builder profile — but the table has
**zero rows** and there is no seeding, admin entry point, or import feeding it.

```
$ sqlite3 traditionbuilders.db "SELECT COUNT(*) FROM projects;"
0
```

### Where it is used

- `internal/store/professional.go` — `ListProjects()` queries `projects` by
  `professional_id`.
- `templates/profile.templ:71` — renders `<div ... data-testid="projects">`.
- E2E / render-contract tests assert the projects **section is visible**, but
  with no data it is effectively an empty feature.

### Impact

- On launch, every builder profile shows an empty/placeholder portfolio section.
- Users see a portfolio promise with no portfolio.
- Tests that assert the section exists pass on structure, not content — a
  false sense of completeness.

## Requirements

Choose one before go-live:

1. **Populate it** — provide a way to add projects (seed data, admin form, or
   import) so profiles have real portfolio entries; or
2. **Hide it** — suppress the projects section when a builder has no projects,
   with a clean empty state ("Portfolio coming soon" / nothing at all); or
3. **Defer** — explicitly mark the portfolio feature as post-launch and ensure
   the empty section does not appear.

## Acceptance criteria

- A builder profile either shows real project entries or a deliberate, polished
  empty state — never a blank section.
- Render-contract/E2E tests assert the chosen behavior, not just that the
  container exists.
- If populated: seed/entry mechanism is documented and reproducible.

## Open questions

- Is portfolio content expected at launch, or is it a fast-follow?
- If populated manually, what fields are required (`title`, `description`,
  `cost_estimate`, `completed_at`)?
