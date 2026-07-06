## Expected

- First stdout line is `msg_<n>` received within 1 second of start.
- Id line matches `msg_<n>` format before max-wait deadline elapses.

## Exit Code

non-zero (busy session times out) or 0 if delivery completes — id timing is the assertion focus

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertMsgIDLine(t, resp.MsgID)
	if resp.IdLineLatency >= time.Second {
		t.Fatalf("expected id line within 1s, latency=%v", resp.IdLineLatency)
	}
}
```