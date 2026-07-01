# Scenario

**Feature**: CrushServerClient operations gated on the crush binary being in PATH

## Preconditions
- These tests require a running crush server.
- The `crush` binary is auto-detected via `exec.LookPath` — leaves skip if not found.

## Steps
1. Set `req.Mode = "server-client"`.
2. Set `req.ServerOperation` to the desired operation.
3. Root `Run` creates a `CrushServerClient`, ensures server is running, executes the operation, and returns results.

```go
import (
	osexec "os/exec"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	_, err := osexec.LookPath("crush")
	if err != nil {
		t.Skip("crush not in PATH; skip integration test")
		return nil
	}
	req.Mode = "server-client"
	return nil
}
```
