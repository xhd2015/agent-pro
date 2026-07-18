---
name: verify-on-behalf-of-user/tty
description: >-
  Interactive CLI / real-PTY verification with tty-watch: run --detach (non-blocking),
  send input, snapshot scrollback, always kill to reclaim. Avoid pipe-only and
  raw-openpty false hangs.
---

# TTY / interactive CLI verification

Use this topic when a claim needs a **real interactive terminal** (raw mode,
shell attach, “requires interactive terminal”, remote bash/terminal sessions).

Drive the CLI with **`tty-watch`**, not ad-hoc `openpty` Python or pipe-only
Attach helpers. `tty-watch` is agent-pro’s PTY session manager
(`script/tty-watch`, library `pkgs/ttywatch`).

## When TTY is required

- Product rejects non-TTY (`requires an interactive terminal on stdin/stdout`)
- Claim is interactive shell / session lifecycle under a PTY
- Pipe-based library tests alone would miss `MakeRaw` / real TTY behavior

If the claim is non-interactive CLI only, this topic does not apply (use
normal subprocess + stdout/stderr checks).

## Golden rules (non-negotiable)

### 1. Start with `run --detach` (non-blocking)

```sh
tty-watch run --detach --session-id <id> -- <command> [args...]
```

- Prints `session-id: …` and **returns immediately**.
- **Golden form** for agents and verify scripts.
- **Do not** use bare `tty-watch run …` without `--detach` in verify — that
  attaches as writer and **blocks** on the controlling TTY.
- Do not use `--headless` for this recipe (`--headless` and `--detach` are
  mutually exclusive).

### 2. End with `tty-watch kill` (reclaim resources)

```sh
tty-watch kill <session-id>
```

- **Required teardown**, even when the child already exited and snapshot shows
  an end marker (e.g. `[Terminal exited]`).
- Registry entries can **linger after process exit**; without `kill`, sessions
  leak under `TTY_WATCH_HOME` and confuse later `list` / locks.
- Prefer `trap 'tty-watch kill "$SID" 2>/dev/null || true' EXIT` so reclaim
  always runs.

### 3. Isolate the registry under sandbox HOME

```sh
export TTY_WATCH_HOME="$HOME/.verify-tty-watch"   # HOME is sandbox default-home
mkdir -p "$TTY_WATCH_HOME"
```

Never use a shared real-user tty-watch home for parallel verifies.

## Canonical recipe

```sh
source "${VERIFY_SKILL_ROOT}/scripts/enter-sandbox.sh"

# ensure tty-watch on PATH (build into $SANDBOX_BIN if needed)
# go build -o "$SANDBOX_BIN/tty-watch" ./script/tty-watch   # from agent-pro root

export TTY_WATCH_HOME="$HOME/.verify-tty-watch"
mkdir -p "$TTY_WATCH_HOME"
SID="verify-$$"

cleanup() { tty-watch kill "$SID" 2>/dev/null || true; }
trap cleanup EXIT

tty-watch run --detach --session-id "$SID" -- \
  "$SANDBOX_BIN/<cli>" <args that need a real TTY>

# optional: poll readiness
# tty-watch snapshot "$SID"

tty-watch send "$SID" $'exit\n'    # or claim-specific keystrokes

# poll end-of-session evidence via snapshot (not list emptiness alone)
# tty-watch snapshot "$SID"
# expect claim-specific success markers; ban unexpected error strings

# trap runs: tty-watch kill "$SID"
```

### End detection

- Prefer **`tty-watch snapshot <id>`** containing a session-end signal (e.g.
  `[Terminal exited]`) or claim-specific success text.
- Do **not** rely only on `tty-watch list` being empty — exited sessions may
  still appear until `kill`.
- Record snapshots and timings in the verify transcript.

### PASS / FAIL (interactive exit-style claims)

| Result | When |
|--------|------|
| **PASS** | Snapshot shows expected end within timeout; no banned error strings; `kill` ran |
| **FAIL** | Timeout without end marker; banned text (e.g. `unexpected EOF` when clean exit claimed); `tty-watch` missing |
| Control | Non-TTY path may still correctly fail with “requires interactive terminal” |

## Anti-patterns

```text
# WRONG — product may correctly reject; not interactive success
cli-that-needs-tty </dev/null

# WRONG — library/pipe path; may PASS while interactive CLI still broken
AttachWithIO(..., pipe, pipe) with SkipTTYCheck as sole evidence

# WRONG — blocks the verify agent
tty-watch run -- $SANDBOX_BIN/cli …

# WRONG — leaks registry / PTY resources
tty-watch run --detach … ; … ; # never kill

# WRONG — raw openpty + treat EIO as hang without waitpid
# (easy false FAIL; use tty-watch instead)
```

## Missing `tty-watch`

If the claim requires TTY and `tty-watch` is not on `PATH`:

- Verdict is **FAIL** (not skip), with a short install/build hint:
  - build from agent-pro: `go build -o "$SANDBOX_BIN/tty-watch" ./script/tty-watch`
  - or ensure the installed `tty-watch` binary is on `PATH`
- Same spirit as missing browser-agent for UI surfaces.

## Doctest vs this skill

| Layer | Role |
|-------|------|
| Doctests | May use pipes / `SkipTTYCheck` / harnesses for unit speed |
| **This skill** | Human-shaped: real PTY via **tty-watch** when the claim is interactive |

Doctest GREEN does not replace a TTY scenario when the user journey is interactive.

## Transcript notes

In Phase 4 scenarios, record:

- `TTY_WATCH_HOME`, session id
- `run --detach` command line
- key `send` payloads
- final `snapshot` snippet (truncate if huge)
- confirmation that `kill` ran (or trap cleanup)
