# Scenario

**Feature**: prompt injection waits for `GROK_TTY_BANNER` before writing to PTY

```
fake TUI delays banner 300ms → run still succeeds (no banner timeout error)
```

## Preconditions

- If run injected prompt before banner appeared, fake TUI `read` would miss input and
  the run would fail or capture empty output.

## Steps

1. Use delayed-banner fake TUI script.
2. Run with prompt `hi`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokTTYCommand = fakeTUIDelayedBanner()
	req.Args = append(req.Args, "hi")
	return nil
}
```