## Expected

- Returned JSON is the file contents.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "JSON", resp.JSON, `{"a":1}`)
}
```
