# Scenario

**Feature**: parser finds status fields amid MCP warnings and tips

```
# fake TUI prints noise before status box -> stdout still three canonical lines
codex-show-status -> fake codex (MCP noise + status) -> parse -> print
```

## Preconditions

- Fake TUI prints MCP boot warnings and tips **before** the monthly credit limit line.

## Steps

1. Set `ShowStatusCommand` to `fakeTUIExtraNoise()`.
2. Run and assert stdout matches the default canonical three lines.

## Context

- Mirrors real Codex TUI scrollback where `/status` output is buried under startup noise.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ShowStatusCommand = fakeTUIExtraNoise()
	return nil
}
```