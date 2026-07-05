## Expected

- `llm-mock run codex` exits 0 with fake codex hook.
- Output contains `CODEX_HOME=` confirming codex child ran.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "CODEX_HOME=")
}
```