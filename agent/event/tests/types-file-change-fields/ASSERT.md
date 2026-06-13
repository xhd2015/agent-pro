## Expected
- JSON contains `path` and `kind` fields.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"path":"bar.go"`)
	assertContains(t, resp.Stdout, `"kind":"modify"`)
}
```
