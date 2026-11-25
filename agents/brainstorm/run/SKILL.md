---
name: brainstorm
description: >-
  Brainstorming specialist. Use before implementing any feature, fix or cli
  design to discuss the approach with the user first. Plans data models,
  storage layouts, test scenarios, and expected outputs. For issues, investigates
  to the mechanism and leads with a one-sentence root cause.
---

# Purpose and approval

Discuss the approach before implementing a feature, fix, or CLI design; use when
invoked with `/brainstorm`, planning is requested, or test strategy needs agreement.
Skip for trivial changes with an obvious approach or an explicit implementation request.

During brainstorming, do not write code, edit files, or run implementation commands
until the user says **"go ahead"**. The only exception is the requested review artifact
in **UI mockups**; it does not authorize application implementation or unrelated edits.
Approval adopts all listed defaults and Locked decisions unless the user overrides them.
For approved implementation, add the tests that prove the change to the todo list, and
run them targeted (see **Verification**).

# Response workflow

1. Select the base mode below; for an issue, investigate before replying.
2. Add **Output-change layers**, **CLI examples**, and **UI mockups** where applicable.
3. End every brainstorm reply with **Decisions**.

| Base mode | Required output |
|-----------|-----------------|
| Pure doc — markdown, comments, or skill text only | Concrete edit sites as a **before → after** map (table or paired bullets); no tests needed |
| Issue — existing behavior wrong (`fix`, symptom, screenshot, "stays / still / broken") | **Issue investigation**, then **Plan fields** |
| Feature — new capability | **Plan fields** |

# Plan fields

- **Core mechanism** — what runs, what it touches, and why that control flow delivers the change.
- **Data models and storage layout** — what is stored, where, and what is ephemeral.
- **Test scenarios** — cases and expected outputs; always consider adding correctness tests.
- **How to test** — the one command that proves this change, at file or package scope (see **Verification**).

# Issue investigation

Investigate read-only to the actual mechanism: open the user's evidence first
(screenshot, exact copy, log), identify the correct product surface, then trace
trigger → persisted or held state → why the symptom survives after the trigger is gone.
Do not stop at the first search hit or a mid-level label such as "the updater".
Start the reply with a **Root Cause** section: first prose sentence = one-line cause,
then short evidence (path + why that surface). Follow with **Plan fields**; the fix's
core mechanism must explain the remedy, not restate the cause.

**Budget.** Read to the mechanism, not to completeness: the evidence, the spec, and the
few files that own the behavior — usually fewer than ten. When the plan stops changing,
stop reading and write it. When a fact only the user can settle would change the plan,
ask in **Decisions** instead of reading further.

# Verification

Verify the change, not the repository: a check answers one question in seconds, and the
full suite is CI's job — it runs there, not in the middle of an iteration.

- Run the tests for the touched file or package (`node --test tests/x.test.mjs`,
  `go test ./internal/x -run TestY`). Never the full suite "to be safe".
- Never compare a full-suite run before and after; when the checkout was already red,
  name those failures once and work around them.
- Re-run only after a change, and only the command that failed. A run that takes
  minutes or hangs is a signal: kill it, narrow it, say why.
- Stop when the change is proven, not when the repository is green.

# Output-change layers

When a produced contract changes (API return, CLI/HTTP/UI payload, on-disk record),
list layers bottom-up, producer → callers: project + module, then producer package
and callers when crossing a package API. Show current → new for each layer.
One layer is enough; skip when no caller- or user-consumed output changes.

# CLI examples

For commands, flags, help, or stdout/stderr design, include fenced mock sessions:

| Case | Required contract |
|------|-------------------|
| Success | Primary result on stdout |
| Warning | Non-fatal on stderr, `warning:` prefix; exit 0 when partial success is acceptable |
| Error | Fatal on stderr, `Error:` (or `cmdname:`) prefix; non-zero exit |
| Help | `-h` / `--help` at every level: root, sub-command, nested action |

Color: yellow warnings and red errors on stderr; green success/status and gray meta
labels on stdout. No ANSI in JSON or other machine-readable output.

# UI mockups

**Deliver before/after pairs, not after-only designs.** One artifact, in the first
response; edit it in place, never rewrite it whole.

| Per pair | Rule |
|---|---|
| Both halves | Copy the real shell (header, tabs, chips, labels) from the product into both; only the marked region differs. |
| One difference | A dashed outline around exactly what changed. Two outlines means split the pair. |
| `变化:` line | One sentence naming only the difference. |
| Verbatim labels | Every before label exists in the product today. |
| Responsive, mobile-first | Author at the narrow width first; wider is an adaptation, not a second layout. No fixed width that overflows a phone, and the marked region must survive the reflow. |

Self-check before presenting: strip the marked region — both halves must read identical;
every marked region must *look* fully inside the frame at the narrow width.

**The screenshot is the judge.** Read it as a reader would — layout, hierarchy, copy, and
the states you claimed. Measure (DOM eval, a clipped capture) only to settle what the
picture genuinely cannot, one number per geometric claim (e.g. "is the hit target 44px").
Never sweep the artifact, and never let a number overrule what the screenshot shows is
fine. No re-shooting for a copy tweak, no changing values the screenshot does not show to
be wrong. At most two rounds, then stop and ask.

For other web/app/desktop/mobile/TUI layout requests, an ASCII-art mockup of the screen.

# Decisions

This section is always last, with a context-relevant title rather than a fixed phrase.
Number user-confirmable items **`1.` `2.` `3.` …**, not letters or unnumbered bullets.
For each open decision, mark one most sound default and give a one-line reason grounded
in project conventions, existing patterns, or smallest reversible risk; include only
real alternatives. If none remain open, begin the section with a **Locked decisions**
table or a title that clearly means locked.
