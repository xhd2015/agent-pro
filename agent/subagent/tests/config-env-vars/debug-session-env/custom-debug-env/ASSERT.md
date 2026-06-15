## Expected
- Stdout lists `debug_custom_session` from the custom debug env dir.
- The custom `DebugSessionEnv` override was respected.
- Sessions from the `SessionBase` dir are NOT listed.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "debug_custom_session") {
        t.Fatalf("expected 'debug_custom_session' in output (from custom debug env), got:\n%s", resp.Stdout)
    }
}
```
