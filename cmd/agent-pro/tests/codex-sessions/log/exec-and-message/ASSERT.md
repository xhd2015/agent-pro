## Expected

- Log output contains `ASSISTANT` and message text `Hello from codex`.
- Log output contains `RUN` and command/output related to `echo hi`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "ASSISTANT")
	assertContains(t, resp.Output, "Hello from codex")
	assertContains(t, resp.Output, "RUN")
	assertContains(t, resp.Output, "echo hi")
}
```