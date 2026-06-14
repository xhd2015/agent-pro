## Expected
- Exit code 0.
- Marker file exists with `session_id: "sess_autodiscover"`.

```go
import (
    "encoding/json"
    "os"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    data, err := os.ReadFile(req.MarkerPath)
    if err != nil {
        t.Fatalf("marker file not written (plugin not auto-discovered): %v", err)
    }
    var obj map[string]any
    if err := json.Unmarshal(data, &obj); err != nil {
        t.Fatalf("invalid marker JSON: %v", err)
    }
    sid, _ := obj["session_id"].(string)
    if sid != "sess_autodiscover" {
        t.Fatalf("expected session_id=sess_autodiscover, got %q", sid)
    }
}
```
