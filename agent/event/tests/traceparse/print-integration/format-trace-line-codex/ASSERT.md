## Expected
- Behavior matches consolidated trace parsing semantics for `print-integration/format-trace-line-codex`.

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

	assertContains(t, resp.Output, "ASSISTANT")
	assertContains(t, resp.Output, "Here is the answer.")

}
```
