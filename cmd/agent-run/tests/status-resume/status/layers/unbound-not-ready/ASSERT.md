---
label: e2e
---

## Expected

- Exit code 0.
- Runner status unbound (empty session id).
- Resume ready no.
- Stdout ends with trailing `\n`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	out := resp.Stdout
	low := strings.ToLower(out)
	assertContains(t, out, req.Runner+"/"+req.SessionID)
	assertContains(t, low, "unbound")
	assertContainsAny(t, low, "ready: no", "ready:no", "ready:     no")
	assertTrailingNewline(t, out, "status stdout")
}
```
