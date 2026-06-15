## Preconditions
- These tests require a running crush server.
- They are gated by `CRUSH_INTEGRATION_TEST=1` — leaves skip if not set.
- The `crush` binary must be in PATH.

## Steps
1. Set `req.Mode = "server-client"`.
2. Set `req.ServerOperation` to the desired operation.
3. Root `Run` creates a `CrushServerClient`, ensures server is running, executes the operation, and returns results.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if os.Getenv("CRUSH_INTEGRATION_TEST") != "1" {
		t.Skip("CRUSH_INTEGRATION_TEST not set; skip integration test")
		return nil
	}
	req.Mode = "server-client"
	return nil
}
```
