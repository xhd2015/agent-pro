## Expected

- Find error exactly:
  `ambiguous grok-session-id <uuid>: multiple matches: first-match, second-match`
- Gen bump from CreateSession forces rebuild so the second match is visible.

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
	want := fmt.Sprintf("ambiguous grok-session-id %s: multiple matches: first-match, second-match", req.QueryID)
	assertExactErr(t, resp.Err, want)
}
```
