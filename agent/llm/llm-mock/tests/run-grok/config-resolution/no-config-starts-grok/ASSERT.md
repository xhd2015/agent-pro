## Expected

- Exit code 0 (orchestrator starts mock server and fake grok without config env).
- Combined stdout/stderr contains `GROK_HOME=` (mock + grok plumbing succeeded).

## Errors

- Must not fail with missing-config error before grok starts.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "GROK_HOME=")
}
```