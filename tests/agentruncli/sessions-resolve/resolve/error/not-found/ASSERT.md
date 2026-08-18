## Errors

- Exact: `session not found: no grok session with runner_session_id <uuid>`
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
	uuid := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	want := fmt.Sprintf("session not found: no grok session with runner_session_id %s", uuid)
	assertExactErr(t, resp.Err, want)
	assertStdoutEmpty(t, resp.Stdout)
}
```
