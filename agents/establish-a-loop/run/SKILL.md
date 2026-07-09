---
name: establish-a-loop
description: >-
  Establish an agent-runnable verification loop as LOOP_<YYYY-MM-DD>_<slug>.md
  with build, deploy/update, run, inspect, and fix steps. Classifies loop kind
  (bug-repro, health-check, regression), reproduces user-reported symptoms before
  any fix, dry-runs without user input, and polishes until reliable. Use when the
  user runs /establish-a-loop, says "establish a loop", "create a debug loop",
  or needs a self-verifiable runbook before iterating on a fix. Pairs with
  loop-workflow for goal-driven execution.
---

You establish a **self-verifiable loop** the agent can run end-to-end without
user input. Deliverable: `LOOP_<YYYY-MM-DD>_<slug>.md`. Pair with `loop-workflow`
after the loop is verified.

# When to use

- User needs a repeatable build → deploy → run → inspect → fix cycle documented
- User reports a **symptom** (log lines, errors) and wants a loop to reproduce it
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

# Loop kinds

Classify every request in Phase 0. The kind controls what inspect must prove.

| Kind | User signal | Inspect in steps 1–4 | Dry-run complete when |
|------|-------------|----------------------|------------------------|
| **bug-repro** | "I see X in logs", pasted errors, failing status | Symptom **present** (RED) | **REPRO PASS** |
| **health-check** | "verify ping for 10 min", "ensure daemon stays up" | Healthy end state (GREEN) | **VERIFY PASS** |
| **regression** | "after fix, ensure X never returns" | Symptom present, then absent post-fix | **REPRO PASS** then **VERIFY PASS** |

Default to **bug-repro** when the user pastes failure logs or error excerpts.

# Aux scripts (optional)

The LOOP doc is **markdown only**. When a step needs code (sustained probe,
parsing, multi-check orchestration), add:

```
script/debug/<purpose>/main.go
```

Reference the script path from the LOOP doc. Do not embed Go in the LOOP file.
Follow `loop-workflow` inspect conventions if the aux script gates success
(`PASS:` / `FAIL:` or `REPRO:` / `VERIFY:`, exit 0 = pass for the active mode).

**Inspect script rules by kind:**

| Kind | First script behavior |
|------|----------------------|
| `bug-repro` | Exits **non-zero** on unfixed system; prints `REPRO:` + matched symptom lines |
| `health-check` | Exits **zero** when healthy (GREEN gate only) |
| `regression` | Repro mode (default): non-zero + `REPRO:`; add `--expect=healthy` or `*-verify` script for GREEN |

Do not ship a bug-repro inspect script that only passes when the system is already
healthy.

# Phases

| Phase | Name | Gate |
|-------|------|------|
| 0 | Intake | Kind + symptom/goal + slug + verification surface |
| 1 | Prerequisite audit | Blockers documented or cleared (non-destructive only) |
| 2 | Write LOOP doc | All steps + kind-specific inspect |
| 3 | Dry-run | Repro recorded first; dry-run log filled |
| 4 | Polish | Loop reliable; hand off or report BLOCKED |

## Phase 0 — Intake

1. Classify **loop kind**: `bug-repro` | `health-check` | `regression`.
2. For `bug-repro` / `regression`: capture the user's **symptom verbatim**
   (log lines, status fields, exit codes). This becomes the step-4 inspect
   target — not the desired end state.
3. Restate **goal** in kind-appropriate terms:
   - `bug-repro`: "steps 1–4 reproduce the symptom on command"
   - `health-check`: "steps 1–4 confirm healthy end state"
   - `regression`: "steps 1–4 reproduce symptom; after step 5, symptom gone"
4. Choose **slug** (kebab-case).
5. Map surfaces: what to build, where to deploy, how to run, what inspect checks.

## Phase 1 — Prerequisite audit

Run **non-destructive** probes before writing steps:

| Pitfall | Detection | Action |
|---------|-----------|--------|
| Interactive auth | Command hangs or prompts | BLOCKER — document unblock; never ask mid-loop |
| Missing binary | `which <tool>` fails | Install via project tooling or `smc`; else BLOCKER |
| Missing secret/env | Probe fails | BLOCKER with exact env/key needed |
| Human-only UI | No scriptable path | `debug-with-user` or BLOCKER |

**No remediation during audit** (for `bug-repro` / `regression`): do not `mkdir`,
edit config, install packages, or restart services to "fix" the symptom while
establishing the loop. Record repro **preconditions** (e.g. "WorkingDir missing")
in the LOOP doc Symptom section — satisfy them only in **step 5 Fix**.

**Automation rule:** Every loop step must run without user confirmation. Use
non-interactive flags (`-y`, `--yes`) or piped answers (`echo -ne $'n\nn\n' |
cmd`) when the CLI has no flag. Never stop with "shall I continue?"

