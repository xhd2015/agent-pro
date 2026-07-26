## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/codex/file-change`.

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

	if !resp.OK { t.Fatal("expected parse ok") }
	assertContains(t, resp.Output, `"tool_name":"File Change"`)
	assertContains(t, resp.Output, "/tmp/code-commits.json")
	assertContains(t, resp.Output, `"kind":"add"`)
	assertContains(t, resp.Output, "+ /tmp/code-commits.json")

}
```
