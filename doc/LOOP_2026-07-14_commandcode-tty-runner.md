---
title: "agent-run: add commandcode-tty as a TTY runner"
created: 2026-07-14
slug: commandcode-tty-runner
path: doc/LOOP_2026-07-14_commandcode-tty-runner.md
loop_kind: bug-repro
dry_run_status: "REPRO PASS"
---

# LOOP: commandcode-tty-runner

## Kind

**bug-repro** — `agent-run run --agent-runner commandcode-tty` fails with
"unknown agent runner" because `commandcode-tty` is not registered anywhere.

## Symptom

```
$ agent-run run --agent-runner commandcode-tty --agent-runner-binary /tmp/llm-mock-run-commandcode "Hello"
agent-run: unknown agent runner: commandcode-tty
```

**Root cause:** No provider is registered for `commandcode-tty` in:
- `pkgs/agenttty/provider.go` (TTY registry `init()`)
- `cmd/cmd/agent-run/runner_validate.go` (CLI validation switch)

**Repro preconditions:**
1. `agent-run` binary built from `./cmd/agent-run` (requires `go run ./script/agent-run/build-frontend` first).
2. `llm-mock-run-commandcode` binary built from `./agent/llm/llm-mock/llm-mock-run-commandcode`.
3. `cmd` CLI installed and on PATH (`which cmd`).

## Goal

Steps 1–4 reliably reproduce the "unknown agent runner" failure.
Step 5 fix implements the `commandcode-tty` TTY runner registration so `agent-run run --agent-runner commandcode-tty`:
1. Starts the TTY (PTY + `__serve__`)
2. Runs `llm-mock-run-commandcode` inside the PTY
3. The mock server responds to `/alpha/generate` NDJSON
4. `cmd` parses the response and outputs to scrollback
5. agent-run reads the scrollback and returns text

## Steps

### 1. Build

Build both binaries:

```sh
# Build frontend dist (required for agent-run)
go run ./script/agent-run/build-frontend

# Build agent-run
go build -o /tmp/agent-run ./cmd/agent-run

# Build llm-mock-run-commandcode
go build -o /tmp/llm-mock-run-commandcode ./agent/llm/llm-mock/llm-mock-run-commandcode
```

**Verify:** binaries exist

```sh
test -x /tmp/agent-run && echo "agent-run: OK"
test -x /tmp/llm-mock-run-commandcode && echo "llm-mock-run-commandcode: OK"
```

### 2. Deploy / Update

No deploy. Local binaries in `/tmp/`.

### 3. Run

Attempt to use `commandcode-tty` as a runner:

```sh
/tmp/agent-run run --agent-runner commandcode-tty --agent-runner-binary /tmp/llm-mock-run-commandcode "Hello" 2>&1
echo "EXIT: $?"
```

**Expected (symptom):** Non-zero exit. Stderr contains `unknown agent runner: commandcode-tty`.

### 4. Inspect

```sh
echo "EXIT_CODE=$RC"
echo "STDERR: $STDERR_OUTPUT"
echo "$STDERR_OUTPUT" | grep -q "unknown agent runner: commandcode-tty" && echo "REPRO: symptom confirmed" || echo "CHECK: symptom not matched"
```

**Expected (REPRO):** Exit code 1. Stderr: `agent-run: unknown agent runner: commandcode-tty`.

### 5. Fix

**5a. Register TTY provider** (`pkgs/agenttty/provider.go`):

Added `commandcode-tty` registration in `init()` with `BannerProvider: "commandcode"`, `DisableTail: true`, and hook functions.

**5b. New file: `pkgs/agenttty/commandcode.go`**:

- `BuildCommandcodeCommandArgv` — resolves `--agent-runner-binary`, supports `AGENT_RUN_COMMANDCODE_TTY_COMMAND` env override
- `detectCommandcodeScreenStatus` — returns "idle" when scrollback non-empty, "starting" otherwise
- `checkCommandcodeWritable` — always returns `{Ready: true}`
- `waitContentAppears` — polls scrollback until content appears (used before serve exits)

**5c. Banner detection** (`pkgs/agenttty/banner.go`):

Added `provider == "commandcode"` case: `return strings.TrimSpace(plain) != ""` (any content = banner ready).

**5d. Prompt argv handling** (`pkgs/agenttty/run.go`):

- Headless (non-open): injects `-p` flag so `cmd` runs in print mode and exits cleanly after responding
- `--open`: omits `-p` so `cmd` stays interactive and the PTY survives for attach
- Pre-capture snapshot via `waitContentAppears` before `WaitHeadless`
- Added `commandcode-tty` to scrollback-fallback emit condition (alongside `codex-tty`)

## Pitfalls & blockers

| Pitfall | Mitigation |
|---------|------------|
| Frontend dist missing for agent-run build | Run `go run ./script/agent-run/build-frontend` first |
| `cmd` requires `-p` flag for non-interactive mode | `BuildCommandcodeCommandArgv` injects `-p`; prompt is positional arg appended by run.go |
| Scrollback snapshot lost after serve exit | `waitContentAppears` polls for content before `WaitHeadless`; pre-captured snapshot used |
| Banner detection needs any non-empty scrollback | `bannerDetectedConfig` for `commandcode` provider: `plain != ""` |
| `/alpha/generate` response format mismatch | Mock server already streams NDJSON since LOOP_2026-07-14_llm-mock-alpha-generate |
| Command Code mock responses are verbose mock-gen output | Acceptable for E2E; real Command Code would return proper LLM text |

## Dry-run log

| Step | Timestamp | Result | Evidence |
|------|-----------|--------|----------|
| 1 Build | 10:30 | OK | `/tmp/agent-run` and `/tmp/llm-mock-run-commandcode` built |
| 2 Deploy | 10:30 | OK | No deploy needed; local binaries in `/tmp/` |
| 3 Run | 10:30 | REPRO | `RC=1`, stderr: `agent-run: unknown agent runner: commandcode-tty` |
| 4 Inspect | 10:30 | **REPRO PASS** | `grep -q "unknown agent runner: commandcode-tty"` matched |
| 5 Fix | 10:45 | Applied | Registered provider in `pkgs/agenttty/provider.go`, implemented `commandcode.go` (BuildArgv, banner, writable, waitContentAppears), added scrollback capture in `run.go`, added banner detection for `commandcode` in `banner.go` |
| 3b Run (post-fix) | 10:49 | VERIFY | Exit 0, JSON output: `{"type":"message","role":"assistant","text":"First, let me check what's needed for Hello..."}` |
| 4b Inspect (post-fix) | 10:49 | **VERIFY PASS** | `cmd -p` inside PTY exits cleanly; scrollback captured and emitted as ActionMessage event |
