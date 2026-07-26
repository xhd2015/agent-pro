---
label: e2e
---

## Expected

- Exit code 0 within 5 seconds.
- stderr must **not** contain `mirror sessions: not ready` when no session has `events.jsonl`
  (normal interactive `/exit` with empty session — not a failure).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertNotContains(t, combined, "mirror sessions: not ready")
}
```