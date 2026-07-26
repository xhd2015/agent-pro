# Scenario

**Feature**: failed grok bind emits error event instead of scrollback assistant message

```
empty GROK_HOME + PTY chrome hook (no updates.jsonl)
  -> DiscoverSession fails
  -> events.jsonl: think then error; no assistant fallback
```

## Steps

1. Configure failure binding env (empty `GROK_HOME`, hook prints PTY chrome only).
2. Start web and POST `grok-tty` session.
3. `Run` waits for `finished` and reads `events.jsonl`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "binding-failure-emits-error"
	configureBindingFailureEnv(t, req, "bind failure probe")
	startWebGrokSession(t, req)
	return nil
}
```