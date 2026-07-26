## Expected
- fetch --limit 100 returns events from at least 1 partition (multi-partition if enough).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "ccross-"+t.Name(), "--limit", "100")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    events, _ := obj["events"].([]any)
    if events == nil || len(events) < 2 {
        t.Fatalf("expected at least 2 events across partitions, got %v", len(events))
    }
    // events should have at least one partition
    partitions := map[string]bool{}
    for _, e := range events {
        env := e.(map[string]any)
        if p, ok := env["partition"].(string); ok {
            partitions[p] = true
        }
    }
    // In a simple test we may only have 1 partition (same day), which is acceptable
    if len(partitions) < 1 {
        t.Fatal("expected at least one partition")
    }
}
```
