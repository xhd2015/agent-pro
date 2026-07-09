## Expected

- Parent completes within the harness timeout (does not hang).
- Exit code 0 and PTY output contains `yes`.
- After parent exit (+ 2.5s grace), **no** process remains whose command line
  contains the session id `bash-c-no-orphan` (no KeepAlive `__serve__` orphan).
- Registry entry for the session is gone.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.TimedOut {
		t.Fatalf("tty-watch run bash -c 'echo yes' hung (timed out after %s), output %q", resp.Elapsed, resp.Combined)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d output %q", resp.ExitCode, resp.Combined)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if !strings.Contains(combined, "yes") {
		t.Fatalf("expected yes in PTY output, got %q", combined)
	}
	if resp.SessionRunning {
		t.Fatalf("orphan __serve__ still running for session %q after short bash -c run (KeepAlive leak); registry ids %v output %q",
			resp.SessionID, resp.RegistryIDs, combined)
	}
	if resp.RegistryExists {
		t.Fatalf("registry entry should be cleaned after attached exit, ids %v", resp.RegistryIDs)
	}
}
```
