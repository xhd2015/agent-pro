## Expected

- No flag parse error on stderr; headless starts successfully.
- Host stdout is only the session-id line.
- PTY output (via `watch`) contains `--not-a-flag` proving the token reached the command.

## Exit Code

0

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
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "unknown flag") || strings.Contains(lower, "flag provided") {
		t.Fatalf("command token parsed as flag: stderr=%q", resp.Stderr)
	}
	if !strings.HasPrefix(strings.TrimSpace(resp.Stdout), "session-id: ") {
		t.Fatalf("missing session-id line: %q", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "--not-a-flag") {
		t.Fatalf("PTY echo leaked to host stdout: %q", resp.Stdout)
	}
	if !strings.Contains(resp.WatchOutput, "--not-a-flag") {
		t.Fatalf("watch should show command output --not-a-flag, got %q", resp.WatchOutput)
	}
}
```