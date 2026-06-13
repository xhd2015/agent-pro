## Expected
- First fetch returns 1 event.
- previous_cursor == next_cursor.
- Second fetch returns same event (cursor not advanced).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "cp-"+t.Name(), "--peek")
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
    if prevOff != nextOff {
        t.Fatalf("peek: expected previous_cursor == next_cursor, got %v vs %v", prevOff, nextOff)
    }

    // fetch again without peek - should get same event again
    r2, err := runAgentHub(t, req, "fetch", "--consumer-id", "cp-"+t.Name())
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r2)
    obj2 := parseJSON(t, r2.Stdout)
    events2, _ := obj2["events"].([]any)
    if events2 == nil || len(events2) < 1 {
        t.Fatal("expected same event again")
    }
}
```
