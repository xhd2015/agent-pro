## Expected

- Exit code 1.
- Stderr mentions unknown runner (case-insensitive substring `unknown`).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	if !strings.Contains(strings.ToLower(resp.Stderr), "unknown") {
		t.Fatalf("expected stderr to mention unknown runner, got:\n%s", resp.Stderr)
	}
}
```