## Expected
- {"sessions":[]}.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "sessions")
    if err != nil {
        t.Fatalf("sessions error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    sessions, _ := obj["sessions"].([]any)
    if sessions == nil || len(sessions) != 0 {
        t.Fatalf("expected empty sessions, got %v", len(sessions))
    }
}
```
