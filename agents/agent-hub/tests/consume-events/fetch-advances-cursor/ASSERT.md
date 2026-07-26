## Expected
- First fetch returns event at offset 0.
- Second fetch returns event at offset 1.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    cid := "ca-" + t.Name()

    r, err := runAgentHub(t, req, "fetch", "--consumer-id", cid, "--limit", "1")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    events, _ := obj["events"].([]any)
    if events == nil || len(events) != 1 {
        t.Fatalf("expected 1 event, got %v", len(events))
    }
    nc1, _ := obj["next_cursor"].(map[string]any)
    off1, _ := toInt(nc1["offset"])
    if off1 != 1 {
        t.Fatalf("expected first next_cursor offset 1, got %v", off1)
    }

    r2, err := runAgentHub(t, req, "fetch", "--consumer-id", cid, "--limit", "1")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r2)
    obj2 := parseJSON(t, r2.Stdout)
    nc2, _ := obj2["next_cursor"].(map[string]any)
    off2, _ := toInt(nc2["offset"])
    if off2 != 2 {
        t.Fatalf("expected second next_cursor offset 2, got %v", off2)
    }
}
```
