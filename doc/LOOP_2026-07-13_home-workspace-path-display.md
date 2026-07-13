---
title: "Home workspace path display (top bar + session cards)"
created: 2026-07-13
slug: home-workspace-path-display
path: doc/LOOP_2026-07-13_home-workspace-path-display.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: home-workspace-path-display

## Kind

**bug-repro** → after fix, **health-check** with inspect.

## Symptom (user)

On `http://127.0.0.1:8821/`:

1. **Top path** does not show the full workspace path.
2. **Each session card** shows a bad / hard-to-read path.

### Live evidence (2026-07-13)

| Surface | DOM text | `title` (full) |
|---------|----------|----------------|
| Top bar `[data-testid=workspace-label]` | `…/xhd2015/working` | `/Users/xhd2015/Projects/xhd2015/working` |
| Card `[data-testid=session-workspace]` | `…/codelens-workspace/project-api-capture` | long GOPATH path |
| Card short home | `Users/xhd2015` (no leading `/`) | `/Users/xhd2015` |

Also: `.session-item-meta { text-transform: lowercase }` forces paths to lowercase in cards.

## Root cause

| # | Mechanism | File |
|---|-----------|------|
| 1 | `shortWorkspaceLabel` always keeps **only last 2 path segments** with `…/` prefix | `frontend-agent-run/src/sessionDisplay.ts` |
| 2 | Home top bar always uses short label when `onOpenSelector` is set (no expand-to-full) | `WorkspacePath.tsx` |
| 3 | Session rows use `sessionWorkspaceLabel` → same shortener | `SessionList.tsx` |
| 4 | CSS `text-transform: lowercase` on meta row mangles path casing | `SessionList.css` |

Full path exists in API (`/api/agent-run/status`, `sessions[].workspace`) and in `title=` — **display truncation is client-side only**.

## Goal (healthy)

| Surface | Pass when |
|---------|-----------|
| Top bar | Prefer **full path** readable (wrap or single-line ellipsis from the **start**/middle, not forced 2-segment shorten). Tap still opens `/workspace`. Full path in `title` + optional Copy. |
| Session card | Prefer **tail-meaningful** label that keeps repo/worktree identity (e.g. last 1–3 segments **or** smart home/`~` prefix + tail), not opaque `…/parent/child` that drops the distinctive middle. Full path in `title`. **No forced lowercase** on path text. |
| Short paths | `/Users/xhd2015` → keep leading `/` (or `~` if under home). |

## Proposed fix (do not apply until approved)

### Option A — Prefer CSS ellipsis, keep full string (Recommended for top bar)

1. **Top bar** (`WorkspacePath` when `onOpenSelector`):
   - Show **full** `path` in the label (not `shortWorkspaceLabel`).
   - CSS: single-line `overflow: hidden; text-overflow: ellipsis; direction: rtl` **or** standard LTR ellipsis with `title=full` (user sees end of path with `direction: rtl` + unicode isolate trick, or show full with wrap on 2 lines).
   - Recommended UX: **2-line wrap** with `word-break: break-all` on home workspace row (full path always readable on mobile without tap-expand).

2. **Session cards**:
   - Replace naive last-2-segments with **smarter shortener**:
     - If path starts with `$HOME` or known prefixes (`/Users/xhd2015`, `~`), show `~/…/last2` or `~/Projects/…/name`.
     - Prefer **basename** when unique enough: `project-api-capture` with `title=full`.
     - Cap display length (e.g. 40 chars) with **middle or start** ellipsis, preserving basename.
   - Remove `text-transform: lowercase` from workspace span (keep runner lowercase if desired via separate class).

3. **Tests**: unit tests for new label helper; playwright leaf for top label contains full segments; card `title` === API workspace.

### Option B — Keep short labels, improve quality only

- Only change `shortWorkspaceLabel` algorithm (home/`~`, basename-first).
- Top bar stays short — user relies on `title` hover (bad on mobile).

**Recommend A for top bar + improved shortener for cards.**

### Concrete API sketch

```ts
// sessionDisplay.ts
formatWorkspaceLabel(path: string, opts?: {
  mode: 'full' | 'compact'
  home?: string  // from status.home
  maxChars?: number
}): string
```

- `full`: return cleaned absolute path (for top bar).
- `compact`: `~/…/basename` or last segment when path is very long; never strip leading `/` incorrectly.

