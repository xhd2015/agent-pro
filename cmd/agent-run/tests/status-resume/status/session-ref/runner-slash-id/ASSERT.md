---
label: e2e
---

## Expected

- Exit code 0.
- Stdout identifies the session (`grok-tty/test-ref-s1` display and/or bare `test-ref-s1`).
- Stdout ends with trailing newline `\n`.

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
	compound := req.Runner + "/" + req.SessionID
	if !strings.Contains(resp.Stdout, compound) && !strings.Contains(resp.Stdout, req.SessionID) {
		t.Fatalf("status stdout missing session id %q or %q:\n%s", compound, req.SessionID, resp.Stdout)
	}
	assertTrailingNewline(t, resp.Stdout, "status stdout")
}
```
