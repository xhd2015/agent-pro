## Expected

- Exit code 1.
- Stderr mentions the session (id or "session").

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "session") && !strings.Contains(lower, "missing") {
		t.Fatalf("expected stderr to mention missing session, got:\n%s", resp.Stderr)
	}
}
```
