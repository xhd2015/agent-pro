# Scenario

**Feature**: two CrushServerClient instances share a single crush daemon process

## Preconditions
- The crush server should not be running before the test.
- Two separate `CrushServerClient` instances (A and B) are created.
- Only one `crush server` OS process should exist after both call `EnsureServer`.

## Steps
1. Set `ServerOperation` to `"server-reuse"`.
2. Root `Run` creates two clients, ensures server via A, then ensures server via B, and returns process counts and health status.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ServerOperation = "server-reuse"
	return nil
}
```
