## Expected

- Exit code 1 (or other non-zero).
- Combined output indicates mutual exclusion between `--grok-session-id` and
  positional session id (or similar conflict wording).

## Exit Code

1

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
		t.Fatalf("expected non-zero exit for mutex flag+positional, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	// Prefer exact exit 1 when the feature is implemented.
	assertExitCode(t, resp, 1)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"mutually exclusive",
		"exclusive",
		"cannot be used with",
		"cannot use both",
		"not both",
		"conflict",
	)
}
```
