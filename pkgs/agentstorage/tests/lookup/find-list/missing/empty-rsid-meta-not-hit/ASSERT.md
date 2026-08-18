## Expected

- Find returns not-found for the query UUID.
- Empty `runner_session_id` on a seeded meta does not produce a hit.

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
}
```
