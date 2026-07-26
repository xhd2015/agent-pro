## Expected
- One grok event with type `text` and `data` containing the message text.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"data":"Hello world"`)
}
```
