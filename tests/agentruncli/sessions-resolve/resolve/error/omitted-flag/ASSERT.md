## Errors

- Exact: `sessions resolve requires --grok-session-id`
- Stdout empty.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	assertExactErr(t, resp.Err, "sessions resolve requires --grok-session-id")
	assertStdoutEmpty(t, resp.Stdout)
}
```
