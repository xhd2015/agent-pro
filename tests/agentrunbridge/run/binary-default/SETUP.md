# Scenario

**Feature**: empty Binary defaults to `agent-run` for LookPath

```
Run(Binary="") -> LookPath("agent-run") -> launch ok
```

## Preconditions

- Binary field empty string.
- LookPath succeeds; no wait-ready.

## Steps

1. Set prompt/session; Binary empty; WaitReady false.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "default binary"
	req.SessionID = "sess-bin-default"
	req.Binary = ""
	req.KeepTTY = true
	req.WaitReady = false
	return nil
}
```
