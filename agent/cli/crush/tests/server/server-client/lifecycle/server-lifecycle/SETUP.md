# Scenario

**Feature**: crush daemon lifecycle start, health-check, kill, and confirm stopped

## Preconditions
- A crush server should NOT be running before the test.
- After `EnsureServer`, the server should be running.
- After manual kill, the server process should be gone and health check should fail.

## Steps
1. Set `ServerOperation` to `"server-lifecycle"`.
2. Root `Run` executes the lifecycle operation and returns a JSON result with process counts and health status at each stage.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ServerOperation = "server-lifecycle"
	return nil
}
```
