## Expected
- One grok event with type `thought` and `data` containing the thinking text.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"thought"`)
	assertContains(t, resp.Output, `"data":"Let me think about this..."`)
}
```
