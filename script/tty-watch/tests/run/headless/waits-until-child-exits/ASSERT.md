## Expected

- Prints `session-id: ...` then exits promptly (<10s).
- Exit code 0.
- Registry entry removed after exit.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Elapsed > 10*time.Second {
		t.Fatalf("headless true took too long: %s", resp.Elapsed)
	}
	if !strings.HasPrefix(strings.TrimSpace(resp.Stdout), "session-id: ") {
		t.Fatalf("missing session-id line: %q", resp.Stdout)
	}
	if resp.RegistryExists || len(resp.RegistryIDs) != 0 {
		t.Fatalf("registry should be empty after natural exit, ids=%v", resp.RegistryIDs)
	}
}
```