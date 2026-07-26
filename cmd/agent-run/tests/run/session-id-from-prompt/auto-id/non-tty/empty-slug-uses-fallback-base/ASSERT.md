---
label: e2e
---

## Expected

- Exit code 0.
- Session id matches auto-id shape.
- Base is `sess` or `task` (implementation may choose either fallback name).

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

	id := singleSessionID(t, req.Home, "fake-codex")
	base, ts, _, ok := splitAutoSessionID(id)
	if !ok {
		t.Fatalf("session id %q does not match auto-id shape", id)
	}
	if base != "sess" && base != "task" {
		t.Fatalf("fallback base = %q, want sess or task (id=%q)", base, id)
	}
	if ts == "" {
		t.Fatalf("missing timestamp in id %q", id)
	}
}
```
