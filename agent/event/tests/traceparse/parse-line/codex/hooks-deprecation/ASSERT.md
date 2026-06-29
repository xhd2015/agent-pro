## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/codex/hooks-deprecation`.

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
	assertContains(t, resp.Output, `"tool_name":"Config Warning"`)
	assertContains(t, resp.Output, `"kind":"warning"`)
	assertContains(t, resp.Output, `"status":"warning"`)

}
```
