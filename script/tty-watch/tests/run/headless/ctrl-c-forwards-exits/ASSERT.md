## Expected

- Harness sends SIGINT to headless parent after session-id line.
- Inner trap runs; headless parent exits 0.
- Registry cleaned after shutdown.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 after forwarded ctrl-c, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.RegistryExists || len(resp.RegistryIDs) != 0 {
		t.Fatalf("registry should be cleaned after ctrl-c shutdown, ids=%v", resp.RegistryIDs)
	}
}
```