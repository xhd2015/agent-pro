---
label: e2e
---

## Expected

- Exit code 0.
- Fake codex prints `CODEX_HOME=` line (orchestrator started mock + codex).

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