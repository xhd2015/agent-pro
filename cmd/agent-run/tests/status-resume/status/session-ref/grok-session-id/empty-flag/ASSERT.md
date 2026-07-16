## Expected

- Exit code ≠ 0 (parse failure, validation, or usage error).
- Stderr/stdout is non-empty error text (not a successful multi-layer status dump).

## Exit Code

≠ 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for empty --grok-session-id, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	combined := strings.TrimSpace(resp.Stderr + "\n" + resp.Stdout)
	if combined == "" {
		t.Fatal("expected error output for empty --grok-session-id")
	}
}
```
