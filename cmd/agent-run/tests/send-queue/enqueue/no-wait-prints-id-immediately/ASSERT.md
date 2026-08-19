---
label: e2e
---

## Expected

- Exit code 0.
- Stdout is a single `msg_<n>\n` line.
- Command returns in under 1 second.

## Exit Code

0

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	assertMsgIDLine(t, resp.Stdout)
	if resp.SendDuration >= 3*time.Second {
		t.Fatalf("expected --no-wait return in <3s, took %v", resp.SendDuration)
	}
}
```