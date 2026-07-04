## Expected

- PTY capture contains `yes`.
- Command completes with exit 0 without hanging.

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
}
```