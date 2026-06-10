# Runner Integration

This document describes production installation for Codex and opencode runner
hooks. Start Agent Hub before relying on production hooks:

```sh
agent-hub daemon start
```

The hook commands below publish native runner payloads to:

```sh
agent-hub hook notify --runner <runner> --event <event>
```

## Codex Hooks

Codex hooks can call `agent-hub hook notify` for lifecycle and tool events.
Codex hook trust must be reviewed in Codex before non-managed hooks run. Use
Codex's hook review flow after adding or changing hook definitions.

Example `.codex/hooks.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|resume",
        "hooks": [
          {
            "type": "command",
            "command": "agent-hub hook notify --runner codex --event SessionStart"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "agent-hub hook notify --runner codex --event UserPromptSubmit"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "agent-hub hook notify --runner codex --event Stop"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "agent-hub hook notify --runner codex --event PreToolUse"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "agent-hub hook notify --runner codex --event PostToolUse"
          }
        ]
      }
    ]
  }
}
```

Recommended Codex event mapping:

| Codex event | Agent Hub event |
| --- | --- |
| `SessionStart` | `agent.session.started` |
| `UserPromptSubmit` | `agent.prompt.submitted` |
| `Stop` | `agent.session.finished` |
| `PreToolUse` | `agent.tool.started` |
| `PostToolUse` | `agent.tool.finished` |
| `PermissionRequest` | `agent.permission.requested` |

## opencode Plugin

Install the plugin locally under `.opencode/plugins/agent-hub.ts` or globally
under the opencode user config plugin directory.

```ts
export const AgentHubPlugin = async ({ $ }) => {
  const notify = async (event: string, payload: unknown) => {
    await $`agent-hub hook notify --runner opencode --event ${event}`.stdin(
      JSON.stringify(payload),
    )
  }

  return {
    event: async ({ event, properties }) => {
      switch (event.type) {
        case "session.created":
          await notify("session.created", properties)
          break
        case "message.updated":
          await notify("message.updated", properties)
          break
        case "session.idle":
          await notify("session.idle", properties)
          break
        case "session.error":
          await notify("session.error", properties)
          break
        case "tool.execute.before":
          await notify("tool.execute.before", properties)
          break
        case "tool.execute.after":
          await notify("tool.execute.after", properties)
          break
      }
    },
  }
}
```

Recommended opencode event mapping:

| opencode event | Agent Hub event |
| --- | --- |
| `session.created` | `agent.session.started` |
| `message.updated` | `agent.prompt.submitted` |
| `session.idle` | `agent.session.finished` |
| `session.error` | `agent.session.failed` |
| `tool.execute.before` | `agent.tool.started` |
| `tool.execute.after` | `agent.tool.finished` |

## Test Boundary

Automated tests use fake runners. They should use `cmd/fake-codex` and
`cmd/fake-opencode` with `--mock-config`, temporary `AGENT_HUB_HOME`, temporary
`CODEX_HOME`, and temporary `OPENCODE_CONFIG_DIR`. Automated tests must not
write to `~/.codex`, `~/.config/opencode`, or `~/.agent-hub`.

