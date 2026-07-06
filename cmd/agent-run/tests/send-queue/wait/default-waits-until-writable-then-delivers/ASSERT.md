## Expected

- Exit code 0 after terminal becomes writable.
- Send blocks longer than the previous 10s writable cap (>10s).
- Message injected when writable.

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
	if resp.SendDuration <= 10*time.Second {
		t.Fatalf("expected default wait to exceed old 10s cap, took %v", resp.SendDuration)
	}
	if !containsString(resp.InjectedMessages, "writable-wait-probe") {
		t.Fatalf("expected message injected, seen=%v", resp.InjectedMessages)
	}
}
```