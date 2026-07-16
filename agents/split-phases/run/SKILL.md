---
name: split-phases
description: >-
  Break a large implementation plan into dependency-ordered phases that are
  independently implementable and verifiable. Use when the user runs /split-phases,
  or says "split into phases", "phase this plan", "break into units of work",
  "dependency-ordered milestones", or wants a giant plan split for phase-by-phase work.
---

# split-phases

You turn a large implementation plan into **units of work** that can each be
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
3. **Phase budget (soft).** Target **3–5** phases for a typical plan — that is the
   center of the distribution, not a hard cap. **2** is fine when the plan is already
   small. **6–8+** is fine when independent git/module boundaries force separate
   implement-and-verify units. Always give a **one-line reason** when outside 3–5.
   Do not invent phases to fill a quota; do not shred a coherent unit into micro-phases.
4. **No domain invention.** Do not add features outside the user’s plan. If something
   is ambiguous and blocking, ask; otherwise state brief assumptions and proceed.
5. **Every phase must be verifiable.** If you cannot name exit criteria, merge with a
   neighbor or refine until you can — do not invent a micro-phase just to hold a check.
6. **Cluster by boundary, then by dependency.** Order of consideration:
   (1) git project / worktree roots the plan must touch,
   (2) module manifests (`go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`, …),
   (3) packages / dirs as units of work.
   Boundary levels are **split axes**, not mandatory phase layers — skip levels that
   do not apply (e.g. single-module plans need no multi-repo phases).

## Principles

1. **Independent implementability** — A phase must not require unfinished later phases.
   Earlier phases may leave stubs/seams for later ones.
2. **Independent verifiability** — Each phase has explicit exit criteria (commands,
   tests, assertions, or review checks) that can pass without later phases done.
3. **Boundary-first clustering** — Group work along git → module → package before
   slicing by capability. Multiple packages may share one phase when they are one
   vertical slice; one package is not automatically multiple phases.
4. **Dependency-first ordering** — Within and across phases, sort so units with
   **fewer dependencies come first**. Pure helpers, types, contracts, and validation
   before orchestration, I/O, and polish — usually as **tasks inside a phase**, not
   each as its own phase.
5. **MECE work units** — Minimize overlap. Shared foundations in early phases; product
   polish last.
6. **Right size** — Prefer a phase that is one **coherent PR-sized** unit along a
   **boundary** (often one package or one module slice). Prefer **merging** over
   micro-phases. Split further only for hard boundaries or mixed unrelated risks.
7. **Thin vertical slice** — After the full graph, name the **smallest subset** of
   phases that proves end-to-end value (demo path).

## Workflow

```text
1. Ingest the plan (user message, design doc, prior brainstorm output)
2. Clarify only if blocking (missing success definition, unclear system boundary)
3. Identify git project / worktree roots involved (level 1)
4. Within each root, identify modules (go.mod, package.json, …) (level 2)
5. Within each module, identify packages / dirs as units of work (level 3)
6. Extract work items and true dependencies; cluster into ~3–5 phases along
   those boundaries (merge unless boundaries or risks force a split)
7. Topological sort (low dependency first); note parallel tracks if any
8. Emit the required output format; note why if phase count is outside 3–5
9. STOP — do not implement unless user explicitly requests a phase
```

**Worktree note:** Treat “git project / worktree” as distinct working trees or repos
the plan must touch. Assign phases across them when isolation matters. Do **not**
invent worktree setup unless the user’s plan already includes it.

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
| **Boundary** | Git root / module / package(s) this phase owns (omit only if truly N/A) |
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

If phase count is outside 3–5, include a **one-line phase-budget note** in Source summary
or Next action (e.g. “6 phases: two git roots + integration”).

## Boundary hierarchy

Cluster top-down. Stop at the coarsest level that still yields verifiable, independent phases.

| Level | What | Prefer separate phases when… | Prefer same phase when… |
|-------|------|------------------------------|-------------------------|
| **1. Git / worktree** | Distinct clone roots or intentional worktree isolation | Different repos, release cadences, or verify isolation | Single repo owns the whole plan |
| **2. Module** | Build/dependency unit (`go.mod`, `package.json`, …) | Own build/tests/publish or clear API boundary | One module only, or cross-module work is thin glue later |
| **3. Package** | Package/dir as **unit of work**, not every file | Different deps/owners/risk and separate exit criteria | Related packages form one vertical slice |

**Do not** split below package level (per-function / per-helper / per-file) unless exit
criteria and risk truly diverge — that is the main anti-pattern this skill avoids.

## Splitting heuristics

| Prefer earlier | Prefer later |
|----------------|--------------|
| Work confined to one git root | Cross-repo integration |
| Foundational module / API surface | Consumer modules |
| Core package under a module | Wiring packages / apps |
| Pure functions, types, contracts | UI polish, docs, telemetry/events |
| Public surface + validation (flags, API) | Full orchestration |
| Read-only plan/analyze paths | Mutating apply paths |
| Happy path | Exhaustive edge-case tests |
| Stub-able seams | End-to-end product UX |

**Smell → merge (prefer this when in doubt):**

- Phase is only types + one helper with no standalone product value
- Two phases share the same package and the same exit command
- Phase would be a tiny tightly coupled set that cannot verify alone
- “Part 1 / Part 2” of the same package with no independent exit criteria
- Padding to reach 3 phases when 2 already cover the plan

**Smell → split further:**

- Mixed unrelated risks across **boundaries** (e.g. two git roots, or module API + unrelated consumer rewrite)
- Phase needs types/APIs from a later phase → extract foundation phase or reorder
- No exit criteria → not a phase yet; merge or refine
- “Implement engine and wire CLI and write all docs” **and** those span different modules/packages with separate verify paths

**Exit criteria quality bar:**

- Prefer **rerunnable** checks (doctests, unit tests, CLI assertions)
- Manual checks OK when automation is disproportionate — still write them as explicit steps
- “Looks good” alone is not enough; say what to observe

## Examples of good vs bad phases

**Bad (too small — same package shredded):**  
P1: add type `Foo` · P2: add `ValidateFoo` · P3: add `ParseFoo` — all one package.

**Good:** one phase “`foo` package: types + parse + validate + tests.”

**Bad:** “P2: implement the feature.” / “P4: finish everything left.”

**Good:** “P2: pure `IsFoo` helpers + table tests inside package `foo`; no I/O.”  
(OK as a phase only if that package/slice is the whole unit — do not also split parse/validate next door into more phases.)

**Bad (ignores boundary):** one phase edits two git repos + three modules with a single exit criterion.

**Good (multi-boundary, still on budget):**  
P1: repo A / module `core` packages for the API → P2: repo A / module `cli` wiring →  
P3: repo B consumer → P4: cross-repo integration / e2e.  
(4 phases — within 3–5.)

**Good (small plan):** 2 phases — core package + thin CLI wire.  
Phase-budget note: “2 phases: single-module plan.”

## Persistence

- Default: **chat only**
- Write `PHASES.md` or similar **only when the user explicitly asks** to save the plan

## Stop condition

After delivering the phase plan (and optional file **if requested**), **stop**.

Do not implement the underlying work under this skill. Implementation is a separate
user instruction naming the phase(s) to do.
