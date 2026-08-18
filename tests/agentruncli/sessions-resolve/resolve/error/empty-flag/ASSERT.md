## Errors

- Exact: `--grok-session-id requires a non-empty value`
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
	assertExactErr(t, resp.Err, "--grok-session-id requires a non-empty value")
	assertStdoutEmpty(t, resp.Stdout)
}
```
