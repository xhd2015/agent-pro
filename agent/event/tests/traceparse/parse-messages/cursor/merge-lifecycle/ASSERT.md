## Expected
- Behavior matches consolidated trace parsing semantics for `parse-messages/cursor/merge-lifecycle`.

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
	if msgs[0]["finished_at"] == nil { t.Fatal("finished_at should be set") }
	assertContains(t, resp.Output, `"call_id":"cursor_1"`)
	assertContains(t, resp.Output, `"tool_name":"Shell"`)
	assertContains(t, resp.Output, `"status":"completed"`)

}
```
