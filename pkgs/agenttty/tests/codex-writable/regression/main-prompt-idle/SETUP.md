# Scenario

**Feature**: main chat prompt after boot is idle even when MCP startup incomplete

```
snapshot has "MCP startup incomplete" + main chat › + model ready
  -> CheckWritable returns ready=true, state=idle
```

## Preconditions

- Fixture `codex-mcp-incomplete-prompt.txt` from live capture
  (`/tmp/codex-status-fixtures-for-req/mcp-incomplete-idle-prompt.txt`).
- Existing rule: MCP incomplete is loading only when `›` is absent.

## Steps

1. Set `req.FixtureFile` to the MCP incomplete + main prompt fixture.

## Context

- F4 ensures `/status` can still be sent after real boot completes (not over-blocked by MCP warnings).
- Alternate capture `codex-main-prompt-mcp-incomplete.txt` is covered by the fixture table.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FixtureFile = fixtureMainPromptMCP
	return nil
}
```
