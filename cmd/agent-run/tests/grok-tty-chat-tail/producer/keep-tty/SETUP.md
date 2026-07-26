# Scenario

**Bug**: `--keep-tty` producer must not cancel grok tail when `tailState.streamed` is set

```
agent-run run --agent-runner grok-tty --keep-tty
  -> llm-mock chrome holds PTY
  -> pre-seeded pending tool_call in updates.jsonl
  -> scheduled completion append while TTY alive
```

## Preconditions

- Mirrors web `KeepTerminalAlive` path: `--keep-tty` + `--agent-runner-binary llm-mock-run-grok`.
- `GROK_HOME` pre-seeded with partial turn; completion lines appended on schedule.

## Steps

1. Grouping setup calls `configureProducerKeepTTYEnv` and wires chrome hold seconds.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.ChromeHoldSeconds <= 0 {
		req.ChromeHoldSeconds = defaultChromeHoldSec
	}
	configureProducerKeepTTYEnv(t, req)
	return nil
}
```