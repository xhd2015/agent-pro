---
name: consolidate-code
description: >-
  Incremental deduplication workflow: map overlapping implementations, plan a
  PR stack into shared packages, seal doctests, migrate callers without behavior
  change. Use when the user mentions consolidate, dedupe, shared package,
  move enrichment into dot-pkgs, eliminate duplication, or /consolidate-code.
  Pair with investigate (overlap map) and doctest-tdd (sealed tests → impl).
---

You guide the agent through **safe, incremental code consolidation** — extracting
duplicated logic into shared packages while preserving caller-facing behavior.
The goal is one canonical implementation per concern, with rerunnable regressions
proving nothing broke.

# When to use

- Same concern implemented in 2+ places (git subprocess, porcelain parse, enrichment)
- A feature repo reinvented utilities that already exist in a shared module
- User wants PR-sized consolidation (not a big-bang refactor)
- After `investigate` identified overlap and user asked to fix it

# When not to use

- Greenfield feature with no duplication to merge
- Doc-only or comment-only changes
- Behavior change is intentional — use `brainstorm` + `doctest-tdd` as a feature
- User only wants root-cause explanation → `investigate` alone

# Related skills

| Skill | Role in consolidation |
|-------|----------------------|
| `investigate` | Map duplication matrix before moving code |
| `brainstorm` / `followup` | Lock API shapes and PR boundaries with user |
| `doctest-tdd` | Designer → RED → seal → implementer (orchestrator never edits `.go`) |
| `doctest-tdd-lite` | Single-agent TDD when overlap is small and localized |
| `explore` | Fast grep/read pass to find all copies of a pattern |

# Required workflow

## Phase 1 — Overlap map (investigate)

Build a **concern matrix** before writing code:

| Concern | Locations (absolute paths) | API differences | Callers | Consolidation target |
|---------|---------------------------|-----------------|---------|---------------------|

**Gather evidence**

- `rg` for identical patterns (`gitOutput`, `parsePorcelain`, `exec.Command("git"`)
- Read each copy; note semantic differences (e.g. wrk `changed` vs backup `modified`)
- List doctest/unit regressions that lock current behavior per caller

**Classify each concern**

- **Merge** — same semantics → one shared function
- **Parameterize** — same parser, different format/taxonomy → `FormatStyle` or parallel types (`WrkCounts` vs `Counts`)
- **Keep separate** — different jobs (mutating git + verbose log vs read-only `git/cmd`)
- **Defer** — next PR; document in REQUIREMENT-DESIGN

**Response slice:** Observation → Findings table → recommended PR stack.

## Phase 2 — PR stack plan

Split into **foundation → callers → domain glue**. Typical order:

1. **Foundation** — lowest layer with no caller imports (`git/cmd`, parsers, types)
2. **Mid-layer** — composes foundation (`git/checkout.Enrich`, `git/reposnapshot.Build`)
3. **First caller migration** — highest-value duplicate removed; regressions pass
4. **Remaining callers** — wrkcli, scan_repo, worktree, app server package
5. **Delete dead code** — only after all callers migrated

Each PR must:

- Preserve stdout/JSON/API contracts unless explicitly changing behavior
- State what stays in the caller (ANSI color, summary layout, CLI flags)
- Name locked decisions in a table (taxonomy, error durability, path callbacks)

Write `REQUIREMENT-DESIGN-<slug>.md` per PR. Get user `/go ahead` before TDD.

## Phase 3 — Test strategy (before implementation)

**Prefer doctest leaves** for behavior locked by CLI or subprocess output.

| Layer | Tests |
|-------|-------|
| New shared package | `pkg/tests/` leaves (parse, format, enrich) |
| Existing CLI | Re-run full existing doctest tree — no new leaves if behavior unchanged |
| App integration | Targeted regression leaves (e.g. `git-repos-summary`) |

Rules:

- Seal tests once (`git add tests/`) — implementer must not modify sealed trees
- External/gitignored modules: seal in the **canonical module repo**, not the consumer
- For CLI stdout: assert ends with trailing newline when using doctest templates

## Phase 4 — Implement (doctest-tdd)

Default: **orchestrator delegates** — never hand-edit `.go` in TDD mode.

```
designer → doctest vet → doctest test (RED) → git add tests → implementer → GREEN
```

Implementer checklist:

- Resolve **name collisions** early (const vs function in same package)
- Route read-only git through shared `cmd.Run`; keep mutating `exec.Cmd` where logging/streaming differs
- Caller wrappers stay thin (color, JSON map, summary section headers)
- **Durable enrichment** — partial fields + per-step `error`; never fail whole scan for one bad repo

## Phase 5 — Verify

```sh
doctest vet ./<new-test-tree>
doctest test ./<new-test-tree>/...
doctest test ./<caller-regression-trees>/...
go test ./<affected-packages>/... -count=1
git diff ./<sealed-test-tree>   # must be clean
```

Report: pass counts, lines removed per file, accepted test amendments (with rationale).

# Consolidation patterns (proven)

| Pattern | Example |
|---------|---------|
| Subprocess layer | `git/cmd.Run` replaces private `gitOutput` in N packages |
| Parser + formatter split | `ParsePorcelain` + `Format(Backup)`; wrk gets `ParsePorcelainWrk` + `FormatWrk` |
| Options struct | `checkout.Options{StatusStyle, PorcelainUntracked}` for caller variants |
| Export list parser | `worktree.ParseListPorcelain`; `scan_repo` calls `worktree.ListCtx` |
| Domain glue stays up-stack | `machinebackup` keeps JSON version + `GIT REPOS:` summary layout |
| Is-clean helper | `worktree.IsCleanWrk` — wrkcli delegates instead of inline parse |

# Anti-patterns

- Big-bang delete-then-fix across repos
- Merging different semantics because strings look similar
- Moving ANSI/CLI formatting into shared packages
- Modifying sealed doctests to match a wrong implementation
- Skipping caller regressions because unit tests pass

# Response shape

1. **Overlap summary** — concern matrix (or PR-N scope if continuing a stack)
2. **PR plan** — ordered list with in/out of scope per PR
3. **Locked decisions** — table (API, taxonomy, what stays in caller)
4. **Test plan** — new leaves + regression commands
5. **Next action** — `REQUIREMENT-DESIGN` draft or spawn designer (after user go-ahead)

# Commands

```sh
agent-pro skill consolidate-code --show
agent-pro skill consolidate-code install
agent-pro skill investigate --show
agent-pro skill doctest-tdd --show    # if installed in .agents/skills/
```

# Case study (dot-pkgs git stack)

PR-1–4: `git/cmd`, `git/status`, `git/checkout`, `git/reposnapshot` ← `machinebackup/git_repos.go`
PR-5: wrkcli → `checkout.Enrich` + `FormatWrk`; 35 wrk status doctests unchanged
PR-6: `scan_repo` → `worktree.ListCtx`; `IsCleanWrk`; deleted duplicate parsers

Lesson: wrk and backup **share subprocess and parse machinery** but **not format strings**.