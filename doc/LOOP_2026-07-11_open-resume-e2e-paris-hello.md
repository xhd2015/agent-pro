---
title: "E2E: open Paris → /exit → exited → resume --open hello"
created: 2026-07-11
slug: open-resume-e2e-paris-hello
path: doc/LOOP_2026-07-11_open-resume-e2e-paris-hello.md
loop_kind: health-check
dry_run_status: "VERIFY PASS"
---

# LOOP: open-resume-e2e-paris-hello

## Kind

**health-check** — full real-grok path for open, wait for answer, exit, resume open.

## Goal

On command, without human interaction:

1. `agent-run run --session-id=test-open-v8 --agent-runner=grok-tty --open "one word of France capital"`
2. Wait until **snapshot** contains **Paris**
3. `agent-run send … /exit`
4. Wait until **status** shows **`exited: true`**
5. `agent-run resume test-open-v8 --open "hello"`
6. **Snapshot** after resume shows proper text (followup `hello` and/or interactive UI)

## Preconditions

- Real `grok` on PATH
- Built/installed `agent-run` with open-bind, exited-after-exit, resume reclaim
- Project workspace (script uses `cwd` — not bare `/tmp`, avoids project-directory picker)
- Codesign valid (`codesign -s - -f` after install if `Killed: 9`)

## Surfaces

| Step | Surface |
|------|---------|
| Build | `go build ./cmd/agent-run` |
| Deploy | binary path + optional codesign |
| Run+Inspect | `script/debug/open-resume-e2e` |
| Fix | reclaim/bind/exited regressions |

## Pitfalls

| Issue | Mitigation |
|-------|------------|
| Project dir picker on `/tmp` | Use project cwd |
| Interactive attach hangs | `AGENT_RUN_OPEN_ATTACH_INSTANT=1` in inspect |
| SIGKILL on go/bin | `codesign -s - -f "$(which agent-run)"` |
| Zombie already-in-use on resume | reclaim fix; inspect fails if still broken |
| Leftover serves | inspect kills registry PIDs under smoke home |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"
mkdir -p frontend-agent-run/dist frontend/dist
test -f frontend-agent-run/dist/index.html || echo '<!doctype html><title>stub</title>' > frontend-agent-run/dist/index.html
BIN="${TMPDIR:-/tmp}/agent-run-open-resume-e2e"
go build -o "$BIN" ./cmd/agent-run
test -x "$BIN"
```

**Verify:** `$BIN --help` lists `run`, `resume`, `status`, `snapshot`, `send`.

---

## Step 2 — Deploy

```sh
export LOOP_AGENT_RUN="${TMPDIR:-/tmp}/agent-run-open-resume-e2e"
# or reinstall:
# cp "$LOOP_AGENT_RUN" "$(go env GOPATH)/bin/agent-run"
# codesign -s - -f "$(go env GOPATH)/bin/agent-run"
ls -la "$LOOP_AGENT_RUN"
codesign -vv "$LOOP_AGENT_RUN" 2>&1 | head -3
```

**Verify:** binary executable; codesign valid on disk (or re-sign).

---

## Step 3 — Run + Step 4 — Inspect (healthy)

Single automated script runs the full user flow:

```sh
cd "$(git rev-parse --show-toplevel)"
BIN="${LOOP_AGENT_RUN:-$(which agent-run)}"
go run ./script/debug/open-resume-e2e \
  --agent-run "$BIN" \
  --session-id=test-open-v8 \
  --timeout=3m
```

**VERIFY (GREEN) criteria — all required:**

| Check | Signal |
|-------|--------|
| Open completes | bind / no fatal open error |
| First snapshot | contains `Paris` (or case-insensitive) |
| After `/exit` | status `exited: true` (human or JSON) |
| Resume `--open hello` | exit 0, **not** `already in use` |
| Resume snapshot | non-empty; contains `hello` and/or interactive UI; not dead-only footer |

**Exit code:** `0` + `RESULT: PASS` / `VERIFY: PASS …`

**FAIL:** non-zero + `RESULT: FAIL` + reason.

---

## Step 5 — Fix (if inspect fails)

| Failure | Likely fix area |
|---------|-----------------|
| Paris never appears | prompt inject, project picker, grok mock |
| exited stays false | `computeRunnerExited` / zombie signals |
| resume already in use | `prepareResumeTerminalID` / reclaim |
| empty snapshot after resume | open/resume inject, snapshot path |
| Killed: 9 | codesign ad-hoc re-sign |

Return to step 1 after code changes.

---

## Dry-run log

| When (UTC) | Step | Result | Evidence |
|------------|------|--------|----------|
| 2026-07-11T13:07:24Z | 1 Build | OK | `go build` + ad-hoc codesign |
| 2026-07-11T13:07:24Z | 2 Deploy | OK | binary valid on disk |
| 2026-07-11T13:07:57Z | 3+4 Inspect VERIFY | **VERIFY PASS** | Paris in snapshot; exited=true; resume --open exit 0; snapshot shows resumed interactive UI (`Loading session…`, prompt) |
| 2026-07-11T13:07:57Z | dry_run_status | **VERIFY PASS** | Full user flow automated |

---

## Cleanup

Inspect script kills serve PIDs under its temp `AGENT_RUN_HOME`. Manual:

```sh
# only for leftover global sessions if needed
agent-run status test-open-v8
```