## Phase 2 — Write LOOP doc

Create the file at the resolved path. Required sections depend on kind; see
`references/LOOP-template.md`.

**All kinds** — steps 1–5:

1. **Build** — compile/bundle; verify artifact exists or build exits 0
2. **Deploy / Update** — push/upload/restart; verify target state
3. **Run** — trigger the action under test (**no remedial setup** for bug-repro)
4. **Inspect / Feedback** — see kind-specific rules below
5. **Fix** — what to change when repro inspect fails; return to step 1

Every section needs:

- Copy-paste **commands** in fenced `sh` blocks
- **Verify** line: command + expected signal (exit code, substring, HTTP status)

Include YAML frontmatter (`title`, `created`, `slug`, `path`, `loop_kind`,
`dry_run_status`), **Symptom** (when applicable), **Pitfalls & blockers**, and an
empty **Dry-run log** table. See `references/LOOP-template.md`.

### Bug-repro rules (mandatory for `bug-repro` and `regression`)

1. **Run (step 3) triggers the failure only** — no `mkdir`, config edits, installs,
   or other remedies in steps 1–4. Those belong **only in step 5 Fix**.
2. **Inspect (step 4) asserts the symptom** the user reported:
   - exact log substrings from the Symptom section, and/or
   - `status: error` + `Last Error: …`, and/or
   - inspect script exits non-zero and prints `REPRO:` with evidence.
3. Document repro **preconditions** in Symptom (state required to see the failure).
   Do not auto-satisfy preconditions during dry-run steps 1–4.
4. For `regression`, add **step 4b Inspect — Verify** (post-fix GREEN) after step 5.

### Health-check rules

Step 4 inspect asserts the healthy end state (GREEN). No separate repro phase.

## Phase 3 — Dry-run (mandatory)

Execute the loop yourself in order. After each step, append a row to **Dry-run log**
with timestamp, result, and evidence (stdout excerpt, log path).

**Order for `bug-repro` / `regression`:**

1. Run steps 1–4 **without applying step 5**.
2. Record **REPRO PASS** only when step 4 shows the user's symptom (inspect RED).
3. For `regression` only: apply step 5, re-run 3–4b, record **VERIFY PASS**.

**Dry-run status values:**

| Status | Meaning |
|--------|---------|
| **REPRO PASS** | Steps 1–4 reliably produce the user's symptom |
| **VERIFY PASS** | After step 5, inspect GREEN (`health-check` or regression post-fix) |
| **BLOCKED** | Prerequisite missing; unblock documented in Pitfalls |

- `bug-repro`: establishment complete at **REPRO PASS** — do **not** require GREEN.
- `health-check`: establishment complete at **VERIFY PASS**.
- `regression`: require **REPRO PASS** first (with failure excerpt), then **VERIFY PASS**.

If step 4 does not show the symptom → fix LOOP commands/verify criteria, not the
environment. Do not loosen inspect assertions to get PASS.

Do not claim the loop is established until the kind-appropriate dry-run status is
recorded or **BLOCKED** with clear documentation.

## Phase 4 — Polish & handoff

1. Remove dead commands; add timeouts, idempotency notes, cleanup steps
2. Confirm inspect output is agent-readable (structured, not vague)
3. Report: LOOP path, loop kind, dry-run status (REPRO / VERIFY / BLOCKED), aux
   script paths (if any)
4. For `bug-repro` with REPRO PASS → suggest `/loop-workflow <fix goal>` using
   step 4b verify criteria (or `--expect=healthy` inspect) as the GREEN gate
5. For `health-check` with VERIFY PASS → hand off to monitoring or close

# Anti-patterns

- **Fixing during establishment** — applying mkdir, config change, or install
  during Phase 1–3 before repro inspect is recorded
- **Baking the fix into Run** — step 3 must not include the remedy from step 5
- **GREEN-only inspect for a reported bug** — inspect that only passes when
  healthy cannot reproduce what the user sees
- **Loosening verify to get PASS** — widening log tail or dropping symptom
  assertions after inspect fails post-fix
- **Conflating "loop runs" with "symptom reproduced"** — REPRO PASS requires
  evidence of the user's failure signature in the dry-run log
- Stopping for user confirmation mid-loop
- LOOP doc without verify sub-steps per section
- Skipping dry-run and claiming success
- Putting Go code inside the LOOP markdown (use `script/debug/` instead)
- Vague inspect ("check it works") without measurable criteria

# Quick start

```
/establish-a-loop <what to verify and how>
```

Examples:

```
/establish-a-loop reproduce openclaw service log error: fork/exec /bin/bash —
build linux binary, deploy, restart service, tail service log
```

```
/establish-a-loop debug ai-critic server on remote: build linux binary, upload,
start daemon, ping for 10 minutes without restart
```