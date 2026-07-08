# Scenario

**Bug**: known terminal_session_id must not trigger repeating /terminal polls

```
session detail exposes terminal_session_id + live registry
  -> chat page passive watch
  -> terminal GET bounded (<=1)
```

## Preconditions

- Session `meta.json` includes `terminal_session_id`.
- Mapped PTY registry is live at page load.

## Steps

1. Seed finished `grok-tty` session with terminal mapping and registry.
2. Open session page; passive network watch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "known-terminal-id"
	return nil
}
```