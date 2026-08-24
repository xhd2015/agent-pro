## Expected

- Dry-run prints would-capture plan; Contents not called.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoContents(t, resp)
	if !strings.Contains(resp.Stdout, "Would capture") {
		t.Fatalf("stdout missing Would capture:\n%s", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
Would capture: window 3, tab 1
  grok id:   `+req.SessionID+`
  iterm id:  w2t1p0
  source:    iterm
`)
}
```
