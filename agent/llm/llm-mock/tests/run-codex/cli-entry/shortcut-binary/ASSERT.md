---
label: e2e
---

## Expected

- `llm-mock-run-codex` exits 0 with fake codex hook.
- Output contains `CODEX_HOME=` confirming orchestrator ran.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "CODEX_HOME=")
}
```