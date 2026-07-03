---
name: debug-with-user
description: >-
  Human-in-the-loop macOS debugging checkpoints. Use when automation needs a
  human to confirm UI state, report observations, or proceed through manual steps.
---

# debug-with-user

Human-assisted debugging for macOS workflows. Automate everything scriptable;
at checkpoints, ask the user via native dialogs and parse structured JSON answers.

## When to use

- `/debug-with-user` or explicit human-assisted debug requests
- macOS UI steps that need confirmation (window opened, correct app focused, etc.)
- After automation completes a step and you need a human report

## Workflow

1. **Automate first** — run every scriptable step (open apps, post notifications, etc.).
2. **Checkpoint** — run `debug-with-user ask` with a clear title, message, and absolute paths.
3. **Parse JSON** — read the single-line stdout object:
   - `via == "button"`: use `affirmed` (true/false) for pass/fail semantics.
   - `via == "free_text"`: treat `answer` as the user's report (Customize path).
4. **Exit codes** — `0` answer, `1` dismissed/cancel, `2` error (e.g. non-macOS without dry-run).

## CLI examples

```bash
debug-with-user ask \
  --title "Step 1 — Did VS Code open?" \
  --message "Project folder:\n/tmp/demo" \
  --option "Yes — window opened" \
  --option "No — window did not open" \
  --affirm "Yes — window opened" \
  --cancel "Cancel"
```

The CLI always appends a **Customize** button for free-text reports.

Aliases: `confirm` (`--yes` / `--no`), `proceed` (`--proceed` / `--cancel`).

## Install

If `debug-with-user` is not on PATH:

```bash
go install github.com/xhd2015/agent-pro/agents/debug-with-user@latest
```

## CI / non-macOS

Real dialogs require macOS `osascript`. In CI, set `DEBUG_WITH_USER_DRY_RUN=1` and stage
responses with `DEBUG_WITH_USER_DRY_RUN_BUTTON`, `DEBUG_WITH_USER_DRY_RUN_TEXT`, or
`DEBUG_WITH_USER_DRY_RUN_DISMISSED=1`.

## JSON contract

Preset button:

```json
{"answer":"Yes — window opened","via":"button","affirmed":true}
```

Customize (free text):

```json
{"answer":"VS Code opened but wrong workspace","via":"free_text"}
```