## Expected
- Behavior matches consolidated trace parsing semantics for `parse-messages/codex/plan-and-file-change`.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	var msgs []map[string]any
	if err := json.Unmarshal([]byte(resp.Output), &msgs); err != nil { t.Fatal(err) }
	if len(msgs) != 2 { t.Fatalf("messages = %d, want 2", len(msgs)) }
	out := resp.Output
	assertContains(t, out, `"tool_name":"Plan"`)
	assertContains(t, out, "[x] Inspect Jira comments")
	assertContains(t, out, `"tool_name":"File Change"`)
	assertContains(t, out, "+ /tmp/code-commits.json")

}
```
