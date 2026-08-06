---
name: verify-on-behalf-of-user/workflow
description: >-
  Verify workflow: scope and depth, git sanity, mode-branched build (sandbox bin
  or host install ladder), runtime bring-up, scenario execution (CLI and/or
  browser-agent), evidence and teardown, then write transcript and inline its
  full contents.
---

# Workflow

Choose **Mode** first (`sandbox` default; `host` only on explicit opt-in — see
topics `sandbox` and `host`). In sandbox mode, run phases after sourcing
`enter-sandbox.sh`. In host mode, do **not** source `enter-sandbox.sh`.

Read `scenario` for depth labels, surface gate, and browser-agent rules.

## Phase 0 — Scope, surface, depth, mode

From conversation, requirement files, or user focus:

1. What was claimed done?
2. Which binaries/packages? (**change-scoped targets** — names relevant to the change)
3. **Mode** — `sandbox` (default) or `host` only if user explicitly opted in
   ("host", "outside sandbox", or clear synonyms). If ambiguous → **sandbox**.
4. **Surface** — classify: CLI-only, server/HTTP, frontend/UI, session lifecycle, multi-binary
5. **Depth** — choose `smoke` \| `scenario` \| `full` and **always label it** with a one-line reason
6. Which scenarios (user journeys)?
7. Server or browser needed?

If unclear, ask once. Read a **project recipe** when it exists (e.g.
`docs/verify-recipes/<feature>.md` in the consumer repo).

**Surface → depth rules** (see `scenario` topic):

- Server, frontend/UI, or session lifecycle → depth ≥ **scenario** (runtime + real journey required)
- Frontend/UI → **browser-agent** required; if unusable → **FAIL**
- CLI-only may use **smoke** only when labeled and justified

**Host mode:** emit mandatory warnings before mutating steps (see `host`).

## Phase 1 — Git sanity

Record in transcript:

```sh
$ pwd
$ git status --short
$ git diff --stat
```

For multi-repo work (e.g. `external/*`), run the same in each git root.

**Annotation:** note expected paths vs surprises; list change-scoped target names.

## Phase 2 — Build & install (mode-branched)

### Mode: sandbox (default)

```sh
source "${VERIFY_SKILL_ROOT}/scripts/enter-sandbox.sh"

$ go build -o "$SANDBOX_BIN/<binary>" <build-target>
# exit 0
```

Build only binaries relevant to the claim when practical. Capture exit codes and
relevant stdout/stderr in the transcript.

### Mode: host (opt-in only)

Do **not** source `enter-sandbox.sh`. Follow topic **`host`**:

1. Resolve **change-scoped targets** (e.g. `my-tool` → `./cmd/my-tool`,
   `./script/my-tool/install`, root `./script/install` only if relevant to that tool)
2. Install ladder: `wrk --reinstall-local` (dry-run then apply) → scoped
   `script`/`scripts` installers → `go install ./cmd/<T>` → FAIL if none
3. Prefer `--dry-run` or a written plan before any mutating op; show `warning:` lines
4. Record Install method, Targets, Checked, Chosen, Skipped

## Phase 3 — Runtime bring-up (required when surface needs it)

When server, web UI, or multi-process runtime is in scope:

1. Start processes under the mode's env (sandbox: isolated `HOME`; host: real HOME —
   still prefer ephemeral ports / temp data when possible)
2. **Wait for ready** (port listen / HTTP health) — do not only sleep
3. Record in transcript: PID(s), base URL, log path, ready-check command + result

Skip only for pure CLI-only surface at depth `smoke` (must still be labeled).

## Phase 4 — Scenario execution

Run **user journeys** from the recipe or from the claim — not only help/version.

Rules:

- **Sandbox:** all commands run **after** `enter-sandbox.sh`
- **Host:** use real env; do not claim sandbox isolation
- Use `TSK_HOME`, `TSK_USER`, `TSK_DATE` as needed (under sandbox HOME when Mode is sandbox)
- Assert exit code, stdout, stderr, HTTP, on-disk artifacts
- **UI:** use **browser-agent only** (`session new` → open app URL → eval/screenshot).
  Do **not** use `playwright-debug` here.
- **Interactive TTY:** if the claim needs a real terminal, use **tty-watch** only
  (see topic `tty`). Golden non-blocking start:
  `tty-watch run --detach --session-id <id> -- <cmd>…`. Drive with `send`,
  assert with `snapshot`. **Always** `tty-watch kill <id>` to reclaim (trap EXIT).
  Set `TTY_WATCH_HOME` under sandbox `HOME` when Mode is sandbox.
- Prefer a project recipe’s scenario steps when present
- Truncate huge output: `... (N lines omitted)`
- Label each scenario (S1, S2, …) with expected observables and ✓/✗

**Smoke depth:** still run the minimal real command that backs the claim; never
leave depth unlabeled.

## Phase 5 — Evidence & teardown

Record:

- Screenshot paths (UI), log tails, status dumps
- Teardown: kill leftover PIDs; sandbox mode: note sandbox data left for inspection
- **tty-watch:** every session started with `run --detach` must get
  `tty-watch kill <session-id>` (reclaim registry/PTY). Prefer EXIT trap.

If UI was required and browser-agent/Chrome/session failed → verdict **FAIL**.
If TTY was required and `tty-watch` was missing → verdict **FAIL**.

## Phase 6 — Optional doctest spot-check

If cheap, re-run a small subset and record in transcript. Do not substitute for Phase 3–4.

## Phase 7 — Write transcript and inline full content

1. Write `~/.sandbox/transcripts/<ISO8601>-<slug>.md` (see `transcript` topic)
2. Header must include **Mode** (`sandbox` \| `host`) and labeled depth
3. **Print the full file body** in the agent reply (same content as on disk)
4. Lead with path + verdict + labeled depth + mode

**On FAIL:** stop at first hard failure, still save partial transcript and **inline it**.
