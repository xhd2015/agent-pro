## Expected
- The command succeeds.
- stdout contains JSON events.
- Output is deterministic: contains known event types from seed 42.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.Stdout == "" {
        t.Fatal("expected JSON events on stdout")
    }
    assertContains(t, resp.Stdout, `"type":`)
    assertContains(t, resp.Stdout, `"sessionID":`)
}
```
