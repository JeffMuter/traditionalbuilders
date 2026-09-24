# Post-Launch: Request Permission to Use Builder Logos

## Context

We want to show each builder's logo on its directory/profile listing. We scraped
logos from firm websites during build-out and did a permissions sweep. **None of
them are cleared for use**, so **all logos were removed from the site** — the
served tree (`static/images/`) currently contains no logos.

Six logos from *silent* firms are held in `assets/logo-permissions/` (not served)
so we can attach them when we reach out. Everything else was deleted.

Goal: after launch, contact each firm below, ask for permission to display their
logo next to their listing, and only publish on an explicit yes.

## Firms to contact (silent — held for the ask)

| # | Firm | Contact route | Logo file held | Asked | Reply | Permission |
|---|------|---------------|----------------|:-----:|:-----:|:----------:|
| 1 | ART Architects | https://www.artarchitects.com/contact/ | `art-architects.svg` | ☐ | ☐ | ☐ |
| 2 | Goosewing Timberworks | https://www.goosewingtimberworks.com/contact-3/ | `goosewing-timberworks.png` | ☐ | ☐ | ☐ |
| 3 | Feathermark Timber Frames | site contact form (`#CONTACT`) | `feathermark-timberframes.jpg` | ☐ | ☐ | ☐ |
| 4 | Settlement Post & Beam | info@settlementpostandbeam.com | `settlement-post-and-beam.png` | ☐ | ☐ | ☐ |
| 5 | Michael Gimber (Imber) | https://www.michaelgimber.com/contact/ | `michael-gimber.png` | ☐ | ☐ | ☐ |
| 6 | Neoclassical Builders | https://www.neoclassicalbuilders.com/contact | `neoclassical-builders.png` | ☐ | ☐ | ☐ |

> ART Architects and Goosewing carry only a bare `© <year>` notice with no
> "all rights reserved" / trademark language, so they were treated as silent.
> If that judgement is wrong, treat them as restricted and ask anyway.

## Boundary cases (ask only if we actually want the logo)

These carried an explicit copyright/trademark reservation, so we deleted the
files. If we want their logo, request in writing first:

- Vermont Frames — "logo mark is a federally registered trademark … all rights reserved"
- Mellows & Paladino — `robots.txt` disallows AI crawlers (ClaudeBot, GPTBot, CCBot, anthropic-ai)
- Jeff Johnson Timber Frames — "Copyright & Trademark Notice … All Rights Reserved"
- Custom Timber Frames — "© … All Rights Reserved"
- Hardwick Post & Beam — "All contents © … All rights reserved"
- Pinneo Construction — "© 2026 All Rights Reserved"
- Carpenter & MacNeille, McCrery Architects — live sites 403 bots (marks pulled via Wayback, likely stale)
- Patrick Ahearn, Fairfax & Sammons — no grant found
- Duncan G. Stroik, BuildMentor, Buxton Beam — no image logo exists (text-only marks)

## Task checklist

- [ ] Draft a short permission-request email (friendly, links back to their site, offers a listing benefit)
- [ ] Send to the six firms above
- [ ] Track replies in the table
- [ ] On **yes**: record date + exact grant wording, move the file from `assets/logo-permissions/` into `static/images/logos/`, and add it to a served-tree attribution file
- [ ] On **no** or no reply: delete the held file
- [ ] When any logo is live, verify it renders in light + dark themes (white-on-light problem)

## Guardrails

- Nothing in `assets/logo-permissions/` may be moved into `static/` until a written yes exists.
- Directory listings should show **text name + link** regardless — that is always safe.
- Do not re-download the restricted logos.
