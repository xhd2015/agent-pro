## Expected

- Exit code 1 (L2: Handle early `validateRunner` failure).
- Stderr mentions unknown runner (case-insensitive substring `unknown`).

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	if !strings.Contains(strings.ToLower(resp.Stderr), "unknown") {
		t.Fatalf("expected stderr to mention unknown runner, got:\n%s", resp.Stderr)
	}
}
```