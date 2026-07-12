---
title: "After /exit, status still runner.exited=false (zombie serve)"
created: 2026-07-11
slug: open-exit-runner-exited-false
path: doc/LOOP_2026-07-11_open-exit-runner-exited-false.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: open-exit-runner-exited-false

## Kind

**bug-repro** — after user `/exit` inside grok-tty (keep-alive / `--open`),
`agent-run status` still reports `runner.exited: false` and blocks resume, even
though the agent has printed `[Terminal exited]` / `grok --resume <id>` and the
serve process has **no child**.

## Symptom (verbatim)

```text
session:   grok-tty/test-open-v6
status:    running
...
process:
  status:  alive
  pid:     763
  kind:    serve

terminal:
  status:   reachable
  ...
  sendable: no

runner:
  status:     bound
  session_id: 019f511a-...
  exited:     false

resume:
  ready: no
  reason: runner still active (exited: false); use send, not resume
```

Meanwhile open/attach scrollback shows:

```text
Resume this session with:
  grok --resume 019f511a-...
[Terminal exited]
```

### Preconditions

- Real `grok` on PATH; project workspace (not bare `/tmp` picker).
- Session started with `run --open` or keep-alive so **serve stays up** after child exits.
- Current status logic (see Fix): `computeRunnerExited` returns false whenever
  `terminal.status == reachable`.

### Root cause (for Fix step only; do not apply during establish)

`cmd/agent-run/status.go` `computeRunnerExited`:

```go
if report.Terminal.Status == "reachable" {
    return &f // not exited
}
```

Keep-alive `__serve__` stays TCP-reachable after grok child exits → false negative
for `exited`. Resume gate requires `exited==true`.

## Goal

Steps 1–4 reproduce: after open + `/exit`, status shows **exited:false** +
resume blocked while terminal still reachable / process alive.

## Surfaces

| Step | Surface |
|------|---------|
| Build | `go build ./cmd/agent-run` |
| Deploy | binary path for inspect |
| Run | open + send `/exit` (inspect script) |
| Inspect | `script/debug/open-exit-runner-exited` exit 1 + `REPRO:` |

## Pitfalls

| Issue | Notes |
|-------|-------|
| Project dir picker on `/tmp` | Use real project cwd for open |
| Instant attach | Script sets `AGENT_RUN_OPEN_ATTACH_INSTANT=1` |
| Need bound before /exit | Script waits for `runner_session_id` |
| Leftover serve | Kill registry pid after smoke |
| Soft bind fail | Open bind must succeed first (use project dir) |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"
mkdir -p frontend-agent-run/dist frontend/dist
test -f frontend-agent-run/dist/index.html || echo '<!doctype html><title>stub</title>' > frontend-agent-run/dist/index.html
BIN="${TMPDIR:-/tmp}/agent-run-open-exit-loop"
go build -o "$BIN" ./cmd/agent-run
test -x "$BIN"
```

**Verify:** `$BIN --help` lists `status`, `run`, `resume`.

---

## Step 2 — Deploy

```sh
export LOOP_AGENT_RUN="${TMPDIR:-/tmp}/agent-run-open-exit-loop"
# or use reinstalled CLI:
# export LOOP_AGENT_RUN="$(which agent-run)"
ls -la "$LOOP_AGENT_RUN"
```

**Verify:** binary executable.

---

## Step 3 — Run (trigger only)

No remedial status healing. Inspect script performs:

1. `agent-run run --open --agent-runner=grok-tty …` (instant attach)
2. Wait until bound
3. `agent-run send <id> /exit`
4. `agent-run status <id>`

---

## Step 4 — Inspect REPRO

```sh
cd "$(git rev-parse --show-toplevel)"
BIN="${LOOP_AGENT_RUN:-$(which agent-run)}"
go run ./script/debug/open-exit-runner-exited --agent-run "$BIN"
# expect: exit 1
# expect: REPRO: after /exit status reports runner.exited=false ...
```

**REPRO criteria:**

- Exit code **1**
- `REPRO: after /exit status reports runner.exited=false while terminal still reachable`
- And/or `REPRO: resume.ready=no reason claims runner still active`
- Evidence: process alive, sendable no, child gone (optional)

---

## Step 4b — Inspect VERIFY (post-fix only)

```sh
go run ./script/debug/open-exit-runner-exited --agent-run "$BIN" --expect=healthy
# expect: exit 0, VERIFY: PASS after /exit runner.exited=true ...
```

---

## Step 5 — Fix (not during establish)

When REPRO is solid:

1. **`computeRunnerExited`:** do not treat TCP-reachable alone as “not exited”.
   - Prefer: no child process under serve PID → exited true
   - Snapshot contains `[Terminal exited]` or `grok --resume` footer → exited true
   - `sendable: no` + resume footer → exited true
   - Keep `exited: false` only when agent child alive or clearly idle/sendable TUI
2. Optionally update session `meta.status` to `finished` when exit detected
3. Resume gate then becomes ready when bound + exited

Hotspot: `cmd/agent-run/status.go` `computeRunnerExited`.

Return to step 1 after each fix; use step 4b for GREEN.

---

## Dry-run log

| When (UTC) | Step | Result | Evidence |
|------------|------|--------|----------|
| 2026-07-11T12:17:11Z | 1 Build | OK | `go build -o $TMPDIR/agent-run-open-exit-loop` |
| 2026-07-11T12:17:11Z | 2 Deploy | OK | built + installed paths listed |
| 2026-07-11T12:17:28Z | 3+4 Inspect REPRO | **REPRO PASS** | open+/exit → `exited=false`, terminal reachable, process alive, child_gone=true, resume blocked "still active" |
| 2026-07-11T12:17:28Z | dry_run_status | **REPRO PASS** | Matches user test-open-v6 symptom |
| 2026-07-11T12:30:48Z | 5 Fix + 4b VERIFY | **VERIFY PASS** | `computeRunnerExited` uses exit markers/child-gone; smoke exited=true resume.ready=yes; doctest 26/26 |
| 2026-07-11T12:30:48Z | dry_run_status | **VERIFY PASS** | Installed to GOPATH/bin/agent-run |

---

## Cleanup

```sh
# kill leftover smoke serve by registry pid under smoke_home if needed
pkill -f 'smoke-open-exit-exited' 2>/dev/null || true
```
