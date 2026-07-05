## Expected

- Exit code 0.
- Fake opencode prints `OPENCODE_CONFIG_DIR=` line (orchestrator started mock + opencode).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "OPENCODE_CONFIG_DIR=")
}
```