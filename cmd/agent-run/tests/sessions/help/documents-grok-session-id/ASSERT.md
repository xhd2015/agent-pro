---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `--grok-session-id`.
- Stdout still documents `--print` (print mode).

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
	assertContains(t, resp.Stdout, "--grok-session-id")
	assertContains(t, resp.Stdout, "--print")
}
```
