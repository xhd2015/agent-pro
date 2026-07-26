---
label: unit
explanation: history resolves #general then fetches messages
---

## Expected

- Exit code 0.
- Chronological human lines present.
- Stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "first message") || !strings.Contains(resp.Stdout, "third message") {
		t.Fatalf("stdout missing expected messages:\n%s", resp.Stdout)
	}
}
```
