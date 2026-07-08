---
name: establish-a-loop
description: >-
  Establish an agent-runnable verification loop as LOOP_<YYYY-MM-DD>_<slug>.md
  with build, deploy/update, run, inspect, and fix steps. Audits prerequisites,
  writes the loop doc, dry-runs the full cycle without user input, and polishes
  until reliable. Use when the user runs /establish-a-loop, says "establish a
  loop", "create a debug loop", or needs a self-verifiable runbook before
  iterating on a fix. Pairs with loop-workflow for goal-driven execution.
---

You establish a **self-verifiable loop** the agent can run end-to-end without
user input. Deliverable: `LOOP_<YYYY-MM-DD>_<slug>.md`. Pair with `loop-workflow`
after the loop is verified.

# When to use

- User needs a repeatable build → deploy → run → inspect → fix cycle documented
- Verification spans multiple commands (local build, remote deploy, health probe)
- Before `/loop-workflow` — establish *how* to verify, then iterate on fixes

**Do not use** when the task is only "fix until done" and a loop doc already
exists — use `loop-workflow` directly.

# Output path

Filename: `LOOP_<YYYY-MM-DD>_<slug>.md` where date is creation day and slug is
kebab-case from the request.

Resolve directory at repo root (do not ask the user):

```
1. if doc/ exists     → doc/LOOP_<date>_<slug>.md
2. else if docs/ exists → docs/LOOP_<date>_<slug>.md
3. else               → docs/loops/LOOP_<date>_<slug>.md  (mkdir -p)
```

If both `doc/` and `docs/` exist, prefer `doc/`.

Copy structure from `references/LOOP-template.md` (bundled with this skill under
`agents/establish-a-loop/references/` when in repo).

# Aux scripts (optional)

The LOOP doc is **markdown only**. When a step needs code (sustained probe,
parsing, multi-check orchestration), add:

```
script/debug/<purpose>/main.go
```

Reference the script path from the LOOP doc. Do not embed Go in the LOOP file.
Follow `loop-workflow` inspect conventions if the aux script gates success
(`CHECK` / `RESULT` / `REASON`, exit 0 = pass).

# Phases

| Phase | Name | Gate |
|-------|------|------|
| 0 | Intake | Goal + slug + verification surface identified |
| 1 | Prerequisite audit | Blockers documented or cleared |
| 2 | Write LOOP doc | All five steps have commands + verify sub-steps |
| 3 | Dry-run | Agent runs full loop; dry-run log filled |
| 4 | Polish | Loop reliable; hand off or report BLOCKED |

## Phase 0 — Intake

1. Restate **goal** (observable success, not implementation).
2. Choose **slug** (kebab-case).
3. Map surfaces: what to build, where to deploy, how to run, what inspects success.

## Phase 1 — Prerequisite audit

Run non-destructive probes **before** writing steps:

| Pitfall | Detection | Action |
|---------|-----------|--------|
| Interactive auth | Command hangs or prompts | BLOCKER — document unblock; never ask mid-loop |
| Missing binary | `which <tool>` fails | Install via project tooling or `smc`; else BLOCKER |
| Missing secret/env | Probe fails | BLOCKER with exact env/key needed |
| Human-only UI | No scriptable path | `debug-with-user` or BLOCKER |

**Automation rule:** Every loop step must run without user confirmation. Use
non-interactive flags (`-y`, `--yes`) or piped answers (`echo -ne $'n\nn\n' |
cmd`) when the CLI has no flag. Never stop with "shall I continue?"

## Phase 2 — Write LOOP doc

Create the file at the resolved path. Each of the five sections is required:

1. **Build** — compile/bundle; verify artifact exists or build exits 0
2. **Deploy / Update** — push/upload/restart; verify target state
3. **Run** — start process or trigger action; verify started
4. **Inspect / Feedback** — health check, logs, sustained probe; verify goal metric
5. **Fix** — what to change when inspect fails; return to step 1

Every section needs:

- Copy-paste **commands** in fenced `sh` blocks
- **Verify** line: command + expected signal (exit code, substring, HTTP status, duration)

Include **Pitfalls & blockers** and an empty **Dry-run log** table.

## Phase 3 — Dry-run (mandatory)

Execute the loop yourself in order. After each step, append a row to **Dry-run log**
with timestamp, PASS/FAIL/BLOCKED, and evidence (stdout excerpt, log path).

- Step fails → fix the LOOP doc (commands, verify criteria, pitfalls) and re-run
- BLOCKER hit → set dry-run status BLOCKED; document exact user unblock in Pitfalls
- Full cycle passes → set dry-run status PASS

Do not claim the loop is established until dry-run status is PASS or BLOCKED with
clear documentation.

## Phase 4 — Polish & handoff

1. Remove dead commands; add timeouts, idempotency notes, cleanup steps
2. Confirm inspect output is agent-readable (structured, not vague)
3. Report: LOOP path, dry-run status, aux script paths (if any)
4. If goal not yet achieved but loop works → suggest `/loop-workflow <goal>`
   using this LOOP's inspect criteria

# Anti-patterns

- Stopping for user confirmation mid-loop
- LOOP doc without verify sub-steps per section
- Skipping dry-run and claiming success
- Putting Go code inside the LOOP markdown (use `script/debug/` instead)
- Vague inspect ("check it works") without measurable criteria

# Quick start

```
/establish-a-loop <what to verify and how>
```

Example: `/establish-a-loop debug ai-critic server on remote: build linux binary,
upload, start daemon, ping for 10 minutes without restart.`