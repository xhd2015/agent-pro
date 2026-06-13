## Expected
- One codex event with type `error` and message containing the error text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"error"`)
	assertContains(t, resp.Stdout, `"something went wrong"`)
}
```
