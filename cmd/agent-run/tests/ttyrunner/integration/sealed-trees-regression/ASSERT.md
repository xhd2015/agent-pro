## Expected

- Documents sealed tree regression commands.
- No modification to sealed trees.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.SealedTreesDoc == "" { t.Fatal("expected sealed trees doc pointer") }
	if !strings.Contains(resp.SealedTreesDoc, "cmd/agent-run/tests/tty") {
		t.Fatal("doc must reference sealed tty tree")
	}
}
```
