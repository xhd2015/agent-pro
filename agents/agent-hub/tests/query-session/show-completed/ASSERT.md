## Expected
- status:"completed".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    r, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "s_comp")
    if err != nil {
        t.Fatalf("session show error: %v", err)
    }
    assertSuccess(t, r)
    obj := parseJSON(t, r.Stdout)
    if obj["status"] != "completed" {
        t.Fatalf("expected status completed, got %v", obj["status"])
    }
}
```
