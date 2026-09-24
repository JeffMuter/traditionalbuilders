# QUALITY.md — Builder Inclusion Standard (DRAFT)

Status: **draft for review.** Nothing has been enforced against the directory yet.

The landing page (`templates/landing.templ:82`) already promises users "vetted
architects and builders with proven traditional building expertise." This file
is what "vetted" has to mean before that sentence is honest.

## Who the user is

Not a general homeowner. Someone who already knows what Georgian, Federal, and
Greek Revival are, who has looked at a lot of portfolios, and who bounces off
two things fast:

1. **Fake-traditional.** Builder-grade "traditional" — snap-in muntins, vinyl
   returns, garage-forward massing, wrong proportions. They can spot it in one
   photo and it destroys trust in the whole directory.
2. **Price they can never touch.** A flawless $$$$$ estate portfolio is not a
   lead, it's window shopping. The single most useful thing we can tell them is
   *what tier this firm actually operates in.*

Everything below serves those two.

## The check

Five gates. **G1–G3 are hard fails.** G4–G5 are scored and recorded, not fatal.

### G1 — Real, reachable, currently operating (hard fail)
- Site resolves; portfolio is not 404 or JS-only-empty.
- Evidence of work in the **last 3 years** (dated project, news, or copyright).
- A working contact path: phone, email, or form. No contact = not a lead.
- Named principal or firm entity. No anonymous listings.

### G2 — Genuinely traditional (hard fail)
Portfolio must show **traditional as the practice, not as an option**.
- At least **3 projects** in a named traditional idiom (Georgian, Federal,
  Greek/Colonial Revival, Shingle, vernacular, timber frame, classical).
- Correct fundamentals visible in photos: true divided lights, real material
  depth at eaves/cornices/casings, coherent proportion and massing.
- **Fail if** the portfolio is majority modern/contemporary with a token
  traditional project, or if "traditional" means a spec-builder facade.

### G3 — Portfolio is real, photographed work (hard fail)
- **Photographs of built projects**, not renderings, not stock, not mood boards.
- Minimum **6 projects** with **multiple photos each** — one hero shot per
  project is a marketing site, not a portfolio.
- Interiors as well as exteriors. Exterior-only hides the craft.

### G4 — Price tier, recorded (scored — the important one)
Every entry gets a tier. **Unknown is a valid, honest value — do not guess.**

| Tier | Typical project | What it means to the user |
|---|---|---|
| `$$` | Renovation / modest new build, ~<$500k | **The sweet spot.** Attainable. |
| `$$$` | Substantial custom home, ~$500k–1.5M | Reachable, stretch. |
| `$$$$` | Large custom, ~$1.5–4M | Aspirational. |
| `$$$$$` | Estate / institutional, $4M+ | Reference only. Label loudly. |

Signals to read, best first: published project costs · square footages · stated
minimum project size · scale and grounds in the photos · "estate/manor"
vocabulary · institutional-only client list.

Rule: **a `$$` or `$$$` firm outranks a better `$$$$$` firm** in results. We are
not ranking prestige, we are ranking who the user can actually hire.

### G5 — Consistency (scored)
From BUILDERS.md: *"some good homes, some attrocious."* A firm whose floor is
low is a risk even when its ceiling is high.

- `consistent` — portfolio holds a level throughout.
- `uneven` — good work present, but weak projects too. **Show this to the user.**
- `narrow` — strong but only one idiom or one project type.

### Also record (not gates)
Geography served · specialty idioms · design-build vs. architect-only (they hire
differently) · new build vs. restoration vs. both.

## Verdicts
- **Include** — passes G1–G3, tier recorded.
- **Include, flagged** — passes, but `uneven` or `$$$$$`. Surfaced with the caveat.
- **Lead** — looks right, not yet verified. Never shown to users.
- **Reject** — fails any of G1–G3. Record the reason so we don't re-review it.

## Open questions for review
1. Are these the right $ thresholds, or should tiers be regional?
2. Is 6 projects / 3 traditional the right bar, or too strict for good small shops?
3. Should `$$$$$` firms be listed at all, or dropped entirely?
4. Who verifies — manual review, or do we ask firms to self-report tier at `/join`?
