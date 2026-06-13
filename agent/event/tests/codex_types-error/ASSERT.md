## Expected
- One codex event with type `error` and message containing the error text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"something went wrong"`)
}
```
