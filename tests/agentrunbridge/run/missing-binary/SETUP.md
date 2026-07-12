# Scenario

**Feature**: LookPath failure surfaces binary-not-found error

```
LookPath fails -> Run error; LaunchCalls==0
```

## Preconditions

- Valid non-empty prompt.
- `LookPathFail=true`.

## Steps

1. Set prompt/session; enable LookPathFail.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "need binary"
	req.SessionID = "sess-missing-bin"
	req.KeepTTY = true
	req.LookPathFail = true
	req.Binary = "agent-run-missing"
	return nil
}
```
