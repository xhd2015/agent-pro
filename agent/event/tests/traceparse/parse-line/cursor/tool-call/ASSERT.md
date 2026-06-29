## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/cursor/tool-call`.

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

	if !resp.OK { t.Fatal("expected parse ok") }
	assertContains(t, resp.Output, `"call_id":"cursor_1"`)
	assertContains(t, resp.Output, `"tool_name":"Shell"`)
	assertContains(t, resp.Output, `"status":"completed"`)

}
```
