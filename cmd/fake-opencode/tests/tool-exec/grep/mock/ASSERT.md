---
label: e2e
---

## Expected
- The event output is `"fake grep result\nfake_file.txt:1: fake match"` — the mock value.
- The output does **not** contain `REAL_MATCH` (proving mock bypassed real grep).

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    events := parseJSONLines(t, resp.Stdout)
    if len(events) == 0 {
        t.Fatal("no events in stdout")
    }
    event := events[0]
    part, _ := event["part"].(map[string]any)
    state, _ := part["state"].(map[string]any)
    output, _ := state["output"].(string)
    if !strings.Contains(output, "fake grep result") {
        t.Fatalf("expected mock output 'fake grep result', got: %q", output)
    }
    if strings.Contains(output, "REAL_MATCH") {
        t.Fatalf("mock should prevent real grep, but got real match: %q", output)
    }
}
```
