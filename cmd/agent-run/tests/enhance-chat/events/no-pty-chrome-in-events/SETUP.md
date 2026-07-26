# Scenario

**Feature**: PTY chrome from TUI never appears in events.jsonl

```
failure hook prints box-drawing/banner strings to PTY
  -> events.jsonl must not contain those substrings
```

## Steps

1. Configure failure binding env (PTY chrome hook, no grok session dir).
2. Start web and POST `grok-tty` session.
3. `Run` waits for `finished` and reads `events.jsonl`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "no-pty-chrome-in-events"
	configureBindingFailureEnv(t, req, "pty chrome isolation probe")
	startWebGrokSession(t, req)
	return nil
}
```