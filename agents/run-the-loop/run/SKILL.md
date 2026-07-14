---
name: run-the-loop
description: >-
  Fix until done: goal + inspect script under script/debug/, then iterate
  implement → inspect until PASS. Use after establish-a-loop or standalone for
  /run-the-loop, "loop until done", "max effort", "don't stop until".
---

Iterate until an inspect script proves the goal. **Pairing:** if `LOOP_*.md` exists,
read it in Phase 1 — reuse Build/Deploy/Derived ops and use step **4b** (or invert
step 4 `REPRO:`) as the GREEN spec.

# When to use

- Fix/implement until a measurable goal is achieved
- User wants persistent verify-and-fix cycles with evidence
- **After** `establish-a-loop` when SYMPTOM CONFIRMED and step 4b defines GREEN

# Phases

| Phase | Name |
|-------|------|
| 1 | Goal |
| 2 | Inspect script |
| 3 | Fix loop |
| 4 | Complete |

## Phase 1 — Goal

1. Restate goal in one sentence; call `update_goal` with progress message.
2. **If** `doc/LOOP_*.md` (or `docs/`) exists for this task: read Derived operations,
   Workspace tools, step 4b — fold into goal and GREEN criteria.
3. **If** blocking unknowns: ask ≤5 concrete questions (one message); else proceed.
4. Pick inspect path: `script/debug/<goal>/main.go` (kebab-case).

## Phase 2 — Inspect script

Single source of truth for "done." See `references/inspect-script-template.go`.

| Requirement | Detail |
|-------------|--------|
| Runnable | `go run ./script/debug/<path>` from repo root |
| Exit 0 | Goal met; non-zero = not met |
| Output | `PASS:` or `FAIL:` in result block |
| Evidence | Artifacts under `.../out/` |
| Surfaces | UI → `playwright-debug`; API → `curl`; CLI → `os/exec`; CI → `gh run view` |

1. Write `main.go` for GREEN criteria.
2. Run it — expect **FAIL** initially (RED). Fix script until failure is actionable.
3. Gate: agent-readable FAIL output before Phase 3.

## Phase 3 — Fix loop

```
Implement → go run ./script/debug/<path> → exit 0?
                ↑ no: analyze FAIL, hypothesize, repeat
```

Rules:

- Run inspect after every meaningful change; cite `PASS:`/`FAIL:` in progress.
- Root cause before workaround (`sound-fix` when unsure).
- Run affected tests when touching production code.
- `update_goal(message=...)` each iteration.
- **If** same failure 3×: add `LOOP_DEBUG:` logs, then continue.

Stop only when: inspect exits 0 and output matches goal → `update_goal(completed=true)`.
Or: `blocked_reason=...` after one user question (missing creds / 5× same error).

**Never** stop on "should work", eyeballing, or tests green without re-running inspect.

## Phase 4 — Complete

Summarize goal, iterations, key fix, inspect path. Paste final PASS output (abbreviated).
List `out/` artifacts. Keep inspect script for regression.

# Anti-patterns

- Skipping inspect or writing it after claiming success
- `update_goal(completed=true)` without fresh inspect PASS
- Stopping at `go test` when goal needs UI/CI inspect
- Ask-wait-ask clarification loops
- Deleting inspect script after success

# Quick start

```
/run-the-loop <fix goal; optional: verify via screenshot, CI workflow name, etc.>
```