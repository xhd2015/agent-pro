## Expected
- Returns exactly 3 events.
- next_cursor.offset is 3.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "c2-"+t.Name(), "--limit", "3")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    events, _ := obj["events"].([]any)
    if events == nil || len(events) != 3 {
        t.Fatalf("expected 3 events, got %v", len(events))
    }
    nc, _ := obj["next_cursor"].(map[string]any)
    nextOff, _ := toInt(nc["offset"])
    if nextOff != 3 {
        t.Fatalf("expected next_cursor offset 3, got %v", nextOff)
    }
}
```
