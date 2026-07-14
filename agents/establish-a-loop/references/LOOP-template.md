---
title: <title>
created: <YYYY-MM-DD>
slug: <slug>
path: <doc-or-docs>/LOOP_<YYYY-MM-DD>_<slug>.md
loop_kind: <bug-repro|health-check|regression>
establishment_status: PENDING
---

# LOOP: <title>

## User input (verbatim)

<Pasted terminal logs and/or prose step descriptions — unchanged from the user.>

## Derived operations

| # | Intent | Tool class | Derived command / URL | LOOP step | Notes |
|---|--------|------------|----------------------|-----------|-------|
| 1 | <e.g. check status> | workspace | `$LOOP_AGENT_RUN status <id>` | 4 Inspect | |
| 2 | <e.g. trigger flow> | workspace | `$LOOP_AGENT_RUN run --open ...` | 3 Run | |
| 3 | <e.g. API health> | external | `curl -sf http://127.0.0.1:<port>/api/...` | 4 Inspect | |
| 4 | <e.g. UI check> | external | `playwright-debug` → `http://127.0.0.1:<port>/<path>` | 4 Inspect | exact URL |
| 5 | <e.g. orchestration> | aux | `go run ./script/debug/<slug>/` | 3 Run / 4 Inspect | |

**Tool class:** `workspace` (repo `./cmd/...`), `aux` (`script/debug/...`), `external` (Prerequisites only).

Command blocks in steps 1–4 must match this table. No prose substitutes ("start the server").

## Workspace tools

<Include when any derived op uses a repo CLI. Omit section when external-only.>

| Binary | Source | Build | Deploy env / path |
|--------|--------|-------|-------------------|
| <binary> | `./cmd/<binary>` | `go build -o $BIN/<binary> ./cmd/<binary>` | `LOOP_<TOOL>="$BIN/<binary>"` |

## Symptom (bug-repro / regression only)

<User-pasted logs, status fields, or errors — verbatim. Example:>

```
[2026-07-08T23:52:54+08:00] starting service openclaw
[2026-07-08T23:52:54+08:00] failed to start: fork/exec /bin/bash: no such file or directory
```

**Repro precondition:** <state that must hold to see the symptom, e.g. WorkingDir
`/root/my-openclaw` does not exist. Document only — do not adjust in steps 1–4.>

## Observation notes (optional — when Phase 1 observe ran)

<State transitions, timings, unexpected values from exploratory observe run.
Evidence paths under `script/debug/<slug>/out/observation/`.>

## Goal

- **bug-repro:** steps 1–4 reproduce Symptom on command (inspect RED).
- **health-check:** steps 1–4 confirm healthy end state (inspect GREEN).
- **regression:** steps 1–4 reproduce Symptom; after step 5 adjust, symptom absent (inspect GREEN).

<One sentence summary.>

## Prerequisites (agent-auditable)

Run each check before entering the loop. Mark ✅ pass, ❌ fail, or BLOCKER.

| Check | Command | Pass criteria |
|-------|---------|---------------|
| <tool> | `which <tool>` | exit 0 |
| <auth> | `<non-interactive auth probe>` | exit 0, expected output |

**BLOCKER policy:** If any prerequisite is BLOCKER (interactive auth, missing
software the agent cannot install, required secret not in env), stop and document
the exact unblock step in **Pitfalls**. Do not prompt the user mid-loop.

**No remediation during establishment-run steps 1–4** (bug-repro / regression):
do not satisfy repro preconditions until step 5.

## Loop steps

### 1. Build

Build every binary listed in **Workspace tools**. Add frontend/dist stubs when a
cmd embeds UI and dist is missing.

```sh
BIN="${TMPDIR:-/tmp}/loop-<slug>-bin"
mkdir -p "$BIN"
# go build -o "$BIN/<binary>" ./cmd/<binary>  # one line per Workspace tools row
```

**Verify:** `test -x "$BIN/<binary>"` for each workspace tool → exit 0

### 2. Deploy / Update

Export workspace binary paths. Use non-interactive flags or piped answers for Y/n prompts.

```sh
# export LOOP_<TOOL>="$BIN/<binary>"  # one line per Workspace tools row
# remote upload / restart as needed
```

**Verify:** `"$LOOP_<TOOL>" --help` or equivalent → expected subcommand listed

### 3. Run

Trigger the action under test. **No remedial setup here** for bug-repro/regression
(no mkdir, config edits, installs — those belong in step 5 only).

