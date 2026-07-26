## Expected
- fetch returns an agent.prompt.submitted event with prompt:"hello world".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r, err := runAgentHub(t, req, "fetch", "--consumer-id", "test-"+t.Name(), "--limit", "10")
    if err != nil {
        t.Fatalf("fetch error: %v", err)
    }
    assertSuccess(t, r)
    fr := parseJSON(t, r.Stdout)
    events := fr["events"].([]any)
    if events == nil || len(events) < 1 {
        t.Fatal("expected at least 1 event")
    }

    found := false
    for _, e := range events {
        ev := e.(map[string]any)["event"].(map[string]any)
        if ev["event_type"] == "agent.prompt.submitted" && ev["prompt"] == "hello world" {
            found = true
            break
        }
    }
    if !found {
        t.Fatal("agent.prompt.submitted event with prompt 'hello world' not found")
    }
}
```
