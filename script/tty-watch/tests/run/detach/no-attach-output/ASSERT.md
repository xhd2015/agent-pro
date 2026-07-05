## Expected

- Host stdout is only the `session-id: ...` line.
- `DETACH_MARKER` must not appear on host stdout.

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
	lines := strings.Split(strings.TrimRight(resp.Stdout, "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "session-id: ") {
		t.Fatalf("first stdout line must be session-id, got %q", resp.Stdout)
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			t.Fatalf("unexpected host stdout after session-id line: %q", resp.Stdout)
		}
	}
	if strings.Contains(resp.Stdout, "DETACH_MARKER") {
		t.Fatalf("PTY marker leaked to host stdout: %q", resp.Stdout)
	}
}
```