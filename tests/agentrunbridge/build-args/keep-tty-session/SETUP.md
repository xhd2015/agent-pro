# Scenario

**Feature**: session run with KeepTTY and no Open

```
BuildArgs(session + KeepTTY, Open=false)
  -> run --keep-tty --session-id=… …  # no --open
```

## Preconditions

- Non-empty session id; `KeepTTY=true`; `Open=false`.
- CaptureStdout is irrelevant to argv (not asserted here).

## Steps

1. Set session, keep-tty, optional runner, prompt; leave Open false.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-keep"
	req.Prompt = "keep tty prompt"
	req.AgentRunner = "fake-opencode"
	req.KeepTTY = true
	req.Open = false
	return nil
}
```
