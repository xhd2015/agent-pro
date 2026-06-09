# Usage

```sh
# explain "Something to me"
# and continue by repeating  explain "Something to me" "Followup" 
go install github.com/xhd2015/agent-pro/agents/explain@latest

# test case
go install github.com/xhd2015/agent-pro/agents/test-case-design-expert@latest

# brainstorm
go install github.com/xhd2015/agent-pro/agents/idea-expander@latest
```

## Skills

Prompt-backed agent binaries include an embedded `SKILL.md`.

```sh
idea-expander skill show
idea-expander skill install --codex
idea-expander skill install --cursor
idea-expander skill install --opencode
idea-expander skill install
```

Default install behavior:
- No target flag installs to `.agents/skills/<agent-name>`.
- `--codex` installs to `.codex/skills/<agent-name>`.
- `--cursor` installs to `.cursor/skills/<agent-name>`.
- `--opencode` installs to `.opencode/skills/<agent-name>`.

Agents that support skill commands:
- `idea-expander`
- `test-case-design-expert`
- `test-case-tree-design-expert`
- `tdd-expert`