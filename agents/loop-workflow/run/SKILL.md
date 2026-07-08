---
name: loop-workflow
description: >-
  Goal-driven achievement loop. Set a goal, clarify requirements, write and
  verify an inspect script under script/debug/, then iterate with max effort
  until the goal is correctly achieved. Use when the user runs /loop-workflow, says
  "loop until done", "keep going until", "max effort", "don't stop until",
  or wants a persistent fix/implement cycle with automated verification.
---

You are a goal-driven loop operator. You do not stop at "probably fixed" — you
keep iterating until an inspect script confirms the goal is **correctly achieved**.

# Modes

| Phase | Name | Stop condition |
|-------|------|----------------|
| 0 | Goal intake | Goal recorded via `update_goal` |
| 1 | Clarification | Only blocking unknowns resolved (≤5 simple questions) |
| 2 | Inspect script | Script written, run, and verified agent-readable |
| 3 | Achievement loop | `update_goal(completed=true)` after inspect PASS |

**Do not enter Phase 3 until Phase 2 passes.**  
**Do not call `update_goal(completed=true)` until the inspect script exits 0
and you have independently confirmed the output matches the goal.**

# Phase 0 — Goal intake

1. Restate the goal in one sentence: what "done" looks like, not how to get there.
2. Call `update_goal` with a short progress message describing the goal.
3. Identify the **verification surface** — what observable signal proves success:
   - Web UI state (screenshot, DOM text, network response)
   - CLI stdout/stderr or exit code
   - Test suite (`go test`, `doctest test`, project scripts)
   - Remote GitHub Actions workflow (`gh run view`)
   - File artifact (exists, content, diff)
   - Service health (HTTP status, log line)

If the goal is vague, proceed to Phase 1. If it is already precise, skip to Phase 2
planning after one confirmation question.

# Phase 1 — Simple clarification

Ask **at most 5** short, concrete questions. One message, numbered list.
Do **not** enter full brainstorm mode — only resolve blockers.

Good clarification topics:
- Exact success criteria ("what should I see when it works?")
- Scope boundary ("fix only X or also Y?")
- Environment ("local only, or must CI pass too?")
- Constraints ("no API changes?", "must stay backward compatible?")
- Preferred verification ("screenshot OK?", "which workflow name?")

**Stop asking when you can write an inspect script.** Do not wait for "go ahead"
unless the user explicitly invoked `brainstorm` or `followup` in the same thread.

Record answers; fold them into the goal statement and `update_goal` message.

# Phase 2 — Inspect script

## Location

Create the script under one of:

```
script/debug/<name>/main.go
script/debug/<topic>/<sub-dir>/main.go
```

Choose `<name>` or `<topic>/<sub-dir>` from the goal (kebab-case, e.g.
`fix-login-redirect`, `ci/go-debug-linux`).

Add a sibling `README.md` only when the inspect steps are non-obvious (env vars,
tokens, one-time setup). Keep it minimal.

## Contract

The inspect script is the **single source of truth** for "goal achieved."

| Requirement | Detail |
|-------------|--------|
| Runnable | `go run ./script/debug/<path>` from repo root |
| Exit code | `0` = goal met; non-zero = not met |
| Output | Print `PASS:` or `FAIL:` as the first line of the result block |
| Evidence | Write artifacts under `script/debug/<path>/out/` (screenshots, logs, JSON) |
| Agent-readable | Stdout must contain enough detail to diagnose failure without re-running |
| Idempotent | Safe to run repeatedly during the loop |

## Inspect capabilities (use what fits)

Delegate to existing tools — do not reinvent them:

| Surface | Approach |
|---------|----------|
| Web UI | `playwright-debug` skill — screenshot to `out/screenshot.png`, assert DOM text |
| GitHub workflow | `gh run list` + `gh run view --log`; check conclusion and log patterns |
| Local CI replay | `go run ./script/run-github-workflow-via-act/...` when applicable |
| Unit/integration tests | `go test`, `doctest test`, or project `script/run-test` |
| CLI behavior | `os/exec` the real binary; compare stdout to expected template |
| Files | `os.Stat`, diff against golden, grep for required strings |

See `references/inspect-script-template.go` for a starter skeleton.

## Write → verify → fix script

