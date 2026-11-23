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

For an **HTML mockup request**, immediately create a small self-contained HTML file and
present the actual file in the **first response** (file-presentation tool if available,
otherwise a file link). Do not wait for approval or preliminary design answers: use
reasonable defaults, label assumptions, and let the user review the draft — a draft the
user cannot see yet is not a draft. Keep the states under discussion (collapsed and
expanded at one narrow width), edit it in place, and never rewrite the file whole.

Review it from one screenshot as a reader would — layout, hierarchy, copy, and the
states you claimed. Do not measure geometry with Playwright or DOM eval, do not adjust
pixel by pixel, and do not re-shoot for a copy tweak. Fix only what the screenshot
shows is wrong, in at most two rounds, then stop and ask.

For other web, app, desktop, mobile, or TUI layout requests, include an ASCII-art
mockup of the primary screen.

# Decisions

This section is always last, with a context-relevant title rather than a fixed phrase.
Number user-confirmable items **`1.` `2.` `3.` …**, not letters or unnumbered bullets.
For each open decision, mark one most sound default and give a one-line reason grounded
in project conventions, existing patterns, or smallest reversible risk; include only
real alternatives. If none remain open, begin the section with a **Locked decisions**
table or a title that clearly means locked.
