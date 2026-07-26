# Scenario

**Feature**: `--open` stays fully silent (no event stream / discovery think) until after attach

```
agent-run run --agent-runner grok-tty --open "hi"
  -> stdout/stderr have no "Resolve session id", 💭, 💬, NDJSON events
  -> session id line appears only as the post-attach single line (if at all)
```

## Steps

1. Use grouping open args with prompt `hi`.
2. Assert combined output has no forbidden noise.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hi"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", req.Prompt}
	setGrokTTYCommand(req, fakeTUIRespondHi())
	return nil
}
```
