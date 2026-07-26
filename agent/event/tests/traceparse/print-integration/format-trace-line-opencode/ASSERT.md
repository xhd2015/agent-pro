## Expected
- Behavior matches consolidated trace parsing semantics for `print-integration/format-trace-line-opencode`.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, resp.Output, "ASSISTANT")
	assertContains(t, resp.Output, "I have completed the task.")

}
```
