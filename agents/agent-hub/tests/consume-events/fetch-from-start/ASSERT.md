## Expected
- Returns 1 event.
- previous_cursor has offset 0.
- next_cursor is advanced.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "c1-"+t.Name(), "--limit", "10")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    events, _ := obj["events"].([]any)
    if events == nil || len(events) < 1 {
        t.Fatal("expected at least 1 event")
    }
    pc, _ := obj["previous_cursor"].(map[string]any)
    nc, _ := obj["next_cursor"].(map[string]any)
    if pc == nil || nc == nil {
        t.Fatal("cursors missing")
    }
    prevOff, _ := toInt(pc["offset"])
    nextOff, _ := toInt(nc["offset"])
    if prevOff != 0 {
        t.Fatalf("expected previous_cursor offset 0, got %v", prevOff)
    }
    if nextOff <= prevOff {
        t.Fatalf("expected next_cursor > previous_cursor, got next=%v prev=%v", nextOff, prevOff)
    }
}
```
