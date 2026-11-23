# Tasks

Running log of planned work, cost estimates, and modifications for **vs-modlist-generator**.

## Legend

| Status  | Meaning                                      |
|---------|----------------------------------------------|
| `📋 Planned`     | Not yet started                            |
| `⏳ In Progress` | Currently being worked on                  |
| `✅ Done`       | Completed and integrated                   |
| `❌ Blocked`    | Waiting on something external              |
| `~ Cancelled ~`  | Scrapped or superseded                     |

## Active Work

### 📋 Add AGENTS.md convention for changelog discipline
- **Cost**: 5 min (documentation only)
- **Description**: Enforce that every change must be recorded in CHANGELOG.md before committing.
- **Status**: ✅ Done — added to development conventions section of AGENTS.md and applied retroactively.

### 📋 Create tasks.md for project planning and tracking
- **Cost**: 10 min (documentation)
- **Description**: Establish a central file for planning features, estimating cost, and tracking all modifications.
- **Status**: ✅ Done — this file is it.

---

## Backlog

*(Nothing planned yet — add new tasks below as they arise.)*

| Status | Task | Cost Estimate | Notes |
|--------|------|---------------|-------|
| 📋 Planned | *(add here)* | TBD | |

---

## Completed History

| Date | Modification | Cost | Reference |
|------|--------------|------|-----------|
| 2026-11-23 | Created `AGENTS.md` with full project memory | ~5 min | v1.4.1 |
| 2026-11-23 | Added JS source-of-truth note to AGENTS.md | ~2 min | v1.4.1 |
| 2026-11-23 | Added changelog discipline + task tracking conventions | ~5 min | — |
| 2026-11-23 | Created `tasks.md` for planning and modification tracking | ~5 min | — |
| 2026-11-23 | Moved `helpMsg` to highest valid Go position (after imports) | ~30 min (incl. syntax debugging) | — |
| 2026-11-23 | Updated AGENTS.md: added planning & estimation discipline rule | ~5 min | — |

## Lessons Learned

- **Don't skip tasks.md** — every request should be planned, costed, and presented before implementation.
- When moving code in Go, remember imports must appear before other declarations; the highest valid position for a `const` is immediately after the import block closes.
