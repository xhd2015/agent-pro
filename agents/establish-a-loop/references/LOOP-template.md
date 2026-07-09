---
title: <title>
created: <YYYY-MM-DD>
slug: <slug>
path: <doc-or-docs>/LOOP_<YYYY-MM-DD>_<slug>.md
loop_kind: <bug-repro|health-check|regression>
dry_run_status: PENDING
---

# LOOP: <title>

## Symptom (bug-repro / regression only)

<User-pasted logs, status fields, or errors — verbatim. Example:>

```
[2026-07-08T23:52:54+08:00] starting service openclaw
[2026-07-08T23:52:54+08:00] failed to start: fork/exec /bin/bash: no such file or directory
```

**Repro precondition:** <state that must hold to see the symptom, e.g. WorkingDir
`/root/my-openclaw` does not exist. Document only — do not fix in steps 1–4.>

## Goal

- **bug-repro:** steps 1–4 reproduce Symptom on command (inspect RED).
- **health-check:** steps 1–4 confirm healthy end state (inspect GREEN).
- **regression:** steps 1–4 reproduce Symptom; after step 5, symptom absent (inspect GREEN).

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

**No remediation during dry-run steps 1–4** (bug-repro / regression): do not
satisfy repro preconditions until step 5.

## Loop steps

### 1. Build

```sh
<build commands>
```

**Verify:** `<command>` → <expected signal: exit 0, artifact exists, substring>

### 2. Deploy / Update

```sh
<deploy commands — use non-interactive flags or piped answers for Y/n prompts>
```

**Verify:** `<command>` → <expected signal>

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

### 4b. Inspect / Feedback — Verify (health-check, or regression post-fix)

```sh
<health check, sustained probe, log tail for absence of symptom>
```

**Verify:** symptom **absent** and healthy signal present — e.g.:

- `Status: running`, live PID
- log shows `<healthy marker>` and no Symptom substrings
- inspect script exits 0, prints `PASS:` or `VERIFY:`

Skip step 4b during initial bug-repro dry-run (REPRO PASS only).

### 5. Fix → return to step 1

When repro inspect (step 4) confirms the symptom:

- Read evidence (logs, stderr, exit codes).
- Hypothesize root cause.
- Apply fix (config, code, `mkdir`, install — **here only**).
- For `regression`: re-run steps 3 → 4b and record VERIFY PASS.
- For code changes: return to **1. Build**.

Optional aux scripts (only when markdown steps are insufficient):

```
script/debug/<purpose>-repro/main.go   # exits non-zero + REPRO: on unfixed system
script/debug/<purpose>-verify/main.go  # exits 0 + PASS: after fix (regression)
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

## Dry-run log

Update `dry_run_status` in the YAML header when the kind-appropriate gate is met.

| Step | Time | Result | Evidence |
|------|------|--------|----------|
| Prerequisites | | | |
| 1. Build | | | |
| 2. Deploy | | | |
| 3. Run | | | |
| 4. Inspect (repro) | | REPRO PASS / FAIL | <symptom excerpt or REPRO: output> |
| 4b. Inspect (verify) | | VERIFY PASS / SKIP | <only after step 5 for regression> |
| 5. Fix | | | <what changed, if applied during dry-run> |

**Dry-run status:** `PENDING` | `REPRO PASS` | `VERIFY PASS` | `BLOCKED` (must match YAML `dry_run_status`)

- `bug-repro`: complete at **REPRO PASS**
- `health-check`: complete at **VERIFY PASS** (step 4b or step 4 if no repro phase)
- `regression`: **REPRO PASS** required before **VERIFY PASS**