# Scenario

**Feature**: ensureServer auto-starts a detached crush daemon when none is running

## Preconditions
- No crush server is currently running.
- `ensureServer` should detect the absence, start the server as a detached daemon, and wait up to 10s for health.

## Steps
1. Set `ServerOperation` to `"auto-start"` to test automatic server startup.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ServerOperation = "auto-start"
	return nil
}
```
