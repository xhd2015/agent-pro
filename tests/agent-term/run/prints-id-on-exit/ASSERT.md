## Expected

- Exit code 0.
- Stdout is a single line matching `session-N` pattern.
- Stderr is empty (no attach noise during session).

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
		t.Fatalf("exit code %d stderr %q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", resp.Stderr)
	}
	line := strings.TrimSpace(resp.Stdout)
	if !strings.HasPrefix(line, "session-") {
		t.Fatalf("stdout should be session id only, got %q", resp.Stdout)
	}
	if strings.Count(resp.Stdout, "\n") > 1 {
		t.Fatalf("stdout should be single line, got %q", resp.Stdout)
	}
}
```