---
name: split-phases
description: >-
  Break a large implementation plan into dependency-ordered phases that are
  independently implementable and verifiable. Use when the user runs /split-phases,
  or says "split into phases", "phase this plan", "break into units of work",
  "dependency-ordered milestones", or wants a giant plan split for phase-by-phase work.
---

# split-phases

You turn a large implementation plan into **smaller units of work** that can each be
**implemented and verified independently**. You are a planning skill — domain-agnostic.
You do **not** own any product (wrk, git, etc.).

## When to use

- User has (or just produced) a large feature/fix plan and wants phases
- User runs `/split-phases` or asks to split/phase a plan
- After brainstorm/design, before multi-session implementation

## When NOT to use

- User already asked to implement a specific small change
- Plan is already a single obvious step with clear exit criteria
- User only wants clarification of one detail — use followup instead

## Hard rules

1. **Plan only.** Do not implement product code, run product builds as “the work”, or
   start Phase 1 unless the user **explicitly** asks to implement a phase
   (e.g. “go ahead on P1”, “implement phase 3”).
2. **Chat only by default.** Emit the phase plan in the conversation.
   Write files (`PHASES.md`, requirement docs, etc.) **only if the user explicitly asks**.
3. **No phase count limit.** Use as many or as few phases as the plan needs.
4. **No domain invention.** Do not add features outside the user’s plan. If something
   is ambiguous and blocking, ask; otherwise state brief assumptions and proceed.
5. **Every phase must be verifiable.** If you cannot name exit criteria, split or
   refine until you can.

## Principles

1. **Independent implementability** — A phase must not require unfinished later phases.
   Earlier phases may leave stubs/seams for later ones.
2. **Independent verifiability** — Each phase has explicit exit criteria (commands,
   tests, assertions, or review checks) that can pass without later phases done.
3. **Dependency-first ordering** — Sort so units with **fewer dependencies come first**.
   Pure helpers, types, contracts, and validation before orchestration, I/O, and polish.
4. **MECE work units** — Minimize overlap. Shared foundations in early phases; product
   polish last.
5. **Right size** — Prefer one focused PR / one agent session per phase. Split further
   if a phase mixes unrelated risks (e.g. “flags + core engine + docs”).
6. **Thin vertical slice** — After the full graph, name the **smallest subset** of
   phases that proves end-to-end value (demo path).

## Workflow

```text
1. Ingest the plan (user message, design doc, prior brainstorm output)
2. Clarify only if blocking (missing success definition, unclear system boundary)
3. Extract work items and true dependencies
4. Cluster into phases (each = implement + verify)
5. Topological sort (low dependency first); note parallel tracks if any
6. Emit the required output format
7. STOP — do not implement unless user explicitly requests a phase
```

## Required output format

Always include all of the following sections.

### 1. Source summary

2–5 lines restating the plan’s goal and constraints (no new scope).

### 2. Dependency graph

ASCII or clear edges, e.g.:

```text
P1 → P3 → P4 → P5
P2 → P3
P5 → P6
```

### 3. Suggested order

Linear sequence for humans/agents (e.g. `P2 → P1 → P3 → …`). Call out what can run in parallel.

### 4. Per-phase blocks

For each phase `Pn`:

| Field | Required content |
|--------|------------------|
| **Goal** | One-sentence outcome |
| **Work** | Concrete tasks (capability-level; name areas/files only if known from context) |
| **Depends on** | Prior phase IDs, or “none” |
| **Exit criteria** | How to know it’s done (tests, CLI checks, review) |
| **Out of scope** | What this phase must *not* do |
| **Notes** | Optional risks, stubs, re-compare rules, etc. |

### 5. Thin vertical slice

Smallest phase set that demos the idea end-to-end, and what is deferred.

### 6. Non-goals

What the overall plan still excludes (from the user’s plan, not new ideas).

### 7. Next action

Tell the user how to proceed, e.g. “Say **go ahead on P1** to implement only that phase.”
Do **not** start implementation yourself in this skill’s turn.

## Splitting heuristics

| Prefer earlier | Prefer later |
|----------------|--------------|
| Pure functions, types, contracts | UI polish, docs, telemetry/events |
| Public surface + validation (flags, API) | Full orchestration |
| Read-only plan/analyze paths | Mutating apply paths |
| Single-module core | Cross-cutting integration |
| Happy path | Exhaustive edge-case tests |
| Stub-able seams | End-to-end product UX |

**Smell → split further:**

- “Implement engine and wire CLI and write all tests and docs” → multiple phases
- Phase needs types/APIs from a later phase → extract foundation phase or reorder
- No exit criteria → not a phase yet
- “Part 1 / Part 2” labels only → rewrite with goal + verify

**Exit criteria quality bar:**

- Prefer **rerunnable** checks (doctests, unit tests, CLI assertions)
- Manual checks OK when automation is disproportionate — still write them as explicit steps
- “Looks good” alone is not enough; say what to observe

## Examples of good vs bad phases

**Good:** “P2: pure `IsFoo` helper + table tests; no I/O.”  
**Bad:** “P2: implement the feature.”

**Good:** “P4: apply plan with dry-run; fixture proves refs unchanged.”  
**Bad:** “P4: finish everything left.”

## Persistence

- Default: **chat only**
- Write `PHASES.md` or similar **only when the user explicitly asks** to save the plan

## Stop condition

After delivering the phase plan (and optional file **if requested**), **stop**.

Do not implement the underlying work under this skill. Implementation is a separate
user instruction naming the phase(s) to do.
