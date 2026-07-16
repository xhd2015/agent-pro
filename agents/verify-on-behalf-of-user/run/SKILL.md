---
name: verify-on-behalf-of-user
description: >-
  Run human-style scenario verification after an agent claims work is done: git
  status, build into ~/.sandbox/bin, sandbox HOME=~/.sandbox/default-home, real
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

## When to use

- Agent finished a feature/fix and reported GREEN doctests
- User would otherwise start servers, exercise the product, and open the UI
- Before user checks in or commits

## When NOT to use

- Deep adversarial code review → use project-specific review skills
- Automated regression → use `doctest test` (headless tools like playwright-debug belong there)
- User said `no verify:` or only wants a diff review

## Topics

- `sandbox` — `~/.sandbox` paths, `enter-sandbox.sh`, forbidden writes, strict mode
- `workflow` — phases: scope → git → build → runtime → scenarios → evidence → transcript
- `scenario` — depth labels, surface gate, browser-agent UI rules, FAIL when UI path broken
- `transcript` — format rules, template, write file **and** inline full content

## Usage

```bash
# root index (this file)
agent-pro skill verify-on-behalf-of-user --show

# nested topics (both flag orders)
agent-pro skill --show verify-on-behalf-of-user sandbox
agent-pro skill verify-on-behalf-of-user workflow --show
agent-pro skill --show verify-on-behalf-of-user/scenario
agent-pro skill --show verify-on-behalf-of-user/transcript
```

Utility files install with the skill (not in topic bodies):

- `scripts/enter-sandbox.sh`, `scripts/sandbox-verify.sb`
- `templates/transcript.md`

## Core rules (non-negotiable)

1. **Always label depth** — every transcript and final reply must include
   `Depth: smoke | scenario | full` plus a one-line reason. Never leave depth unlabeled.
2. **Scenario-first** — default is real integration (bring-up + user journey). Smoke is a
   **labeled downgrade** when justified (e.g. pure CLI with no runtime surface).
3. **UI uses browser-agent only** — for any UI/frontend surface, drive a live browser via
   `browser-agent`. Do **not** use `playwright-debug` or other headless runners here;
   those are for doctests/CI.
4. **Missing UI path = FAIL** — if the surface requires UI and browser-agent/Chrome/session
   is unusable, verdict is **FAIL** (not skip, not BLOCKED).
5. **Transcript = file + inline** — write `~/.sandbox/transcripts/<timestamp>-<slug>.md`,
   then put the **full file body** in the agent reply for direct review.

## Relationship to other skills

| Skill | Role |
|-------|------|
| doctest-tdd | Automated sealed tests (may use headless browsers) |
| verify-with-prototype | POC before risky host integration |
| **verify-on-behalf-of-user** | Human-shaped sandbox **scenarios** + transcript (file + inline) |

Suggested order: doctests GREEN → this skill → user commits.

## Project recipes

Consumer repos add `docs/verify-recipes/<feature>.md` with surface, depth, bring-up,
scenarios, teardown, and expected observables. See `references/README.md`.
