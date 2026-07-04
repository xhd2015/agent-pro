## Expected

- Exit code 0.
- Fake grok prints `GROK_HOME=` line (orchestrator started mock + grok).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "GROK_HOME=")
}
```
