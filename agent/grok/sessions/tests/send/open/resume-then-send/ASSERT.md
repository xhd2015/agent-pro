## Expected

- Opens resume window, then sends; two stdout lines.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.Opened) != 1 {
		t.Fatalf("Opened = %v, want 1", resp.Opened)
	}
	if !strings.Contains(resp.Opened[0], "--resume") || !strings.Contains(resp.Opened[0], req.SessionID) {
		t.Fatalf("Opened follow-up missing resume: %q", resp.Opened[0])
	}
	assertSentDefaults(t, resp, "hello")
	assert.Output(t, resp.Stdout, `---
version: 3
---
opened: new window; resuming `+req.SessionID+`
sent to session `+req.SessionID+`
`)
}
```
