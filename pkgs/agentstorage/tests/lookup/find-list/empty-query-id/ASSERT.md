## Expected

- Both Find and List return exactly:
  `--grok-session-id requires a non-empty value`
- No session is created or matched.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	const want = "--grok-session-id requires a non-empty value"
	assertExactErr(t, resp.FindErr, want)
	assertExactErr(t, resp.ListErr, want)
}
```
