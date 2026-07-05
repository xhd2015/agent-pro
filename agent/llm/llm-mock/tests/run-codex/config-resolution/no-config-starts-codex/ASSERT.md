## Expected

- Exit code 0 (orchestrator starts mock server and fake codex without config env).
- Combined stdout/stderr contains `CODEX_HOME=` (mock + codex plumbing succeeded).

## Errors

- Must not fail with missing-config error before codex starts.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "CODEX_HOME=")
}
```