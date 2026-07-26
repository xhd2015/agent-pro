---
label: e2e
---

## Expected

- Exit code 1.
- Stderr mentions the session (id or "session").

## Exit Code

1

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "session") && !strings.Contains(lower, "missing") {
		t.Fatalf("expected stderr to mention missing session, got:\n%s", resp.Stderr)
	}
}
```
