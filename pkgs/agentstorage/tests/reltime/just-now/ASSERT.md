## Expected

- All three cases return exactly `just now` (no trailing ` ago`).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertCases(t, req, resp, err)
}
```
