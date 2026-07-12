## Expected

- `FormatRelativeAge(now, time.Time{})` returns exactly `-`.
- No `ago` suffix and not `just now`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertCases(t, req, resp, err)
}
```
