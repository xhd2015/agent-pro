---
name: verify-on-behalf-of-user/workflow
description: >-
  Verify workflow: scope and depth, git sanity, sandbox build, runtime bring-up,
  scenario execution (CLI and/or browser-agent), evidence and teardown, then
  write transcript and inline its full contents.
---

# Workflow

Run every phase **after** sourcing `enter-sandbox.sh` (see `sandbox` topic).
Read `scenario` for depth labels, surface gate, and browser-agent rules.

## Phase 0 — Scope, surface, depth

From conversation, requirement files, or user focus:

1. What was claimed done?
2. Which binaries/packages?
3. **Surface** — classify: CLI-only, server/HTTP, frontend/UI, session lifecycle, multi-binary
4. **Depth** — choose `smoke` \| `scenario` \| `full` and **always label it** with a one-line reason
5. Which scenarios (user journeys)?
6. Server or browser needed?

If unclear, ask once. Read a **project recipe** when it exists (e.g.
`docs/verify-recipes/<feature>.md` in the consumer repo).

**Surface → depth rules** (see `scenario` topic):

- Server, frontend/UI, or session lifecycle → depth ≥ **scenario** (runtime + real journey required)
- Frontend/UI → **browser-agent** required; if unusable → **FAIL**
- CLI-only may use **smoke** only when labeled and justified

## Phase 1 — Git sanity

Record in transcript:

```sh
$ pwd
$ git status --short
$ git diff --stat
```

For multi-repo work (e.g. `external/*`), run the same in each git root.

**Annotation:** note expected paths vs surprises.

## Phase 2 — Build & install (sandbox bin only)

```sh
source "${VERIFY_SKILL_ROOT}/scripts/enter-sandbox.sh"

$ go build -o "$SANDBOX_BIN/<binary>" <build-target>
# exit 0
```

Capture exit codes and relevant stdout/stderr in the transcript.

## Phase 3 — Runtime bring-up (required when surface needs it)

When server, web UI, or multi-process runtime is in scope:

1. Start processes under sandbox env (`HOME`, isolated data homes as needed)
2. **Wait for ready** (port listen / HTTP health) — do not only sleep
3. Record in transcript: PID(s), base URL, log path, ready-check command + result

Skip only for pure CLI-only surface at depth `smoke` (must still be labeled).

## Phase 4 — Scenario execution

Run **user journeys** from the recipe or from the claim — not only help/version.

Rules:

- All commands run **after** `enter-sandbox.sh`
- Use `TSK_HOME="$HOME/.tsk"`, `TSK_USER`, `TSK_DATE` as needed
- Assert exit code, stdout, stderr, HTTP, on-disk artifacts
- **UI:** use **browser-agent only** (`session new` → open app URL → eval/screenshot).
  Do **not** use `playwright-debug` here.
- **Interactive TTY:** if the claim needs a real terminal, use **tty-watch** only
  (see topic `tty`). Golden non-blocking start:
  `tty-watch run --detach --session-id <id> -- <cmd>…`. Drive with `send`,
  assert with `snapshot`. **Always** `tty-watch kill <id>` to reclaim (trap EXIT).
  Set `TTY_WATCH_HOME` under sandbox `HOME`. Do not use pipe-only or raw openpty
  as sole interactive evidence.
- Prefer a project recipe’s scenario steps when present
- Truncate huge output: `... (N lines omitted)`
- Label each scenario (S1, S2, …) with expected observables and ✓/✗

**Smoke depth:** still run the minimal real command that backs the claim; never
leave depth unlabeled.

## Phase 5 — Evidence & teardown

Record:

- Screenshot paths (UI), log tails, status dumps
- Teardown: kill leftover PIDs, note sandbox data left for inspection
- **tty-watch:** every session started with `run --detach` must get
  `tty-watch kill <session-id>` (reclaim registry/PTY). Prefer EXIT trap.

If UI was required and browser-agent/Chrome/session failed → verdict **FAIL**.
If TTY was required and `tty-watch` was missing → verdict **FAIL**.

## Phase 6 — Optional doctest spot-check

If cheap, re-run a small subset and record in transcript. Do not substitute for Phase 3–4.

## Phase 7 — Write transcript and inline full content

1. Write `~/.sandbox/transcripts/<ISO8601>-<slug>.md` (see `transcript` topic)
2. **Print the full file body** in the agent reply (same content as on disk)
3. Lead with path + verdict + labeled depth

**On FAIL:** stop at first hard failure, still save partial transcript and **inline it**.
