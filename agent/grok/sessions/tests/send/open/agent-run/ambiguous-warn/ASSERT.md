## Expected

- Ambiguous mapping warns on stderr and falls back to bare grok `--resume`.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if !strings.Contains(resp.Stderr, "warning:") || !strings.Contains(resp.Stderr, "ambiguous") {
		t.Fatalf("stderr want ambiguous warning:\n%s", resp.Stderr)
	}
	if len(resp.Opened) != 1 || !strings.Contains(resp.Opened[0], "--resume") {
		t.Fatalf("Opened=%v, want grok --resume fallback", resp.Opened)
	}
	if len(resp.SendCalls) != 1 {
		t.Fatalf("SendCalls=%#v", resp.SendCalls)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
opened: new window; resuming `+req.SessionID+`
sent to session `+req.SessionID+`
`)
}
```
