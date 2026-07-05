## Expected

- `llm-mock run opencode` exits 0 with fake opencode hook.
- Output contains `OPENCODE_CONFIG_DIR=` confirming opencode child ran.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "OPENCODE_CONFIG_DIR=")
}
```