```sh
<restart service / execute target — failure trigger only>
```

**Verify:** `<command>` → <expected signal: process started or restart accepted>

### 4. Inspect / Feedback — Repro (bug-repro / regression)

```sh
<tail logs, service list, status check — non-streaming preferred>
```

**Verify:** symptom **present** — e.g.:

- log tail contains `<exact substring from Symptom>`
- `Status: error` and `Last Error: <expected>`
- `go run ./script/debug/<repro-script>/` exits non-zero, prints `REPRO:` + evidence

### 4b. Inspect / Feedback — Verify (health-check, or regression post-adjust)

```sh
<health check, sustained probe, log tail for absence of symptom>
```

**Verify:** symptom **absent** and healthy signal present — e.g.:

- `Status: running`, live PID
- log shows `<healthy marker>` and no Symptom substrings
- inspect script exits 0, prints `PASS:` or `VERIFY:`

Skip step 4b during initial bug-repro establishment (SYMPTOM CONFIRMED only).

### 5. Adjust → return to step 1

When inspect (step 4) is unclear or confirms the symptom:

- Read evidence (logs, stderr, exit codes).
- Hypothesize root cause.
- Apply one or more sub-actions (**here only**):
  - **Instrument** — add or refine debug logs (`LOOP_DEBUG:`); document paths and grep patterns
  - **Code** — change application logic
  - **Environment** — config, `mkdir`, install
- Remove or gate temporary debug logs after diagnosis (before final code/env adjust).
- For `regression`: re-run steps 3 → 4b and record VERIFY PASS.
- After instrument or code changes: return to **1. Build**.

Optional aux scripts (only when markdown steps are insufficient):

```
script/debug/<purpose>-repro/main.go   # exits non-zero + REPRO: on unfixed system
script/debug/<purpose>-verify/main.go  # exits 0 + PASS: after adjust (regression)
```

## Pitfalls & blockers

- Interactive prompts — bypass with flags (`-y`, `--yes`) or piped input
  (`echo -ne $'n\nn\n' | cmd`) when no flag exists.
- Missing binary — try project install path or `smc` plugin; else BLOCKER.
- Flaky network / rate limits — add retries and timeouts in commands.
- Streaming log commands hang — prefer `remote-agent exec sh -c 'tail -n N …'`
  over `service logs` for non-interactive inspect.
- Misleading errors — document actual root cause in step 5, not only the surface
  symptom string.
- Temporary debug logs — document `LOOP_DEBUG:` locations; remove or gate in step 5
  Adjust after diagnosis.

## Run log

Update `establishment_status` in the YAML header when the kind-appropriate gate is met.

| Step | Time | Result | Evidence |
|------|------|--------|----------|
| Plan (expand / observe) | | | <derived ops; observation/ paths if used> |
| Prerequisites | | | |
| Instrument retry | | | <LOOP_DEBUG: files if Phase 3 repeated> |
| 1. Build | | | |
| 2. Deploy | | | |
| 3. Run | | | |
| 4. Inspect (repro) | | SYMPTOM CONFIRMED / FAIL | <symptom excerpt or REPRO: output> |
| 4b. Inspect (verify) | | VERIFY PASS / SKIP | <only after step 5 for regression> |
| 5. Adjust | | | <sub-action: instrument / code / env; what changed> |

**Establishment status:** `PENDING` | `SYMPTOM CONFIRMED` | `VERIFY PASS` | `BLOCKED`
(must match YAML `establishment_status`)

- `bug-repro`: complete at **SYMPTOM CONFIRMED** (bug reproduced — not a "pass")
- `health-check`: complete at **VERIFY PASS** (step 4b or step 4 if no repro phase)
- `regression`: **SYMPTOM CONFIRMED** required before **VERIFY PASS**

## Expand to operations (reference)

Phase 1 always produces **User input**, **Derived operations**, and **Workspace tools**
(when applicable). Rules:

- CLI: full command + flags; workspace binaries via `$LOOP_<TOOL>` after Deploy.
- API: `curl -sf` + full URL; UI: `playwright-debug` + full URL (scheme, host, port, path).
- Workspace `./cmd/*` → Build + Deploy; never assume `which` alone.
- Symptom section = verbatim failure evidence only; commands live in Derived operations.
- Frontend embed stubs in Build when dist is missing (see existing LOOP docs in `doc/`).