## Expected

- Log output contains `EDIT` label for the apply_patch tool call.
- Output references the patched file `src/main.go`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "EDIT")
	assertContains(t, resp.Output, "src/main.go")
}
```