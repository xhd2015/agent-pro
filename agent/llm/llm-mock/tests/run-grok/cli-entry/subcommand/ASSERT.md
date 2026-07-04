## Expected

- `llm-mock run grok` exits 0 with fake grok hook.
- Output contains `GROK_HOME=` confirming grok child ran.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout+resp.Stderr, "GROK_HOME=")
}
```
