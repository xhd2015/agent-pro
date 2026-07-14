---
title: "Session list UX: load-more placement, no auto-load, preserve scroll"
created: 2026-07-13
slug: session-list-ux-enhance
path: doc/LOOP_2026-07-13_session-list-ux-enhance.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: session-list-ux-enhance

## Kind

**bug-repro** — prove three home session-list UX defects on
`http://127.0.0.1:8821/` (and isolated seeded web). After product fix, use
`--expect=healthy` for VERIFY.

## Symptom (verbatim class)

User request (enhance session list UX):

1. **Load more** should only appear at the **end of the session list**.
2. When the user **scrolls**, do **not** auto-load; user must **explicitly click**
   Load more.
3. When the user **enters a session and goes back**, the session list **scroll
   state** should be **preserved** — lift session list state up to the toplevel
   **App** entrance.

**Observed (current product, live + seeded):**

| # | Defect | Evidence |
|---|--------|----------|
| 1 | Load more sits **outside** `[data-testid="session-list"]` (sibling under `.session-list-region`), always visible at `scrollTop≈0` | `btnInsideList=false`, `btnVisibleAtTop=true` |
| 2 | Near-bottom scroll **auto-appends** the next page | item count `30 → 60` without button click |
| 3 | Session → Back remounts `HomePage`; list resets | mid `scrollTop` → `0`, loaded pages lost |

Inspect evidence (orchestrator default / repro mode):

```text
REPRO: session list UX defects present (load-more placement / auto-load / scroll preserve)
REPRO: load more button is not at the end of the session list scroll content
REPRO: scroll near bottom auto-loaded pages (30 -> 60)
REPRO: session enter/back lost scroll (saved=… after=0)
```

### Preconditions (do not “fix” in steps 1–4)

- Product code as-is. Do **not** during steps 1–4:
  - move Load more inside the list / scroll container
  - remove `handleListScroll` auto `loadMore()`
  - lift HomePage session-list state to `App.tsx`
- `go`, `bun`, `playwright-debug` on `PATH`.
- Run from repo root.
- Isolated `AGENT_RUN_HOME` with **≥55** seeded sessions (orchestrator) so
  `PAGE_SIZE=30` yields `has_more` and a scrollable list.

### Optional live URL

```sh
go run ./script/debug/session-list-ux-enhance -url=http://127.0.0.1:8821
```

(Live server must already have `total > 30` sessions.)

## Goal

Steps 1–4 **reproduce** the three UX symptoms on command (**REPRO PASS**).
Step 5 is fix guidance; iterate with `/run-the-loop` until:

```sh
go run ./script/debug/session-list-ux-enhance --expect=healthy
```

exits 0 (**VERIFY**).

### Healthy criteria

| Check | Pass when |
|-------|-----------|
| Load more placement | Button is **inside** the session-list scroll content (or a shared scroll parent with list items), **not** a sticky sibling always visible at `scrollTop=0` while the list overflows |
| No auto-load | Scrolling to list bottom **does not** increase `session-item` count; only `[data-testid="session-load-more"]` click loads next page |
| Scroll preserve | After open session → `.back-link` → `/`, `scrollTop` stays within ~60px of pre-nav value; loaded item count not reset to first page only |

## Surfaces

| Step | Surface |
|------|---------|
| Build | `bun run build` in `frontend-agent-run`; `go build` `agent-run` + inspect |
| Deploy | Local `agent-run web` + seeded flat `sessions/` (≥55) |
| Run | Playwright mobile 390×844: home list scroll / load-more / session back |
| Inspect | `script/debug/session-list-ux-enhance` exit **1** + `REPRO:` |

## Pitfalls & blockers

| Issue | Notes |
|-------|-------|
| Auth | Open API (no token) in orchestrator |
| Too few sessions | Need `total > PAGE_SIZE` (30) or load-more / scroll checks are void |
| Nested sessions layout | Seed **flat** `sessions/<id>/` + `.layout` v2 |
| Poll noise | Home polls `/sessions?offset=0`; auto-load symptom uses **item count growth** and `offset>0` |
| Live :8821 assets | Product changes need rebuild/restart of the running web binary |
| Do not fix during REPRO | No HomePage/App state lift in steps 1–4 |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"

(cd frontend-agent-run && bun run build)
test -f frontend-agent-run/dist/index.html

OUT="${TMPDIR:-/tmp}/session-list-ux-enhance-build"
mkdir -p "$OUT/bin"
go build -o "$OUT/bin/agent-run" ./cmd/agent-run
go build -o /dev/null ./script/debug/session-list-ux-enhance
test -x "$OUT/bin/agent-run"
```

**Verify:** builds exit 0; dist + binary exist.

---

## Step 2 — Deploy / Update

Local only. Orchestrator seeds home + starts web (unless `-url=`):

```sh
which go bun playwright-debug
# Orchestrator:
#   AGENT_RUN_HOME with flat sessions home-list-001..055
#   agent-run web --no-open --agent-runner fake-codex
```

**Verify:** `playwright-debug` and `bun` on `PATH`.

---

## Step 3 — Run

```sh
cd "$(git rev-parse --show-toplevel)"

