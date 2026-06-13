## Expected
- Exit code 0.
- Stdout contains runner entry for "opencode".

```go
import (
    "encoding/json"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)

    var obj map[string]any
    if err := json.Unmarshal([]byte(resp.Stdout), &obj); err != nil {
        t.Fatalf("invalid JSON: %v", err)
    }
    _, ok := obj["opencode"]
    if !ok {
        t.Fatal("expected opencode entry in status output")
    }
}
```
