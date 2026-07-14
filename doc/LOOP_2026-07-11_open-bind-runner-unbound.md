---
title: "Open bind: runner unbound after run --open (real grok)"
created: 2026-07-11
slug: open-bind-runner-unbound
path: doc/LOOP_2026-07-11_open-bind-runner-unbound.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: open-bind-runner-unbound

## Kind

**bug-repro** — reproduce user-reported failure where after `agent-run run --open`
with real `grok-tty`, `agent-run status <id>` shows `runner: unbound` and
`resume.ready: no` (missing `runner_session_id`).

## Symptom (verbatim class)

User status (and smoke) looks like:

```text
runner:
  status:     unbound
  kind:       grok
  session_id:
  exited:     false

resume:
  ready: no
  reason: runner session not bound (missing runner_session_id)
```

Additional smoke evidence (bind soft path):

```text
bind.json: {"state":"failed",...,"error":"context deadline exceeded"}
open exit 0, stderr only: grok-tty: <session-id>   # no "grok session" line
```

### Preconditions (do not “fix” in steps 1–4)

- Real `grok` on PATH (`~/.local/bin/grok` or equivalent).
- **Do not** set `GROK_HOME` / `AGENT_RUNNER_CONFIG_HOME` (user default soft path).
- Isolated `AGENT_RUN_HOME` (inspect script creates temp home).
- Use workspace under temp (not only doctest fixtures with pre-seeded GROK_HOME).

### Why doctests alone missed this

Doctests seed `GROK_HOME` + updates early → hard path / long budget / preseed
success. Real smoke uses **default** `~/.grok` with **soft 750ms** discovery →
deadline exceeded → silent unbound, exit 0.

## Goal

Steps 1–4 **reproduce** unbound after `--open` on command (REPRO PASS).
Step 5 is fix guidance only during establishment; full fix via `/run-the-loop`.

## Surfaces

| Step | Surface |
|------|---------|
| Build | `go build -o <bin> ./cmd/agent-run` (ensure frontend dist stubs if needed) |
| Deploy | install/use built binary path for smoke (`--agent-run`) |
| Run | `agent-run run --open --agent-runner=grok-tty` (via inspect script) |
| Inspect | `script/debug/open-bind-smoke` exit **1** + `REPRO:` lines |

## Pitfalls & blockers

| Issue | Notes |
|-------|-------|
| Missing `frontend-agent-run/dist` | Build fails embed; ensure stub dist before build |
| Interactive attach hangs | Smoke sets `AGENT_RUN_OPEN_ATTACH_INSTANT=1` |
| Leftover keep-alive serve | Kill processes matching session id after smoke |
| Soft vs hard path | Unset GROK_HOME to match user; setting it hides the bug |
| Doctest green ≠ smoke green | Full tree can pass while soft path fails |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"
# embed stubs if missing
mkdir -p frontend-agent-run/dist frontend/dist
test -f frontend-agent-run/dist/index.html || echo '<!doctype html><title>stub</title>' > frontend-agent-run/dist/index.html
test -f frontend/dist/index.html || echo '<!doctype html><title>stub</title>' > frontend/dist/index.html

BIN="${TMPDIR:-/tmp}/agent-run-open-bind-loop"
go build -o "$BIN" ./cmd/agent-run
test -x "$BIN"
```

**Verify:** `test -x "$BIN"` and `$BIN --help` lists `run` and `status`.

---

## Step 2 — Deploy / Update

Use the built binary for smoke (no system install required):

```sh
export LOOP_AGENT_RUN="${TMPDIR:-/tmp}/agent-run-open-bind-loop"
"$LOOP_AGENT_RUN" version 2>/dev/null || "$LOOP_AGENT_RUN" --help | head -5
```

Optional install (only if iterating like a user binary):

```sh
# optional — not required for inspect script
# cp "$LOOP_AGENT_RUN" "$(go env GOPATH)/bin/agent-run"
```

**Verify:** `$LOOP_AGENT_RUN` exists and is executable.

---

## Step 3 — Run (trigger failure only)

Do **not** pre-seed GROK_HOME session files. Do **not** set GROK_HOME.

Smoke is driven by the inspect script (step 4), which itself runs:

```text
agent-run run --session-id=… --agent-runner=grok-tty --open --dir <ws> "one word of France capital"
```

with:

- `AGENT_RUN_HOME=<temp>`
- `AGENT_RUN_OPEN_ATTACH_INSTANT=1`
- **no** `GROK_HOME`

Manual equivalent (optional):

```sh
export AGENT_RUN_HOME="$(mktemp -d)"
export AGENT_RUN_OPEN_ATTACH_INSTANT=1
unset GROK_HOME AGENT_RUNNER_CONFIG_HOME
"$LOOP_AGENT_RUN" run --session-id=smoke-manual \
  --agent-runner=grok-tty --open --dir /tmp \
  "one word of France capital"
