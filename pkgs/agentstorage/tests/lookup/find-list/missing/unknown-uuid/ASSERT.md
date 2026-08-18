## Expected

- Find error contains `not found` and the queried UUID.
- Exact shape: `session not found: no grok session with runner_session_id <uuid>`.

```go
import (
	"fmt"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	want := fmt.Sprintf("session not found: no grok session with runner_session_id %s", req.QueryID)
	assertExactErr(t, resp.Err, want)
	assertErrContains(t, resp.Err, "not found", req.QueryID)
}
```
