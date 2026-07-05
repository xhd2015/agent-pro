## Expected

- Exit code 0 (orchestrator starts mock server and fake opencode without config env).
- Combined stdout/stderr contains `OPENCODE_CONFIG_DIR=` (mock + opencode plumbing succeeded).

## Errors

- Must not fail with missing-config error before opencode starts.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "OPENCODE_CONFIG_DIR=")
}
```