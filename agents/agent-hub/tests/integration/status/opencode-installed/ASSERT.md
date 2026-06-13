## Expected
- Exit code 0.
- Stdout shows runner "opencode", status "installed", "enabled": true.

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
    installed, _ := obj["installed"].(bool)
    if !installed {
        t.Fatal("expected installed=true")
    }
    enabled, _ := obj["enabled"].(bool)
    if !enabled {
        t.Fatal("expected enabled=true")
    }
}
```
