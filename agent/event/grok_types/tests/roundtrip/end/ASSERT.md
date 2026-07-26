## Expected
- Roundtripped output contains `type:done` and the session ID is preserved.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"sess_xyz_789"`)
}
```
