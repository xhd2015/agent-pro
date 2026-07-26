---
label: e2e
---

## Expected

- Exit code 0.
- Generated session id base is `hello-world` (lowercase, punctuation → `-`,
  collapsed, trimmed).
- Full id matches auto-id shape with timestamp.

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
	if base != "hello-world" {
		t.Fatalf("slug base = %q, want hello-world (id=%q)", base, id)
	}
	if len(ts) != 15 { // YYYYMMDD-HHMMSS
		t.Fatalf("timestamp part %q unexpected length in id %q", ts, id)
	}
}
```
