## Expected

- Exit code 0 from `--no-wait` send.
- Stdout is a single `msg_<n>\n` line.
- CLI returns in under 1 second (no blocking on writable).
- After CLI exit (no blocking send), message appears in terminal scrollback.
- `msg status` reports `delivered`.
- Queue file no longer contains the message id.

## Side Effects

- Session-owned TTY drainer dequeues and injects without any follow-up CLI send.
- No dependency on CLI-elected `StartDrainer`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	assertMsgIDLine(t, resp.Stdout)
	if resp.SendDuration >= time.Second {
		t.Fatalf("expected --no-wait return in <1s (CLI must not block on deliver), took %v", resp.SendDuration)
	}
	if !containsString(resp.InjectedMessages, "tty-drainer-idle-probe") {
		t.Fatalf("expected tty-drainer-idle-probe injected after CLI exit without blocking send, seen=%v", resp.InjectedMessages)
	}
	if strings.TrimSpace(resp.StatusAfterStdout) != "delivered" {
		t.Fatalf("status after TTY-side delivery want delivered, got %q", resp.StatusAfterStdout)
	}
	if resp.QueueHasMsgID {
		t.Fatalf("queue should not contain delivered %s at %s", resp.MsgID, resp.QueueFilePath)
	}
}
```
