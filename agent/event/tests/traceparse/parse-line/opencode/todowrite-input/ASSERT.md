## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/opencode/todowrite-input`.

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
	assertContains(t, resp.Output, `"tool_name":"Plan"`)
	assertContains(t, resp.Output, "in_progress: Ship feature")

}
```
