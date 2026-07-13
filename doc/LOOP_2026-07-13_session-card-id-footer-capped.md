---
title: "Session card bottom: mid-ellipsis session id footer"
created: 2026-07-13
slug: session-card-id-footer-capped
path: doc/LOOP_2026-07-13_session-card-id-footer-capped.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: session-card-id-footer-capped

## Kind

**bug-repro** — prove home session cards with a prompt still show a **bottom
footer** of mid-ellipsis `shortSessionId` (looks capped/junk). After product fix,
use `--expect=healthy`.

## Symptom (user)

On home (`http://127.0.0.1:8821/`), **bottom area of a session card** (not the
workspace path mid-line) shows a capped id, e.g. screenshot:

```text
brainstorm…3-083040   (under prompt title + runner/path row)
```

Full id is long and human-readable, e.g.:

```text
brainstorm-add-terminal-color-skill-20260713-083040
```

**Product mechanism (unfixed):**

```tsx
// SessionList.tsx — when hasPrompt
<span className="session-item-id" title={s.session_id}>
  {shortSessionId(s.session_id)}  // first10 + … + last8 if len>20
</span>
```

### Preconditions (do not fix in steps 1–4)

- Product as-is: keep `.session-item-id` + `shortSessionId` for prompt cards.
- `go`, `bun`, `playwright-debug` on `PATH`.
- Isolated seed session with **long** `session_id` + non-empty `initial_prompt`.

## Goal

Steps 1–4 **reproduce** the capped id footer (**REPRO PASS**, exit 1).

After step 5 (fix), healthy gate:

```sh
go run ./script/debug/session-card-id-footer-capped --expect=healthy
```

### Healthy criteria

| Check | Pass when |
|-------|-----------|
| Prompt card | Preview shows human prompt |
| Footer | No visible mid-ellipsis `.session-item-id` for long human-readable ids **or** footer removed when prompt present |
| Opaque ids | Optional: short id only for `web_*` / short opaque ids (product choice) |

## Surfaces

| Step | Surface |
|------|---------|
| Build | `bun run build` frontend-agent-run; `go build` agent-run + inspect |
| Deploy | Local `agent-run web` + seeded flat session |
| Run | Playwright mobile home list |
| Inspect | exit **1** + `REPRO:` when footer capped |

## Pitfalls & blockers

| Issue | Notes |
|-------|-------|
| Path vs id | Mid-row workspace is **not** this bug; assert `.session-item-id` only |
| Live :8821 | May need seed with long id; prefer orchestrator seed |
| hasPrompt | Footer only renders when `initial_prompt` non-empty |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"
(cd frontend-agent-run && bun run build)
test -f frontend-agent-run/dist/index.html
go build -o /dev/null ./cmd/agent-run
go build -o /dev/null ./script/debug/session-card-id-footer-capped
```

**Verify:** exit 0.

---

## Step 2 — Deploy / Update

Orchestrator seeds + starts web (or `-url=` for live):

```sh
which go bun playwright-debug
```

**Verify:** tools on PATH.

---

## Step 3 — Run

```sh
cd "$(git rev-parse --show-toplevel)"
OUT="${TMPDIR:-/tmp}/session-card-id-footer-capped-$(date +%s)"
go run ./script/debug/session-card-id-footer-capped -out="$OUT"
# expected exit 1 in repro mode when symptom present
```

**Verify:** exit **1**, not 2.

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
PY
```

**Verify (REPRO):** `symptomPresent=true`, reasons mention capped session id footer.

### Step 4b — Verify (post-fix)

```sh
go run ./script/debug/session-card-id-footer-capped --expect=healthy
# exit 0
```

---

## Step 5 — Fix (not during establishment)

When REPRO is green:

1. **Preferred:** do not render `.session-item-id` when `sessionRowHasPrompt` (prompt already identifies the card; full id remains in link/`title` if needed).
2. **Or:** only show short id for opaque ids (`web_*`, hex-like); leave long slug ids off the footer.
3. **Or:** stop using mid-ellipsis for ids longer than N; monoline CSS ellipsis with full id in `title` only.

Files: `SessionList.tsx`, optionally `sessionDisplay.ts` `shortSessionId` callers.

Return to step 1 after product change.

---

## Dry-run log

| When | Step | Result | Evidence |
|------|------|--------|----------|
| 2026-07-13 | 1 Build | OK | frontend + agent-run + inspect compile |
| 2026-07-13 | 2 Deploy | OK | tools on PATH; orchestrator seeds long id + prompt |
| 2026-07-13 | 3 Run | OK | inspect exit **1** |
| 2026-07-13 | 4 Inspect | **REPRO** | `idText=brainstorm…3-083040` under preview `/brainstorm add terminal color skill` |
| 2026-07-13 | Status | **REPRO PASS** | Bottom footer mid-ellipsis session id reproduced without fix |
| 2026-07-13 | 5 Fix + 4b | **VERIFY PASS** | Removed `.session-item-id` under prompt; `idText` gone; `--expect=healthy` exit 0 |

---

## Handoff

```text
/loop-workflow fix session card: hide or fix capped mid-ellipsis session-id footer under prompt
```

GREEN:

```sh
go run ./script/debug/session-card-id-footer-capped --expect=healthy
```
