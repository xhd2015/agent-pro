# LOOP: <title>

Created: <YYYY-MM-DD>
Slug: <slug>
Path: <doc-or-docs>/LOOP_<YYYY-MM-DD>_<slug>.md

## Goal

<One sentence — observable success criterion. Example: remote-agent ping succeeds
continuously for 10 minutes and the server does not restart.>

## Prerequisites (agent-auditable)

Run each check before entering the loop. Mark ✅ pass, ❌ fail, or BLOCKER.

| Check | Command | Pass criteria |
|-------|---------|---------------|
| <tool> | `which <tool>` | exit 0 |
| <auth> | `<non-interactive auth probe>` | exit 0, expected output |

**BLOCKER policy:** If any prerequisite is BLOCKER (interactive auth, missing
software the agent cannot install, required secret not in env), stop and document
the exact unblock step in **Pitfalls**. Do not prompt the user mid-loop.

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

```sh
<start service / execute target>
```

**Verify:** `<command>` → <expected signal>

### 4. Inspect / Feedback

```sh
<health check, log tail, sustained probe>
```

**Verify:** <quantitative success — e.g. N consecutive successes over T minutes>

### 5. Fix → return to step 1

When inspect fails:

- Read evidence (logs, stderr, exit codes).
- Hypothesize root cause.
- Apply fix to production code or config.
- Return to **1. Build**.

Optional aux script (only when markdown steps are insufficient):

```
script/debug/<purpose>/main.go
```

## Pitfalls & blockers

- Interactive prompts — bypass with flags (`-y`, `--yes`) or piped input
  (`echo -ne $'n\nn\n' | cmd`) when no flag exists.
- Missing binary — try project install path or `smc` plugin; else BLOCKER.
- Flaky network / rate limits — add retries and timeouts in commands.

## Dry-run log

| Step | Time | Result | Evidence |
|------|------|--------|----------|
| Prerequisites | | | |
| 1. Build | | | |
| 2. Deploy | | | |
| 3. Run | | | |
| 4. Inspect | | | |

**Dry-run status:** PENDING | PASS | BLOCKED