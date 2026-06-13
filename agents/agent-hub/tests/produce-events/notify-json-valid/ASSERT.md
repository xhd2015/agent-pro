## Expected
- ExitCode 0.
- Response JSON has event_id (non-empty), partition (YYYY-MM-DD), offset (>=0), received_at (non-empty).
- Fetch returns event with runner:"fake-opencode".

```go
import (
    "encoding/json"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    obj := parseJSON(t, resp.Stdout)
    eventID, _ := obj["event_id"].(string)
    if eventID == "" {
        t.Fatal("event_id is empty")
    }
    if _, ok := obj["partition"].(string); !ok {
        t.Fatal("partition missing")
    }
    if n, ok := toInt(obj["offset"]); !ok || n < 0 {
        t.Fatal("offset missing or negative")
    }
    if _, ok := obj["received_at"].(string); !ok {
        t.Fatal("received_at missing")
    }

    // verify persistence via fetch
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "test-"+t.Name(), "--limit", "1")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    fr := parseJSON(t, r.Stdout)
    events, _ := fr["events"].([]any)
    if events == nil || len(events) < 1 {
        t.Fatal("fetch returned no events")
    }
    e := events[0].(map[string]any)
    ev := e["event"].(map[string]any)
    if ev["runner"] != "fake-opencode" {
        t.Fatalf("expected runner fake-opencode, got %v", ev["runner"])
    }
    _ = json.Unmarshal
}
```
