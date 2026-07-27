# Scenario

**Feature**: scrollback capture is sufficient when no structured sidecar stream exists

```
fake TUI emits Response: hi
  -> scrollback capture emits hi to stdout
  -> sessions/codex-tty/.../events.jsonl records hi
```

## Steps

1. Run with respond fake TUI and prompt `hi`.
2. Assert scrollback capture on stdout/events without relying on a structured sidecar.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SkipCodexSessionDir = true
	req.CodexTTYCommand = fakeTUIRespondHi()
	req.Args = append(req.Args, "hi")
	return nil
}
```
