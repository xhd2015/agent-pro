# Scenario

**Feature**: server-client reuse leaves verifying daemon sharing across clients

## Preconditions
- Parent `server/server-client/` already sets `Mode="server-client"` and gates on crush binary availability.
- These leaves verify that multiple `CrushServerClient` instances reuse the same daemon server.

## Steps
1. Child leaves set `ServerOperation` to `"server-reuse"`.
2. Root `Run` calls `runServerClient` with the reuse operation.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ServerOperation = "server-reuse"
	return nil
}
```
