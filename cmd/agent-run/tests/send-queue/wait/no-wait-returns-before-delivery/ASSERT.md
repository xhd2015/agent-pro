## Expected

- Exit code 0 in under 1 second.
- Stdout prints message id.
- Message not injected within 400ms; still pending in queue.

## Exit Code

0

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	assertMsgIDLine(t, resp.Stdout)
	if resp.SendDuration >= time.Second {
		t.Fatalf("expected return in <1s, took %v", resp.SendDuration)
	}
	if containsString(resp.InjectedMessages, "no-wait-probe") {
		t.Fatalf("message should not be injected yet, seen=%v", resp.InjectedMessages)
	}
	if !resp.QueueHasMsgID {
		t.Fatalf("queue should still contain %s at %s", resp.MsgID, resp.QueueFilePath)
	}
}
```