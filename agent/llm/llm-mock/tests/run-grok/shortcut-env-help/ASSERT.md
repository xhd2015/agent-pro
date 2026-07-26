---
label: e2e
---

## Expected

- Exit code 0.
- Combined output contains `llm-mock run` usage / help text.
- Output does **not** contain `GROK_HOME=` (mock server and grok not started).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "Usage: llm-mock run")
	assertContains(t, combined, "--log-events")
	assertNotContains(t, combined, "GROK_HOME=")
}
```