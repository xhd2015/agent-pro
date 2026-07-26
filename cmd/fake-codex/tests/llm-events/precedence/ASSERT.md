---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains the text from `llm_events`.
- stdout does NOT contain the text from `stdout_events`.

```go
import (
	"testing"
	"strings"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"from llm"`)
    if strings.Contains(resp.Stdout, `"from stdout"`) {
        t.Fatalf("stdout_events content leaked despite llm_events precedence:\n%s", resp.Stdout)
    }
}
```
