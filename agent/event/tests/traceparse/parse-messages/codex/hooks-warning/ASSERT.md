## Expected
- Behavior matches consolidated trace parsing semantics for `parse-messages/codex/hooks-warning`.

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
	if len(msgs) != 1 { t.Fatalf("messages = %d, want 1", len(msgs)) }
	assertContains(t, resp.Output, `"tool_name":"Config Warning"`)
	assertContains(t, resp.Output, `"status":"warning"`)

}
```