## Surfaces

| Step | Command / check |
|------|-----------------|
| Inspect live | Playwright: `workspace-label` text vs `title`; session-workspace text vs title |
| Unit | `bun test` on `sessionDisplay` label helpers |
| UI | `doctest` / playwright: top shows full path or multi-line full; cards not lowercased |

## Inspect (after implement)

```sh
# Live DOM
playwright-debug run <probe> http://127.0.0.1:8821

# Expect:
# workspace-label text includes "Projects" or full path (not only …/xhd2015/working)
# session-workspace title === full path; text not forced lowercased uniquely
# shortWorkspaceLabel('/Users/xhd2015') preserves leading /
```

Optional aux: `script/debug/home-workspace-path-display/` (add when implementing).

## Step 5 — Fix (after approval only)

1. `sessionDisplay.ts` — new/updated label helpers + unit tests.
2. `WorkspacePath.tsx` — top bar uses full path when `onOpenSelector`.
3. `SessionList` / CSS — compact label + no lowercase on workspace.
4. Loop inspect GREEN.

## Dry-run log

| When | Step | Result | Evidence |
|------|------|--------|----------|
| 2026-07-13 | Live API | OK | status.workspace full; sessions[].workspace full |
| 2026-07-13 | Live DOM | **REPRO** | top `…/xhd2015/working`; cards `…/codelens-workspace/…`; meta lowercase |
| 2026-07-13 | Code | OK | shortWorkspaceLabel last-2 segments; WorkspacePath forces short when selector |
| 2026-07-13 | Fix Option A | **VERIFY** | top = full path; cards = `~/…/basename`; unit tests 40 pass |
| 2026-07-13 | Card long basename | **REPRO** | Worktree leaf names >40 chars become `…ter-2026-07-13-add-terminal-color-skill` (tail-only); unreadable. Screenshot user. |

### Follow-up REPRO — capped long basenames (card)

**Observed (user screenshot + unit sim):**

| Full worktree leaf | Current compact (`maxChars=40`) |
|--------------------|----------------------------------|
| `…/agent-pro-master-2026-07-13-add-terminal-color-skill` | `…ter-2026-07-13-add-terminal-color-skill` |
| `…/ai-critic-master-2026-07-13-grok-usage-error` | `…itic-master-2026-07-13-grok-usage-error` |

**Cause:** when basename length > `maxChars`, code does:

```ts
return `…${base.slice(-(maxChars - 1))}`  // keeps TAIL only
```

Worktree names put the distinctive repo slug at the **start** (`skills-master`, `ai-critic-master`) and a long date+topic in the middle/end — tail-only ellipsis drops the useful prefix. CSS end-ellipsis on the meta row can add a second `…`, looking double-capped. Runner also wraps (`grok-` / `tty`) under width pressure.

**Healthy (proposed):**

| Rule | Example |
|------|---------|
| Prefer **middle ellipsis on basename** keeping head + tail | `skills-mas…olor-skill` or `skills…color-skill` |
| Or worktree-aware: under `…/worktrees/<name>` show only smart-short `<name>` | `skills-master-…-color-skill` |
| Never tail-only for basenames longer than maxChars | — |
| Layout: workspace on **own row** (full width under runner · recency) so path isn’t squeezed | no runner wrap |

### Proposed fix (not applied yet)

1. **`ellipsizeMiddle(s, maxChars)`** in `sessionDisplay.ts` — used when compact basename still exceeds budget.
2. **`formatWorkspaceLabel` compact**: after `~/…/${base}`, if still long → `ellipsizeMiddle(base, maxChars)` (optionally prefix `~/…/` only if room).
3. **Optional**: detect `/.wrk/worktrees/` or `/worktrees/` and compact to that leaf only (skip intermediate segments entirely).
4. **CSS/layout** (SessionList):  
   - subhead: row1 = `runner · recency` (nowrap); row2 = full-width workspace (single-line middle-ellipsized label + `title=full`).  
   - `session-item-runner { white-space: nowrap; flex-shrink: 0 }`.
5. **Loop inspect**: unit cases for 50-char worktree basenames must not match `/^…[^/]*2026/` tail-only pattern; must include a head token like `skills` or `critic` or `master`.

Approve to implement via loop.
