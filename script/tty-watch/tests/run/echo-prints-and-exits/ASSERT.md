## Expected

- PTY capture contains `yes` from the child command.
- `tty-watch run echo yes` exits 0 without hanging past the harness timeout.
- Host stays silent (no session-id lines).

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
		t.Fatalf("tty-watch run echo yes hung (timed out after %s), output %q", resp.Elapsed, resp.Combined)
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
	for _, line := range strings.Split(combined, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "session-") {
			t.Fatalf("host leaked session id into PTY output: %q", combined)
		}
	}
}
```