## Expected

- Log output contains `REASONING` label.
- Log output contains `[Redacted]` and does not contain the raw encrypted blob.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "REASONING")
	assertContains(t, resp.Output, "[Redacted]")
	assertNotContains(t, resp.Output, "gAAAAABsecretblob")
}
```