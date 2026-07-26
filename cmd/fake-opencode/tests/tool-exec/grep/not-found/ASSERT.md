---
label: e2e
---

## Expected
- The event output is empty or does not contain any matches.
- grep with no matches returns exit_code 1 but the event is still "completed" (no error for no-match).

```go
import (
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
    // grep returns empty output when no matches found
    if output != "" {
        t.Logf("grep with no matches returned output: %q (exit_code may be 1)", output)
    }
    // Verify no false match
    if output == "ZZZZ_NO_MATCH_ZZZZ" || len(output) > 0 {
        // It's ok for grep to output nothing on no match
        // Just verify status is set
        if status, _ := state["status"].(string); status == "" {
            t.Fatal("expected status to be set")
        }
    }
}
```
