## Expected
- Returns 3 events.
- has_more:false.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "c3-"+t.Name(), "--limit", "5")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    events, _ := obj["events"].([]any)
    if events == nil || len(events) != 3 {
        t.Fatalf("expected 3 events, got %v", len(events))
    }
    if obj["has_more"] != false {
        t.Fatalf("expected has_more false, got %v", obj["has_more"])
    }
}
```
