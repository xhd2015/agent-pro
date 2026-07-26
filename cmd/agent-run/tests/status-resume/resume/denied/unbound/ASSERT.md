---
label: e2e
---

## Expected

- Exit code 1.
- Error indicates runner session not bound / missing runner_session_id.

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
		"not bound",
		"unbound",
		"runner_session_id",
		"no runner session",
		"missing runner",
		"not resolved",
	)
}
```
