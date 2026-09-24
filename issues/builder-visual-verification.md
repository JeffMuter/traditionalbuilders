# Builder Visual Verification & Icons

## Problem Statement

When browsing the platform, users cannot visually verify if a builder's style matches their vision before reaching out. This leads to:

- Low conversion rates (investigated builders who eventually don't contact)
- Unmatched expectations between what users want and what builders deliver
- Frustration with generic listings that say nothing about capabilities

## Requirements

### 1. Builder Icons (Required)

Each builder profile must display a distinct icon/image representing the builder:

- **No icons → Error state** (not empty state)
- Icons should convey builder type/type of work they specialize in
- Examples: bricklayer, architect, carpenter, masonry, restoration specialist

### 2. Ability Preview Widget

Every builder profile needs a visual "what they can do" preview for the user to evaluate style match:

- Must show something that looks like a **completed residential project**
- Positioned prominently (likely above or alongside bio)
- Sufficient quality that users can assess design/approach
- Examples: photo of finished home, room, exterior work, or related craftsmanship

### 3. Empty State Handling

- Builder with no icon → **Error/Skipped state** on directory/gallery pages
- Builder with no project preview → **Error/Skipped state** (or clear "No preview available" indicator)
- Clear messaging about why the builder isn't showing details (e.g., "This builder hasn't provided portfolio samples yet")

## User Experience Goals

Users visiting a builder profile should be able to:

1. Quickly see what this builder does (icon + specialty)
2. Visually confirm if the builder's style matches their vision (project preview)
3. Decide whether to reach out *or* skip to a different builder
4. Understand if a builder is incomplete/matching missing data

## Failure Modes to Avoid

- Blank spaces where icons or previews should be
- "Empty" states that look like bugs
- Users wasting time reaching out only to find misaligned expectations
- Missing builder profiles due to incomplete data, but no clear indication to users

## Questions for Further Clarification

- What size/format for project previews in loader? (carousel vs block vs gallery)
- Should we allow multiple preview images or just one primary?
- Does the icon need to be editable by the builder, or can admin/matcher assign?
- Where exactly do we show these on the directory vs profile pages?