1. **Write** `main.go` implementing the success criteria from Phase 0/1.
2. **Run** `go run ./script/debug/<path>` yourself.
3. **Expect FAIL initially** — the goal is not yet achieved. Confirm:
   - Script compiles and runs
   - Failure output is actionable (not a panic or missing dep)
   - Artifacts land in `out/` when applicable
4. If the script itself is broken, fix it before Phase 3. Log progress via
   `update_goal(message="inspect script verified (RED)")`.

**Phase 2 gate:** inspect script runs cleanly and produces agent-readable FAIL
output. Only then enter Phase 3.

# Phase 3 — Achievement loop

```
┌─────────────────────────────────────────────┐
│  IMPLEMENT / FIX (production code, tests)   │
└──────────────────┬──────────────────────────┘
                   ▼
┌─────────────────────────────────────────────┐
│  RUN INSPECT SCRIPT                         │
│  go run ./script/debug/<path>               │
└──────────────────┬──────────────────────────┘
                   ▼
          exit 0? ──no──► analyze FAIL output
              │              │
              │              └──► hypothesize → implement → loop
              ▼
    confirm PASS output matches goal
              │
              ▼
    update_goal(completed=true, message=summary)
```

## Loop rules (max effort, no premature stop)

1. **No stop on first attempt** — expect multiple iterations.
2. **Evidence every iteration** — run the inspect script after every meaningful
   change; paste the `PASS:`/`FAIL:` line in your progress notes.
3. **Root cause before workaround** — trace failures; prefer sound fixes (see
   `sound-fix` skill when unsure).
4. **Broaden only when stuck** — after 3 failed iterations on the same hypothesis,
   reconsider assumptions, re-read code, or add debug logging.
5. **Remote verification** — when the goal includes CI, push/trigger workflow and
   extend the inspect script to poll `gh run view` (document required `gh auth`
   in README if needed).
6. **Regressions** — run affected existing tests each iteration when touching
   production code.
7. **Progress** — call `update_goal(message=...)` each iteration with iteration
   number and current FAIL reason or partial win.

## When you may stop (only these)

| Situation | Action |
|-----------|--------|
| Inspect script exits 0 and output matches goal | `update_goal(completed=true)` — **success** |
| Inspect script prints PASS but evidence contradicts goal | Keep looping; fix script or implementation |
| Hard external blocker (missing creds, user decision) | `update_goal(blocked_reason=...)` — ask user **once**, then resume when unblocked |
| Same error 5+ times with no new information | `update_goal(blocked_reason=...)` — present findings and ask one targeted question |

**Never** stop because something "should work", "looks correct", or "tests pass
but inspect wasn't re-run."

## After success

1. Summarize: goal, iterations, key fix, inspect script path.
2. Paste final inspect script PASS output (abbreviated if long).
3. List artifact paths under `script/debug/<path>/out/`.
4. Note any inspect script or README the user should keep for regression checks.

# Inspect script conventions

- Package `main`; follow patterns in `script/` siblings (`handle` + subcommands if
  multiple checks).
- Use `github.com/xhd2015/xgo/support/cmd` for running shell commands when in
  the xgo repo; otherwise `os/exec`.
- Print structured sections:

```
CHECK: <what is being verified>
EVIDENCE: <path or excerpt>
RESULT: PASS|FAIL
REASON: <one line if FAIL>
```

- Add `out/` to `.gitignore` at the script dir level if screenshots/logs should
  not be committed:

```
# script/debug/<path>/.gitignore
out/
```

# Anti-patterns

- Skipping the inspect script and relying on manual eyeballing
- Writing the inspect script after claiming success
- Stopping at `go test` green when the goal requires UI or CI verification
- Calling `update_goal(completed=true)` without a fresh inspect PASS
- Ask-wait-ask clarification loops — batch questions once in Phase 1
- Deleting the inspect script after success (keep it for regression)

# Quick start (user-facing)

```
/loop-workflow <goal description>
```

Example: `/loop-workflow fix the login redirect so /dashboard loads after sign-in; verify
with a screenshot of the dashboard heading.`

The agent will clarify, write `script/debug/fix-login-redirect/main.go`, verify
it fails (RED), then loop until inspect PASS.