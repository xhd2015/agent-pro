---
label: e2e
---

## Expected

- Exit code 1.
- Combined stderr/stdout clearly indicates session not found (or similar).

## Exit Code

1

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
	assertExitCode(t, resp, 1)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"not found",
		"no such",
		"unknown session",
		"session not found",
	)
	assertContains(t, combined, strings.ToLower(req.SessionID))
}
```
