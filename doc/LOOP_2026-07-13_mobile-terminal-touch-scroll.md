---
title: "Mobile terminal modal touch-scroll (xterm)"
created: 2026-07-13
slug: mobile-terminal-touch-scroll
path: doc/LOOP_2026-07-13_mobile-terminal-touch-scroll.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: mobile-terminal-touch-scroll

## Kind

**bug-repro** — prove that on a **mobile viewport**, a **finger-style touch pan**
on the terminal popup does **not** scroll xterm history (unlike macOS
trackpad/wheel). After DIY touch→`scrollLines`, use `--expect=healthy`.

## Symptom

User context: session terminal popup (e.g.
`http://127.0.0.1:8821/sessions/web_da6177509c41c080`). On Mac, pointer/trackpad
scrolls the terminal. On phone, touch pan should scroll history the same way —
today it does not (xterm v6 wheel-centric + no app touch bridge).

**Observed (inspect REPRO):**

```text
REPRO: mobile touch pan does not scroll terminal history
REPRO: touch pan did not reveal older scrollback (min LINE before=… after=…)
```

Probe uses mobile viewport `390×844`, opens Terminal modal, streams
`LINE_000`…`LINE_119` into the PTY, then dispatches synthetic touch pan events
on `[data-testid="terminal-surface"]`. Symptom = visible min `LINE_xxx` does
**not** decrease after pan (older history not revealed).

### Preconditions (do not “fix” in steps 1–4)

- Product code as-is (do **not** add touch→`scrollLines` in steps 1–4).
- `go`, `bun`, `playwright-debug` on `PATH`.
- Run from repo root.
- Isolated `AGENT_RUN_HOME` + fake ptywrap (orchestrator).

## Goal

Steps 1–4 **reproduce** broken mobile touch-scroll (**REPRO PASS**).
Step 5 is fix guidance; iterate with `/loop-workflow` until:

```sh
go run ./script/debug/mobile-terminal-touch-scroll --expect=healthy
```

exits 0 (**VERIFY**).

### Healthy criteria

| Check | Pass when |
|-------|-----------|
| Mobile mode | Viewport 390×844 |
| Terminal open | `.xterm` + `LINE_xxx` markers visible |
| Touch pan | Synthetic vertical pan on terminal surface |
| Scroll | `min(LINE_*)` after pan **&lt;** before (older history visible) |

## Surfaces

| Step | Surface |
|------|---------|
| Build | `bun run build` in `frontend-agent-run`; `go build` `agent-run` + inspect |
| Deploy | Local `agent-run web` + fake ptywrap + seeded `codex-tty` session |
| Run | Playwright mobile: session → Terminal → touch pan |
| Inspect | `script/debug/mobile-terminal-touch-scroll` exit **1** + `REPRO:` |

## Pitfalls & blockers

| Issue | Notes |
|-------|-------|
| Auth | Open API (no token) in orchestrator |
| Terminal unavailable | Registry `codex-tty-registry` + `terminal_session_id` required |
| Nested sessions | Seed **flat** `sessions/<id>/` + `.layout` v2 |
| playwright-debug | Must be on `PATH` |
| Do not fix during REPRO | No `TerminalModal` touch handlers in steps 1–4 |
| Alt buffer | Fake PTY uses normal buffer with scrollback lines |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"

(cd frontend-agent-run && bun run build)
test -f frontend-agent-run/dist/index.html

OUT="${TMPDIR:-/tmp}/mobile-terminal-touch-scroll-build"
mkdir -p "$OUT/bin"
go build -o "$OUT/bin/agent-run" ./cmd/agent-run
go build -o /dev/null ./script/debug/mobile-terminal-touch-scroll
test -x "$OUT/bin/agent-run"
```

**Verify:** builds exit 0; dist + binary exist.

---

## Step 2 — Deploy / Update

Local only. Confirm tools; orchestrator starts fake PTY + web:

```sh
which go bun playwright-debug
# Orchestrator:
#   fake ptywrap WS on ephemeral port (LINE_000..LINE_119)
#   AGENT_RUN_HOME with flat session web_term_touch_scroll + codex-tty-registry
#   agent-run web --no-open
```

**Verify:** `playwright-debug` and `bun` on `PATH`.

---

## Step 3 — Run

```sh
cd "$(git rev-parse --show-toplevel)"

OUT="${TMPDIR:-/tmp}/mobile-terminal-touch-scroll-$(date +%s)"
go run ./script/debug/mobile-terminal-touch-scroll -out="$OUT"
# expected exit 1 in repro mode when symptom present
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
print("reason", (r.get("symptom") or {}).get("reason"))
print("beforeMin", (r.get("symptom") or {}).get("beforeMin"))
print("afterTouchMin", (r.get("symptom") or {}).get("afterTouchMin"))
print("screenshots", r.get("screenshots"))
PY
ls -la "$OUT/probe"/*.png
```

**Verify (REPRO):** `symptomPresent=true`, exit code 1.

### Step 4b — Inspect VERIFY (after fix only)

```sh
go run ./script/debug/mobile-terminal-touch-scroll --expect=healthy -out="$OUT-verify"
```

**Verify (VERIFY):** exit 0, stdout contains
`VERIFY: mobile touch pan scrolls terminal history`.

---

## Step 5 — Fix (only after REPRO PASS; not during establishment)

1. Add pure helper `touchDeltaToScrollLines` (optional unit tests).
2. In `TerminalModal`, after `term.open`, attach `touchstart`/`touchmove`/`touchend`
   → `term.scrollLines(n)` with cell-height residual.
3. CSS: `touch-action: pan-y` on `.terminal-surface`.
4. Re-run:

```sh
go run ./script/debug/mobile-terminal-touch-scroll --expect=healthy
```

Return to step 1 after product changes (frontend rebuild is inside orchestrator).

---

## Dry-run log

| Timestamp | Step | Result | Evidence |
|-----------|------|--------|----------|
| 2026-07-13 ~08:43 | 1 Build | PASS | frontend + agent-run + inspect compile OK |
| 2026-07-13 ~08:43 | 2 Deploy | PASS | fake ptywrap `127.0.0.1:63740`, web `http://127.0.0.1:63749`, terminal available |
| 2026-07-13 ~08:43 | 3 Run | PASS | mobile 390×844; Terminal open; LINE_097–119 visible |
| 2026-07-13 ~08:43 | 4 Inspect | **REPRO** (exit 1) | touch pan min LINE stayed 97→97; wheel baseline **did** move 97→95 (`wheelMoved: true`) |
| 2026-07-13 | Dry-run status | **REPRO PASS** | `/tmp/mobile-terminal-touch-scroll-dryrun-1783903424/` |
| 2026-07-13 ~08:45 | 5 Fix + VERIFY | **VERIFY PASS** (exit 0) | DIY touch→`scrollLines`; min LINE **97 → 76** after touch pan; unit tests 58/58 |

---

## Cleanup

```sh
rm -rf /tmp/mobile-terminal-touch-scroll-*
```
