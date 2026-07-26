# Scenario

**Feature**: crush daemon process lifecycle leaves under the server-client subtree

## Preconditions
- Parent `server/server-client/` already sets `Mode="server-client"` and gates on crush binary availability.
- These leaves verify the crush daemon process lifecycle: start, health-check, kill, confirm stopped.

## Steps
1. Child leaves set `ServerOperation` to `"server-lifecycle"`.
2. Root `Run` calls `runServerClient` with the lifecycle operation.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ServerOperation = "server-lifecycle"
	return nil
}
```
