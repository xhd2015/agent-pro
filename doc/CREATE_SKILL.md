# Creating agent-pro Skills

This document describes how to add a new skill to **agent-pro**. The default
and preferred shape is **skill-only** — an embedded `SKILL.md` prompt that
agents load via `agent-pro skill --show <name>` or `agent-pro skill --install <name>`.

## Scope

Skills created through this document are **always agent-pro skills**: embedded
under `agents/<name>/`, registered in `cmd/agent-pro/skill_cmd.go`, and
installed via `agent-pro skill --install <name>` (typically to `.agents/skills/`).

Do **not** use Grok's `~/.grok/skills/` or project `.grok/skills/` for
agent-pro workflow skills unless the user explicitly asks for a Grok-only skill.
Grok's `create-skill` flow is a separate path; agent-pro skill work follows
this document.

Use a **sub-agent CLI** (separate binary with its own LLM runner) only when the
skill needs persistent sessions, model calls, or a standalone executable.

## Skill-only (preferred)

**When to use**

- Workflow is a prompt/guideline the parent agent follows inline
- No dedicated session store or `agent-pro`-external binary required
- Examples: `investigate`, `followup`, `brainstorm`, `verify-with-prototype`

**Layout**

```
agents/<skill-name>/
└── run/
    ├── SKILL.md                 # YAML frontmatter + markdown body
    └── <skill_name>.go          # //go:embed SKILL.md → SkillFile string
```

**`SKILL.md` frontmatter** (required):

```yaml
---
name: my-skill
description: >-
  One-line summary for agent-pro skills list and auto-invocation triggers.
---
```

**Go embed** (`agents/my-skill/run/my_skill.go`):

```go
package run

import _ "embed"

//go:embed SKILL.md
var SkillFile string
```

**Register in agent-pro** (`cmd/agent-pro/skill_cmd.go`):

1. Import `github.com/xhd2015/agent-pro/agents/my-skill/run`
2. Add entry to `knownSkills` using `extractDescription(run.SkillFile)` and `run.SkillFile`
3. Append `"my-skill"` to `knownSkillNames()` and `knownSkillNamesText`
4. Update help strings (`skillHelp`, `skillsHelp`) skill name lists

**Tests**

- `cmd/agent-pro/skill_cmd_test.go` — `TestMySkillIsRegistered`
- Optional doctest tree: `agents/my-skill/tests/agent-pro/` (register + list leaves)

**Install / show (no separate binary)**

```sh
agent-pro skill --show my-skill
agent-pro skill my-skill --show               # both flag orders work
agent-pro skill --install my-skill            # → .agents/skills/my-skill/
agent-pro skill my-skill --install --global
agent-pro skill --install my-skill --cursor
agent-pro skills update                       # refresh installed copies
```

## Sub-agent CLI (optional)

**When to use**

- Skill needs its own LLM session loop (`run.RunArgs`, catch-up, list-sessions)
- Users install a standalone binary (`go install …/agents/foo@latest`)

**Layout**

```
agents/<skill-name>/
├── main.go              # skill --show | skill --install + agent runner dispatch
├── main_test.go
└── run/
    ├── SKILL.md
    ├── <skill>.go       # agent implementation
    └── ...
```

Follow `go-best-practice skill-cli` for `skill --show` / `skill --install` in
`main.go`. Still register in `cmd/agent-pro/skill_cmd.go` so `agent-pro skills`
lists the skill and `agent-pro skill --install <name>` works without the binary.

Examples with sub-agent binaries: `reproduce`.

Examples skill-only: `investigate`, `followup`, `brainstorm`,
`verify-with-prototype`.

## Checklist for a new skill-only skill

1. Create `agents/<name>/run/SKILL.md` with clear workflow and triggers in `description`
2. Add `//go:embed` Go file exporting `SkillFile`
3. Register in `cmd/agent-pro/skill_cmd.go`
4. Add `Test<Name>SkillIsRegistered` in `skill_cmd_test.go`
5. (Recommended) Add `agents/<name>/tests/agent-pro/` doctests for register + list
6. Run `go test ./cmd/agent-pro/ -run Test<Name>Skill`

## Skill body guidelines

- Write for the **parent agent** — imperative steps, decision points, response shape
- State **when to use** and **when not to** (pair with other skills if needed)
- Prefer **rerunnable verification** (doctests, unit tests, manual matrices) for Go work
- Keep under ~200 lines; move long case studies to `topics/` only if embedded in a sub-agent binary

## Related skills

| Skill | Role |
|-------|------|
| `investigate` | Explain why (symptom → cause) |
| `followup` | Clarify design before coding |
| `verify-with-prototype` | POC script before full implementation |
| `brainstorm` | Requirement + test plan approval |
| `consolidate-code` | Map duplication → PR stack → shared packages |
| `establish-a-loop` | Document and dry-run a build/deploy/run/inspect/fix loop (`LOOP_<date>_<slug>.md`) |
| `run-the-loop` | Goal-driven iterate-until-done with inspect script |
| `sound-fix` | Evaluate fix soundness vs workarounds |
| `summarize-a-skill` | Distill session tries into a brief skill SKILL.md |
| `doctest-tdd` | Sealed tests → implementation |