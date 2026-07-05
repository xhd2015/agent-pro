## Expected

- Exit code 0.
- Fake codex prints `CODEX_HOME=` line (orchestrator started mock + codex).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "CODEX_HOME=")
}
```