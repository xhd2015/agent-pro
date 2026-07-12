## Expected

| Age | Output |
|-----|--------|
| 1s | `1s ago` |
| 2s | `2s ago` |
| 1h | `1h ago` |
| 90d | `90d ago` |

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertCases(t, req, resp, err)
}
```
