## Expected

- Exit code 0.
- PTY capture trimmed to lines is a single `session-N` line (no extra stdout noise).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d combined %q", resp.ExitCode, resp.Combined)
	}
	trimmed := strings.TrimSpace(resp.Combined)
	if strings.Count(trimmed, "\n") > 0 {
		t.Fatalf("stdout should be single line session id, got %q", resp.Combined)
	}
	if !strings.HasPrefix(trimmed, "session-") {
		t.Fatalf("stdout should be session id only, got %q", resp.Combined)
	}
}
```