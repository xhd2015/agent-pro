# Scenario

**Feature**: missing session dispatches RunSession only (no binary)

```
AutoSendOrResume(missing session, hooks)
  -> Classify ModeRun
  -> RunSession x1; Send=0; Resume=0
```

## Preconditions

- Session not seeded.
- Hooks installed; prove no agent-run binary is required for MODE=run unit path.

## Steps

1. Unknown SessionID; no probe needed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-auto-run-new"
	req.SeedMeta = false
	req.UseProbe = false
	req.Prompt = "create me"
	return nil
}
```
