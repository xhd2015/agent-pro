## Expected
- First fetch returns 3 events (offset 3).
- Replay resets cursor.
- Second fetch returns 5 events (starts from offset 0).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    cid := "creplay-" + t.Name()

    r, err := runAgentHub(t, req, "fetch", "--consumer-id", cid, "--limit", "3")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    events, _ := obj["events"].([]any)
    if events == nil || len(events) != 3 {
        t.Fatalf("expected 3 events, got %v", len(events))
    }

    // get partition for replay
    pc, _ := obj["previous_cursor"].(map[string]any)
    partition, _ := pc["partition"].(string)
    if partition == "" {
        nc, _ := obj["next_cursor"].(map[string]any)
        partition, _ = nc["partition"].(string)
    }

    // replay to reset cursor
    rr, err := runAgentHub(t, req, "replay", "--consumer-id", cid, "--from", partition+":0")
    if err != nil {
        t.Fatalf("replay error: %v", err)
    }
    assertSuccess(t, rr)

    // fetch again, should get all 5
    r2, err := runAgentHub(t, req, "fetch", "--consumer-id", cid, "--limit", "5")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r2)
    obj2 := parseJSON(t, r2.Stdout)
    events2, _ := obj2["events"].([]any)
    if events2 == nil || len(events2) != 5 {
        t.Fatalf("expected 5 events after replay, got %v", len(events2))
    }
}
```
