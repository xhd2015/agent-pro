## Expected

- Exit code 0 (idempotent prune).
- Stale registry file removed.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("kill stale exit %d combined %q", resp.ExitCode, resp.Combined)
	}
	if resp.RegistryExists {
		t.Fatalf("stale registry %s should be pruned", resp.SessionID)
	}
}
```