---
label: e2e
---

## Expected

- HTTP 200.
- Response includes both `codex-tty` and `grok-tty`.

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
	assertHTTPStatus(t, resp, 200)
	if !strings.Contains(resp.HTTPBody, "codex-tty") || !strings.Contains(resp.HTTPBody, "grok-tty") {
		t.Fatalf("runner list missing tty runners: %s", resp.HTTPBody)
	}
}
```
