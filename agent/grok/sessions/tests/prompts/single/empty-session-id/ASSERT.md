## Errors

- Non-nil error containing `grok session not found`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "grok session not found")
}
```
