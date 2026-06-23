## Expected

- Output contains `msg-four` and `msg-five`.
- Output does not contain `msg-one`, `msg-two`, or `msg-three`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "msg-four")
	assertContains(t, resp.Output, "msg-five")
	assertNotContains(t, resp.Output, "msg-one")
	assertNotContains(t, resp.Output, "msg-two")
	assertNotContains(t, resp.Output, "msg-three")
}
```