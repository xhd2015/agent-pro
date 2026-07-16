# Verify on behalf of user — Usage Reference

## When to use

Run human-style **scenario** verification after an agent claims work is done and
automated tests passed. Mimics what the user would do before commit: `git status`,
build, install to a safe PATH, stand up real processes when needed, run user journeys
(CLI and/or **browser-agent** for UI), write a markdown transcript, and **inline the
full transcript** in the reply.

## When NOT to use

- Automated regression → `doctest test` (headless tools belong there)
- Design clarification before coding → `followup` / `brainstorm`
- Risky host integration POC → `verify-with-prototype`
- User said `no verify:`

## Install and show

```sh
agent-pro skill verify-on-behalf-of-user --show
agent-pro skill verify-on-behalf-of-user workflow --show
agent-pro skill --show verify-on-behalf-of-user/sandbox
agent-pro skill --show verify-on-behalf-of-user/scenario
agent-pro skill --show verify-on-behalf-of-user/transcript
agent-pro skill --install verify-on-behalf-of-user
agent-pro skills update
```

Installs to `.agents/skills/verify-on-behalf-of-user/`:

- `SKILL.md` (index)
- `sandbox/TOPIC.md`, `workflow/TOPIC.md`, `scenario/TOPIC.md`, `transcript/TOPIC.md`
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

1. Markdown terminal transcript at `~/.sandbox/transcripts/<timestamp>-<slug>.md`
   with `$ command` blocks, verbatim output, depth/surface labels, and checks.
2. **Full transcript body inlined** in the agent reply (same content as the file) so
   the user can review without opening the file.
3. One-line path + verdict at the top of the reply.

## Non-negotiable rules

| Rule | Detail |
|------|--------|
| Always label depth | `smoke` \| `scenario` \| `full` + reason in transcript and reply |
| Scenario-first | Real bring-up + user journey by default; smoke only as labeled downgrade |
| UI = browser-agent | Never playwright-debug / headless for this skill |
| UI path broken | **FAIL** (not skip) |
| Transcript | Write file **and** print full contents |

## Pairing

1. `doctest-tdd` → sealed tests GREEN
2. **verify-on-behalf-of-user** → sandbox scenarios + transcript (file + inline)
3. User reviews inlined transcript and commits

## Project recipes

Add `docs/verify-recipes/<feature>.md` in each consumer repo (not in agent-pro).
See `references/README.md` for the scenario-oriented schema.
