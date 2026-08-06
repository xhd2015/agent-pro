# Verify on behalf of user — Usage Reference

## When to use

Run human-style **scenario** verification after an agent claims work is done and
automated tests passed. Mimics what the user would do before commit: `git status`,
build, install to a safe PATH (or opt-in real GOBIN in host mode), stand up real
processes when needed, run user journeys (CLI and/or **browser-agent** for UI),
write a markdown transcript, and **inline the full transcript** in the reply.

## When NOT to use

- Automated regression → `doctest test` (headless tools belong there)
- Design clarification before coding → `followup` / `brainstorm`
- Risky host integration POC → `verify-with-prototype`
- User said `no verify:`

## Modes

| Mode | Selection | Install | Env |
|------|-----------|---------|-----|
| **sandbox** | **Default** always | `go build -o ~/.sandbox/bin/...` after `enter-sandbox.sh` | `HOME=~/.sandbox/default-home` |
| **host** | Only explicit opt-in: "host", "outside sandbox", clear synonyms | Change-scoped ladder: `wrk --reinstall-local` → `script`/`scripts` install → `go install ./cmd/<T>` | Real HOME / GOBIN; warn + dry-run/plan first |

Ambiguous language → **sandbox**. See topic `host`.

## Install and show

```sh
agent-pro skill verify-on-behalf-of-user --show
agent-pro skill verify-on-behalf-of-user workflow --show
agent-pro skill --show verify-on-behalf-of-user/sandbox
agent-pro skill --show verify-on-behalf-of-user/host
agent-pro skill --show verify-on-behalf-of-user/scenario
agent-pro skill --show verify-on-behalf-of-user/tty
agent-pro skill --show verify-on-behalf-of-user/transcript
agent-pro skill --install verify-on-behalf-of-user
agent-pro skills update
```

Installs to `.agents/skills/verify-on-behalf-of-user/`:

- `SKILL.md` (index)
- `sandbox/TOPIC.md`, `host/TOPIC.md`, `workflow/TOPIC.md`, `scenario/TOPIC.md`, `tty/TOPIC.md`, `transcript/TOPIC.md`
- `scripts/enter-sandbox.sh`
- `scripts/sandbox-verify.sb`
- `templates/transcript.md`

## Sandbox paths

| Path | Role |
|------|------|
| `~/.sandbox/default-home` | `HOME` during sandbox-mode verify |
| `~/.sandbox/bin` | Built binaries (sandbox mode) |
| `~/.sandbox/transcripts/` | Output transcripts (both modes) |

```sh
source .agents/skills/verify-on-behalf-of-user/scripts/enter-sandbox.sh
```

## Primary deliverable

1. Markdown terminal transcript at `~/.sandbox/transcripts/<timestamp>-<slug>.md`
   with `$ command` blocks, verbatim output, **Mode**, depth/surface labels, and checks.
2. **Full transcript body inlined** in the agent reply (same content as the file) so
   the user can review without opening the file.
3. One-line path + verdict + mode at the top of the reply.

## Non-negotiable rules

| Rule | Detail |
|------|--------|
| Default Mode sandbox | Host only on explicit opt-in ("host" / "outside sandbox" / clear synonyms) |
| Always label depth | `smoke` \| `scenario` \| `full` + reason in transcript and reply |
| Scenario-first | Real bring-up + user journey by default; smoke only as labeled downgrade |
| UI = browser-agent | Never playwright-debug / headless for this skill |
| UI path broken | **FAIL** (not skip) |
| Interactive TTY | **tty-watch** `run --detach` → `send`/`snapshot` → always **`kill`** (see `tty` topic) |
| Transcript | Write file **and** print full contents; include **Mode** |
| Host safety | Warnings; dry-run/plan before mutate; change-scoped targets only |

## Pairing

1. `doctest-tdd` → sealed tests GREEN
2. **verify-on-behalf-of-user** → scenarios + transcript (file + inline); default sandbox
3. User reviews inlined transcript and commits

## Project recipes

Add `docs/verify-recipes/<feature>.md` in each consumer repo (not in agent-pro).
See `references/README.md` for the scenario-oriented schema.
