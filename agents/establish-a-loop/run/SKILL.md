---
name: establish-a-loop
description: >-
  Produce LOOP_<YYYY-MM-DD>_<slug>.md — a real runnable repro/verify runbook from
  user logs or prose. Expands to precise CLI/URL ops, builds workspace tools,
  runs steps 1–4 for real, records SYMPTOM CONFIRMED or VERIFY PASS. Pair with
  run-the-loop for fixes. Use for /establish-a-loop, "establish a loop", or
  "create a debug loop".
---

Deliverable: `LOOP_<YYYY-MM-DD>_<slug>.md`. **Pairing:** after Phase 4, hand off to
`run-the-loop` for code/env fixes until GREEN (LOOP step 4b).

# When to use

- User pastes logs or prose steps and needs a repeatable repro/verify runbook
- Verification spans build, deploy, run, inspect (local or remote)
- **Not** when a LOOP doc already exists and the task is fix-only → `run-the-loop`

# Loop kinds

| Kind | Inspect in steps 1–4 | Establishment complete |
|------|----------------------|------------------------|
| **bug-repro** (default on log paste) | Symptom present (RED) | **SYMPTOM CONFIRMED** |
| **health-check** | Healthy state (GREEN) | **VERIFY PASS** |
| **regression** | Symptom then absent post-adjust | **SYMPTOM CONFIRMED** → **VERIFY PASS** |

# LOOP doc steps (reference)

```
Build → Deploy → Run → Inspect → Adjust (documented; code/env only in step 5)
  ▲___________________________________________|
```

Steps 1–4 run **for real** (not simulated). Details: `references/LOOP-template.md`.

# Phases

| Phase | Name |
|-------|------|
| 1 | Plan |
| 2 | Write LOOP |
| 3 | Run establishment |
| 4 | Handoff |

## Phase 1 — Plan

1. **Classify** kind, pick kebab **slug**, note verification surface.
2. **Expand to operations** (always): copy user paste → **User input (verbatim)**;
   derive **Derived operations** (exact CLI/`curl`/`playwright-debug` URL per row);
   list **Workspace tools** for any `./cmd/*` used (Build + Deploy — never `which` alone).
3. **Audit** prerequisites (non-destructive). BLOCKER → document in Pitfalls; no fixes.
4. **If** symptom vague or step-4 criteria unclear: minimal real trigger, save evidence
   under `script/debug/<slug>/out/observation/`, add **Observation notes**.

Bug-repro/regression: do not mkdir/config/install to change the symptom during planning.

## Phase 2 — Write LOOP

Create file at resolved path (see template for path rules and all sections):

```
doc/LOOP_<date>_<slug>.md  → else docs/ → else docs/loops/ (mkdir -p)
```

- Command blocks in steps 1–4 must match **Derived operations**.
- Steps 1–5: Build (workspace tools), Deploy, Run (trigger only), Inspect, Adjust
  (instrument | code | env — document only for bug-repro code/env).
- Aux repro: `script/debug/<slug>/` via `go run`; `REPRO:` + non-zero on unfixed system.
- Every step: fenced `sh` + **Verify** line.

## Phase 3 — Run establishment

Execute LOOP steps **1–4 for real**. Append **Run log** rows with evidence.

1. Bug-repro/regression: steps 1–4 **without** step 5 code/env adjust.
2. Set `establishment_status` when gate met (see loop kinds table).
3. Regression: apply step 5 code/env adjust, re-run 3→4b, then **VERIFY PASS**.

**If step 4 fails:** update Derived ops / inspect criteria; **or** instrument-only
Adjust (`LOOP_DEBUG:` logs), rebuild, **repeat Phase 3**. No code/env fix for bug-repro
establishment — that is `run-the-loop`.

| Status | Meaning |
|--------|---------|
| **SYMPTOM CONFIRMED** | User's failure reproduced on command |
| **VERIFY PASS** | Healthy / fixed per step 4b |
| **BLOCKED** | Prerequisite missing |

## Phase 4 — Handoff

Tighten LOOP (timeouts, cleanup, idempotency). Report: path, kind, status, aux scripts.

- **SYMPTOM CONFIRMED** → `/run-the-loop <fix goal>` using step **4b** as GREEN gate.
- **VERIFY PASS** (health-check) → close or monitor.

# Anti-patterns

- Code/env adjust during bug-repro establishment (before SYMPTOM CONFIRMED)
- Remedies in Run (step 3) that belong in Adjust (step 5)
- GREEN-only inspect for a reported bug
- Loosening inspect assertions to force a gate
- Prose commands ("open the app") without derived URL/CLI
- `which` for workspace CLIs when `./cmd/<tool>` exists
- Skipping the real establishment run
- Go code inside LOOP markdown
- Treating SYMPTOM CONFIRMED as "fixed"

# Quick start

```
/establish-a-loop <logs, prose steps, or what to verify>
```