"$LOOP_AGENT_RUN" status smoke-manual
cat "$AGENT_RUN_HOME/sessions/grok-tty/smoke-manual/bind.json" 2>/dev/null || true
cat "$AGENT_RUN_HOME/sessions/grok-tty/smoke-manual/meta.json"
```

**Verify (manual):** status shows unbound / meta missing `runner_session_id`.

---

## Step 4 — Inspect / Feedback (symptom present = REPRO)

```sh
cd "$(git rev-parse --show-toplevel)"
BIN="${LOOP_AGENT_RUN:-${TMPDIR:-/tmp}/agent-run-open-bind-loop}"
go run ./script/debug/open-bind-smoke --agent-run "$BIN"
# expect: exit 1
# expect: lines starting with REPRO:
```

**REPRO criteria (all acceptable evidence):**

- Exit code **1** from inspect script
- Printed `REPRO: runner_session_id missing after agent-run run --open`
- And/or `REPRO: bind.json state=failed error=context deadline exceeded`
- And/or `REPRO: status reports unbound`

**Not REPRO:** exit 0 with `REPRO: FAIL symptom absent — session already bound`.

---

## Step 4b — Inspect verify (post-fix GREEN only)

Use only after step 5 fix is applied:

```sh
go run ./script/debug/open-bind-smoke --agent-run "$BIN" --expect=healthy
# expect: exit 0, VERIFY: PASS bound after --open
```

---

## Step 5 — Fix (guidance for `/run-the-loop`; do not apply during establish)

When REPRO is reliable, fix candidates:

1. **Soft path budget:** default open + real grok should use hard discovery budget
   (e.g. 20s), not `openGrokBindSoftTimeout` (750ms), whenever prompt is non-empty.
2. **`openGrokDiscoveryRequired`:** treat default `GrokHome()` + non-empty prompt as
   require-bind for open (exit ≠ 0 if unresolved after wait).
3. **Start discovery after inject delay** or poll until `updates.jsonl` appears;
   750ms from open start races real grok session materialization.
4. After fix: re-run steps 1–2, then step 4b `--expect=healthy` for VERIFY PASS.

Code hotspots:

- `pkgs/agentui/open_bind.go` — `openGrokBindSoftTimeout`, `startOpenGrokBindWorker`
- `pkgs/agentui/run.go` — `openGrokDiscoveryRequired`, open bind join
- `pkgs/agenttty` — `DiscoverSession` matching workspace/prompt

Return to step 1 after each fix attempt.

---

## Dry-run log

| When (UTC) | Step | Result | Evidence |
|------------|------|--------|----------|
| 2026-07-11T11:52:36Z | 1 Build | OK | `go build -o $TMPDIR/agent-run-open-bind-loop ./cmd/agent-run` |
| 2026-07-11T11:52:36Z | 2 Deploy | OK | binary executable; help lists run/status |
| 2026-07-11T11:52:40Z | 3+4 Inspect REPRO | **REPRO PASS** | `go run ./script/debug/open-bind-smoke` exit 1; `bind_state=failed` `error=context deadline exceeded`; status `unbound`; open exit 0 without `grok session` line |
| 2026-07-11T11:52:40Z | dry_run_status | **REPRO PASS** | Soft path 750ms timeout is the failure signature |
| 2026-07-11T11:56:44Z | 5 Fix + re-inspect | FAIL then fix | Soft 750ms + worker died during attach; /tmp hits grok dir picker |
| 2026-07-11T11:57:10Z | 4b VERIFY | **VERIFY PASS** | `open-bind-smoke --expect=healthy`: `runner_session_id` bound, stderr `grok session`, status bound; doctest 22/22 |
| 2026-07-11T11:57:50Z | dry_run_status | **VERIFY PASS** | Installed to `$(go env GOPATH)/bin/agent-run` |

---

## Cleanup

```sh
pkill -f 'smoke-open-bind' 2>/dev/null || true
pkill -f 'agent-run-open-bind-loop' 2>/dev/null || true
rm -f "${TMPDIR:-/tmp}/agent-run-open-bind-loop"
```
