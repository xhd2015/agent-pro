---
label: e2e
---

## Expected
- Exit code 0.
- Marker file exists and contains `"session_id":"sess_mock"`.

```go
import (
    "encoding/json"
    "os"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    data, err := os.ReadFile(req.MarkerPath)
    if err != nil {
        t.Fatalf("marker file not written: %v", err)
    }
    var obj map[string]any
    if err := json.Unmarshal(data, &obj); err != nil {
        t.Fatalf("invalid marker JSON: %v", err)
    }
    sid, _ := obj["session_id"].(string)
    if sid != "sess_mock" {
        t.Fatalf("expected session_id=sess_mock, got %q", sid)
    }
}
```
