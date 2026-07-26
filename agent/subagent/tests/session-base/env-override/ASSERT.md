## Expected
- The output lists a session with ID `debug_test_123` found from the debug dir.
- The `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` overrides `SessionBase`.
- Sessions in the custom base dir are NOT listed.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "debug_test_123") {
        t.Fatalf("expected session 'debug_test_123' in output from debug dir, got:\n%s", resp.Stdout)
    }
}
```
