## Expected

- `--tab 2` sends to resolved grok pane.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoOpen(t, resp)
	if len(resp.SendCalls) != 1 {
		t.Fatalf("SendCalls = %v, want 1", resp.SendCalls)
	}
	c := resp.SendCalls[0]
	if c.Text != "from-tab" {
		t.Fatalf("text = %q", c.Text)
	}
	if !strings.Contains(c.SessionID, "TAB2-UUID") {
		t.Fatalf("SendText session = %q, want TAB2-UUID", c.SessionID)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
sent to session `+fixtureTabSendSessionID+`
`)
}
```
