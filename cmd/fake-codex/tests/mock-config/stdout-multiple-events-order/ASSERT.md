## Expected
- The command succeeds.
- stdout contains both events in configured order.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    first := strings.Index(resp.Stdout, "first response")
    second := strings.Index(resp.Stdout, "second response")
    if first < 0 || second < 0 || first > second {
        t.Fatalf("events not in order:\n%s", resp.Stdout)
    }
}
```

