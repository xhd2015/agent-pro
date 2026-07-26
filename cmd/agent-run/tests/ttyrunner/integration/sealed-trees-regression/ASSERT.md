---
label: e2e
---

## Expected

- Documents sealed tree regression commands.
- No modification to sealed trees.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.SealedTreesDoc == "" { t.Fatal("expected sealed trees doc pointer") }
	if !strings.Contains(resp.SealedTreesDoc, "cmd/agent-run/tests/tty") {
		t.Fatal("doc must reference sealed tty tree")
	}
}
```
