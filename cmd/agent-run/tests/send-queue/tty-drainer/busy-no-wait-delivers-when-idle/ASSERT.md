---
label: e2e
---

## Expected

- Exit code 0 from `--no-wait` send.
- Stdout is a single `msg_<n>\n` line.
- CLI returns in under 1 second while terminal is still busy.
- After the session becomes idle (no blocking send), message is injected.
- `msg status` reports `delivered`.
- Queue file no longer contains the message id.

## Side Effects

- Complements `wait/no-wait-returns-before-delivery` (return-before-inject on busy):
  this leaf asserts eventual inject by the session-owned drainer after idle.
- No default/blocking CLI send is issued.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	assertMsgIDLine(t, resp.Stdout)
	if resp.SendDuration >= 3*time.Second {
		t.Fatalf("expected --no-wait return in <3s while busy, took %v", resp.SendDuration)
	}
	if !containsString(resp.InjectedMessages, "tty-drainer-busy-probe") {
		t.Fatalf("expected tty-drainer-busy-probe injected after busy→idle without blocking send, seen=%v", resp.InjectedMessages)
	}
	if strings.TrimSpace(resp.StatusAfterStdout) != "delivered" {
		t.Fatalf("status after TTY-side delivery want delivered, got %q", resp.StatusAfterStdout)
	}
	if resp.QueueHasMsgID {
		t.Fatalf("queue should not contain delivered %s at %s", resp.MsgID, resp.QueueFilePath)
	}
}
```
