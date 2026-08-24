## Expected

- Error contains `both`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrContains(t, req, resp, err)
}
```
