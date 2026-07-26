# Scenario

**Bug**: `--keep-tty` web path must not let PTY chrome end discovery before `updates.jsonl` streams

```
agent-run run --agent-runner grok-tty --keep-tty
  -> llm-mock-run-grok + LLM_MOCK_RUN_GROK_COMMAND chrome hook
  -> waitForPersistentTurnRemote uses extraComplete(streamed) for grok
  -> events.jsonl records bind progress and streamed assistant or late error
```

## Preconditions

- Same flags as web `KeepTerminalAlive`: `--keep-tty` on CLI `run`.
- `--agent-runner-binary` points at session-built `llm-mock-run-grok`.
- `GROK_HOME` comes from `--agent-runner-config-home` (empty dir at start for failure leaves).

## Steps

1. Grouping setup wires keep-tty CLI args and llm-mock env defaults.
2. Leaf setup chooses delayed session schedule vs empty-home chrome failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ChromeHoldSeconds = defaultChromeHoldSeconds
	return nil
}
```