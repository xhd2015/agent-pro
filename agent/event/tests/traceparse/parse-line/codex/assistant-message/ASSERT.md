## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/codex/assistant-message`.

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
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, "Here is the answer.")

}
```
