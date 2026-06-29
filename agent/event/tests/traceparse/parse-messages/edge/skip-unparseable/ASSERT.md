## Expected
- Behavior matches consolidated trace parsing semantics for `parse-messages/edge/skip-unparseable`.

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
	assertContains(t, resp.Output, "kept")

}
```
