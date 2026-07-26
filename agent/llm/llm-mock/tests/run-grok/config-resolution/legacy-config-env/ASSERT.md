---
label: e2e
---

## Expected

- Exit code 0 with legacy `LLM_MOCK_CONFIG` only.
- Fake grok prints `GROK_HOME=` confirming orchestrator ran.

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
