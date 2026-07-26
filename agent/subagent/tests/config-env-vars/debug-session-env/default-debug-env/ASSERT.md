## Expected
- Stdout lists `debug_default_session` from the default debug env dir.
- The default `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` was checked and used.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "debug_default_session") {
        t.Fatalf("expected 'debug_default_session' in output (from default debug env), got:\n%s", resp.Stdout)
    }
}
```
