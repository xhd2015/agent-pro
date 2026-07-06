# Scenario

**Bug**: `agent-run snapshot` on a live grok-tty llm-mock session after turn completion
shows only the bottom status bar (`Turn completed`, `Ctrl+.:shortcuts`) instead of the
grok TUI content that `tty-watch snapshot` captures (user prompt, menu, input area).

```
background agent-run run --agent-runner grok-tty --agent-runner-binary llm-mock-run-grok --keep-tty
  -> wait for streamed stdout
  -> agent-run snapshot <session-id>
```

Reproduces user report comparing:

- `agent-run snapshot session-N` after `llm-mock-run-grok` turn (status-bar-only)
- `tty-watch snapshot debug-snapshot` (full grok home/menu screen)

Leaf `grok-mock-run-post-turn` uses a fake grok TUI that replays the post-turn
PTY redraw observed with real `llm-mock-run-grok` (deterministic in CI).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_ = t
	return nil
}
```