OUT="${TMPDIR:-/tmp}/session-list-ux-enhance-$(date +%s)"
go run ./script/debug/session-list-ux-enhance -out="$OUT"
# expected exit 1 in repro mode when symptoms present
```

**Verify:** exit **1** with `REPRO:` lines (not exit 2 infrastructure fail).
Screenshots under `$OUT/probe/*.png`.

---

## Step 4 — Inspect / Feedback

```sh
test -f "$OUT/probe-report.json"
python3 - <<'PY' "$OUT/probe-report.json"
import json,sys
r=json.load(open(sys.argv[1]))
print("symptomPresent", r.get("symptomPresent"))
print("okHealthy", r.get("okHealthy"))
print("reasons", r.get("reasons"))
print("symptom", r.get("symptom"))
print("screenshots", r.get("screenshots"))
PY
ls -la "$OUT/probe"/*.png
```

**Verify (REPRO):** `symptomPresent=true`, `okHealthy=false`, reasons cover placement
and/or auto-load and/or scroll-lost.

### Step 4b — Inspect Verify (post-fix only)

```sh
go run ./script/debug/session-list-ux-enhance --expect=healthy
# exit 0 + VERIFY: lines
```

---

## Step 5 — Fix (do not apply during loop establishment)

When REPRO is green (symptoms present), implement product changes and re-run
with `--expect=healthy`:

1. **Load more at list end**
   - Put `[data-testid="session-load-more"]` **inside** the scrollable list
     content (last child of `[data-testid="session-list"]`, or move
     `overflow-y: auto` to a wrapper that includes items + button).
   - User should scroll to the end of the list to reach the button (not a sticky
     footer always visible while browsing mid-list).

2. **No auto-load on scroll**
   - In `HomePage` `handleListScroll`, remove near-bottom `loadMore()` call.
   - Keep explicit `onClick` on the Load more button only.

3. **Preserve scroll + list state across session detail**
   - Lift session-list state (sessions, hasMore, total, counts, filters, search,
     scrollTop / follow mode) to **toplevel `App`** (same pattern as
     `homeDraft`), pass into `HomePage` as controlled props / context.
   - On remount after `navigate('/')` from `SessionPage`, restore
     `sessionListRef.scrollTop` (or keep Home mounted via layout route if
     preferred — user asked for lift to App entrance).

Primary files:

- `frontend-agent-run/src/App.tsx`
- `frontend-agent-run/src/pages/HomePage.tsx`
- `frontend-agent-run/src/pages/HomePage.css`
- `frontend-agent-run/src/components/SessionList.tsx` / `SessionList.css`

Return to step 1 after edits (rebuild embeds frontend).

---

## Dry-run log

| When (UTC-ish local) | Step | Result | Evidence |
|------|------|--------|----------|
| 2026-07-13 ~08:52 | 1 Build | OK | `bun run build` + `go build` agent-run + inspect compile |
| 2026-07-13 ~08:52 | 2 Deploy | OK | tools on PATH; orchestrator seeds 55 sessions + ephemeral web |
| 2026-07-13 ~08:52 | 3 Run | OK | `go run ./script/debug/session-list-ux-enhance` exit **1** |
| 2026-07-13 ~08:52 | 4 Inspect | **REPRO** | `symptomPresent=true`; reasons: load-more not at list end; auto-load `30→55`; scroll lost `2520→0` |
| 2026-07-13 | Status | **REPRO PASS** | All three user symptoms reproduced without applying step 5 |
| 2026-07-13 ~09:10 | 5 Fix + 4b | **VERIFY PASS** | `go run ./script/debug/session-list-ux-enhance --expect=healthy` exit 0; scroll 1287→1287; load-more inside list; no auto-load; click loads 30→55 |
| 2026-07-13 | snap-back fix | **VERIFY PASS** | Multi-step A→B→C + 3.5s settle stays at C (no jump to B); pin no longer restored on momentum scroll; restore effect only on isActive edge |
| 2026-07-13 | poll traffic | **VERIFY PASS** | Idle 7s: runnersΔ=0 statusΔ=0 (bootstrap only once); sessions poll is sessions-only, 5s when running / 15s idle, visibility-gated |

---

## Handoff

After **REPRO PASS**, run:

```text
/run-the-loop fix session list UX: load-more at list end, no auto-load on scroll, lift state to App for scroll preserve
```

GREEN gate:

```sh
go run ./script/debug/session-list-ux-enhance --expect=healthy
```
