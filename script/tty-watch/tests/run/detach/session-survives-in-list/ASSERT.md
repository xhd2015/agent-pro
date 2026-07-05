## Expected

- Detach parent exits 0 after printing session-id line.
- `list` shows the session with command containing `sleep` and `120`.
- Session TCP endpoint reachable.

## Exit Code

0 (`list` subcommand)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("list exit %d, output %q", resp.ExitCode, resp.ListOutput)
	}
	if !strings.HasPrefix(strings.TrimSpace(resp.Stdout), "session-id: ") {
		t.Fatalf("missing session-id line: %q", resp.Stdout)
	}
	if !resp.RegistryExists {
		t.Fatalf("registry should exist after detach parent exited")
	}
	assertRegistryHasSession(t, req.TTYWatchHome, resp.SessionID)
	if !resp.SessionRunning {
		t.Fatalf("session %s should be reachable after detach", resp.SessionID)
	}
	if !strings.Contains(resp.ListOutput, resp.SessionID) {
		t.Fatalf("list missing session id %q: %q", resp.SessionID, resp.ListOutput)
	}
	if !strings.Contains(resp.ListOutput, "sleep") {
		t.Fatalf("list missing sleep command: %q", resp.ListOutput)
	}
	if !strings.Contains(resp.ListOutput, "120") {
		t.Fatalf("list missing sleep duration: %q", resp.ListOutput)
	}
}
```