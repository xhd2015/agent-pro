## Expected
- One grok event with type `text` and `data` containing the error message.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"data":"something went wrong"`)
}
```
