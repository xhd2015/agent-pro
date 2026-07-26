---
label: e2e
---

## Expected

- Exit code 0 (orchestrator starts mock server and fake grok without config env).
- Combined stdout/stderr contains `GROK_HOME=` (mock + grok plumbing succeeded).

## Errors

- Must not fail with missing-config error before grok starts.

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