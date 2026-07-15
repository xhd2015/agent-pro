---
name: verify-on-behalf-of-user
description: >-
  Run human-style smoke verification after an agent claims work is done: git
  status, build into ~/.sandbox/bin, sandbox HOME=~/.sandbox/default-home, smoke
  commands, markdown terminal transcript. Use for "verify on my behalf", before
  commit. Complements doctests; does not replace them. Load a sub-topic with:
  agent-pro skill --show verify-on-behalf-of-user/<topic>
---

# Verify on behalf of user

After an agent claims work is done and automated tests passed, run the **same steps
a human would** before committing. Write everything to a **markdown transcript**
the user can review like a terminal session.

This is **outside doctests** — it mimics one-time manual verification.

## When to use

- Agent finished a feature/fix and reported GREEN doctests
- User would otherwise run `git status`, `go build`, `go install`, and ad-hoc CLI commands
- Before user checks in or commits

## When NOT to use

- Deep adversarial code review → use project-specific review skills
- Automated regression → use `doctest test`
- User said `no verify:` or only wants a diff review

## Topics

- `sandbox` — `~/.sandbox` paths, `enter-sandbox.sh`, forbidden writes, strict mode
- `workflow` — six-phase verify workflow (scope → git → build → smoke → logs → transcript)
- `transcript` — markdown format rules, template, quick start

## Usage

```bash
# root index (this file)
agent-pro skill verify-on-behalf-of-user --show

# nested topics (both flag orders)
agent-pro skill --show verify-on-behalf-of-user sandbox
agent-pro skill verify-on-behalf-of-user workflow --show
agent-pro skill --show verify-on-behalf-of-user/transcript
```

Utility files install with the skill (not in topic bodies):

- `scripts/enter-sandbox.sh`, `scripts/sandbox-verify.sb`
- `templates/transcript.md`

## Relationship to other skills

| Skill | Role |
|-------|------|
| doctest-tdd | Automated sealed tests |
| verify-with-prototype | POC before risky host integration |
| **verify-on-behalf-of-user** | Human-shaped sandbox smoke + transcript |

Suggested order: doctests GREEN → this skill → user commits.

## Project recipes

Consumer repos add `docs/verify-recipes/<feature>.md` with build commands,
smoke sequence, and expected outputs. See agent-pro `references/README.md`.