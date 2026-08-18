## Errors

- Exact: `ambiguous grok-session-id <uuid>: multiple matches: sess-a, sess-b`
- Stdout empty.

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
	uuid := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	want := fmt.Sprintf("ambiguous grok-session-id %s: multiple matches: sess-a, sess-b", uuid)
	assertExactErr(t, resp.Err, want)
	assertStdoutEmpty(t, resp.Stdout)
}
```
