## Expected
- JSON contains `path` and `kind` fields.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"path":"bar.go"`)
	assertContains(t, resp.Output, `"kind":"modify"`)
}
```
