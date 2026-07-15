# Verify on behalf of user — Usage Reference

## When to use

Run human-style smoke verification after an agent claims work is done and
automated tests passed. Mimics what the user would do before commit: `git status`,
build, install to a safe PATH, run smoke commands, write a markdown transcript.

## When NOT to use

- Automated regression → `doctest test`
- Design clarification before coding → `followup` / `brainstorm`
- Risky host integration POC → `verify-with-prototype`
- User said `no verify:`

## Install and show

```sh
agent-pro skill verify-on-behalf-of-user --show
agent-pro skill verify-on-behalf-of-user workflow --show
agent-pro skill --show verify-on-behalf-of-user/sandbox
agent-pro skill --install verify-on-behalf-of-user
agent-pro skills update
```

Installs to `.agents/skills/verify-on-behalf-of-user/`:

- `SKILL.md` (index)
- `sandbox/TOPIC.md`, `workflow/TOPIC.md`, `transcript/TOPIC.md`
- `scripts/enter-sandbox.sh`
- `scripts/sandbox-verify.sb`
- `templates/transcript.md`

## Sandbox paths

| Path | Role |
|------|------|
| `~/.sandbox/default-home` | `HOME` during verify |
| `~/.sandbox/bin` | Built binaries |
| `~/.sandbox/transcripts/` | Output transcripts |

```sh
source .agents/skills/verify-on-behalf-of-user/scripts/enter-sandbox.sh
```

## Primary deliverable

Markdown terminal transcript at `~/.sandbox/transcripts/<timestamp>-<slug>.md`
with `$ command` blocks, verbatim output, and `> **Annotation:**` notes.

## Pairing

1. `doctest-tdd` → sealed tests GREEN
2. **verify-on-behalf-of-user** → sandbox smoke + transcript
3. User reviews transcript and commits

## Project recipes

Add `docs/verify-recipes/<feature>.md` in each consumer repo (not in agent-pro).