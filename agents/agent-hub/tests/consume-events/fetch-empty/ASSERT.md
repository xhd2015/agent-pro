## Expected
- events:[], has_more:false.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "cempty-"+t.Name())
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    events, _ := obj["events"].([]any)
    if events == nil || len(events) != 0 {
        t.Fatalf("expected empty events, got %v", len(events))
    }
    if obj["has_more"] != false {
        t.Fatalf("expected has_more false, got %v", obj["has_more"])
    }
}
```
