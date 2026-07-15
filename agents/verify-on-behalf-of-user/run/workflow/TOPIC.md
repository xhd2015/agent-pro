---
name: verify-on-behalf-of-user/workflow
description: >-
  Six-phase verify workflow: scope, git sanity, sandbox build, smoke commands,
  browser/logs, optional doctest spot-check, then transcript.
---

# Workflow

Run every phase **after** sourcing `enter-sandbox.sh` (see `sandbox` topic).

## Phase 0 — Scope

From conversation, requirement files, or user focus:

1. What was claimed done?
2. Which binaries/packages?
3. Which smoke commands?
4. Server or browser needed?

If unclear, ask once. Read a **project recipe** when it exists (e.g.
`docs/verify-recipes/<feature>.md` in the consumer repo).

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

## Phase 3 — Smoke commands

Run commands from the project recipe. Rules:

- All commands run **after** `enter-sandbox.sh`
- Use `TSK_HOME="$HOME/.tsk"`, `TSK_USER`, `TSK_DATE` as needed
- Assert exit code, stdout, stderr, on-disk artifacts
- Truncate huge output: `... (N lines omitted)`

## Phase 4 — Browser / logs (if applicable)

For UI or server work, record in transcript:

- URL opened (e.g. `http://localhost:8008/...`)
- Tool used (browser-agent, DevTools, `tail` on server log)
- Screenshot path if captured

Skip when the change is CLI-only.

## Phase 5 — Optional doctest spot-check

If cheap, re-run a small subset and record in transcript. Do not substitute for Phase 3.

## Phase 6 — Write transcript

See the `transcript` topic for format rules and the installed template at
`templates/transcript.md`.

**On FAIL:** stop at first failure, still save partial transcript.