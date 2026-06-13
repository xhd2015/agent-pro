## Expected
- Exit code 0.
- Stdout shows installed=false.

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
    if installed {
        t.Fatal("expected installed=false")
    }
}
```
