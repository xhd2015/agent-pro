## Expected

| Age | Output |
|-----|--------|
| 65s | `1m5s ago` |
| 1h2m | `1h2m ago` |

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertCases(t, req, resp, err)
}
```
