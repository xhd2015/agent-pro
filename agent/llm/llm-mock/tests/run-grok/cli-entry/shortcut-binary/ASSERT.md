---
label: e2e
---

## Expected

- `llm-mock-run-grok` exits 0 with fake grok hook.
- Output contains `GROK_HOME=` confirming orchestrator ran.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "GROK_HOME=")
}
```
