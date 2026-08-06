---
name: verify-on-behalf-of-user
description: >-
  Run human-style scenario verification after an agent claims work is done: git
  status, build into ~/.sandbox/bin (default sandbox mode) or opt-in host mode
  (outside sandbox: wrk --reinstall-local / change-scoped install), real
  integration/E2E scenarios (start servers, browser-agent for UI), always-labeled
  depth, markdown transcript written and inlined for review. Use for "verify on
  my behalf", before commit. Complements doctests; does not replace them. Load a
  sub-topic with: agent-pro skill --show verify-on-behalf-of-user/<topic>
---

# Verify on behalf of user

After an agent claims work is done and automated tests passed, run the **same steps
a human would** before committing — real user journeys, not only smoke CLI checks.
Write a **markdown transcript** and **print its full contents** in the reply so the
user can review without opening the file.

This is **outside doctests** — it mimics one-time manual verification.

## Modes

| Mode id | When | Install / env |
|---------|------|----------------|
| **`sandbox`** (default) | Always, unless the user clearly opts into host | `enter-sandbox.sh`; `~/.sandbox/bin`; `HOME=~/.sandbox/default-home` |
| **`host`** (opt-in) | Only when user explicitly says host / outside sandbox / clear synonyms | Real HOME/PATH; change-scoped install ladder (`wrk` → scripts → `go install`); dry-run/plan first; warnings required |

See topic **`host`** for opt-in phrases, install ladder, and change-scoped targets.
If language is ambiguous → stay **sandbox**.

## When to use

- Agent finished a feature/fix and reported GREEN doctests
- User would otherwise start servers, exercise the product, and open the UI
- Before user checks in or commits

## When NOT to use

- Deep adversarial code review → use project-specific review skills
- Automated regression → use `doctest test` (headless tools like playwright-debug belong there)
- User said `no verify:` or only wants a diff review

## Topics

- `sandbox` — default mode: `~/.sandbox` paths, `enter-sandbox.sh`, forbidden writes, strict mode
- `host` — opt-in outside-sandbox mode: real env, `wrk --reinstall-local` / fallbacks, dry-run, warnings
- `workflow` — phases: scope → git → build → runtime → scenarios → evidence → transcript (mode-branched)
- `scenario` — depth labels, surface gate, browser-agent UI rules, FAIL when UI path broken
- `tty` — interactive CLI via **tty-watch** (`run --detach`, `send`, `snapshot`, always `kill`)
- `transcript` — format rules, template, write file **and** inline full content

## Usage

```bash
# root index (this file)
agent-pro skill verify-on-behalf-of-user --show

# nested topics (both flag orders)
agent-pro skill --show verify-on-behalf-of-user sandbox
agent-pro skill --show verify-on-behalf-of-user host
agent-pro skill verify-on-behalf-of-user workflow --show
agent-pro skill --show verify-on-behalf-of-user/scenario
agent-pro skill --show verify-on-behalf-of-user/tty
agent-pro skill --show verify-on-behalf-of-user/transcript
```

Utility files install with the skill (not in topic bodies):

- `scripts/enter-sandbox.sh`, `scripts/sandbox-verify.sb`
- `templates/transcript.md`

## Core rules (non-negotiable)

1. **Always label depth** — every transcript and final reply must include
   `Depth: smoke | scenario | full` plus a one-line reason. Never leave depth unlabeled.
2. **Default mode is sandbox** — use Mode `host` only on explicit user opt-in
   ("host", "outside sandbox", or clear synonyms). Otherwise always sandboxed.
   See topic `host`.
3. **Scenario-first** — default is real integration (bring-up + user journey). Smoke is a
   **labeled downgrade** when justified (e.g. pure CLI with no runtime surface).
4. **UI uses browser-agent only** — for any UI/frontend surface, drive a live browser via
   `browser-agent`. Do **not** use `playwright-debug` or other headless runners here;
   those are for doctests/CI.
5. **Missing UI path = FAIL** — if the surface requires UI and browser-agent/Chrome/session
   is unusable, verdict is **FAIL** (not skip, not BLOCKED).
6. **Interactive TTY uses tty-watch** — for claims that need a real TTY, drive the CLI with
   `tty-watch run --detach` (non-blocking), then `send` / `snapshot`, and always
   `tty-watch kill` to reclaim. Do not use pipe-only or raw openpty harnesses as sole
   evidence. Missing `tty-watch` when TTY is required → **FAIL**. See topic `tty`.
7. **Transcript = file + inline** — write `~/.sandbox/transcripts/<timestamp>-<slug>.md`,
   then put the **full file body** in the agent reply for direct review. Header must
   include **Mode** (`sandbox` \| `host`).
8. **Host mode safety** — when Mode is `host`: warn the user; prefer `--dry-run` or a
   written plan before mutating ops; install only change-scoped targets (see `host`).

## Relationship to other skills

| Skill | Role |
|-------|------|
| doctest-tdd | Automated sealed tests (may use headless browsers) |
| verify-with-prototype | POC before risky host integration |
| **verify-on-behalf-of-user** | Human-shaped **scenarios** (default sandbox; opt-in host) + transcript (file + inline) |

Suggested order: doctests GREEN → this skill → user commits.

## Project recipes

Consumer repos add `docs/verify-recipes/<feature>.md` with surface, depth, bring-up,
scenarios, teardown, and expected observables. See `references/README.md`.
