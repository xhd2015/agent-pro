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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	assertMsgIDLine(t, resp.Stdout)
	if resp.SendDuration >= time.Second {
		t.Fatalf("expected --no-wait return in <1s, took %v", resp.SendDuration)
	}
}
```