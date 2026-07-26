---
label: e2e
---

## Expected

- Exit code 1.
- Stderr mentions message id and timeout duration.
- Queue file does not contain the timed-out message id.

## Exit Code

1

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertExitCode(t, resp, 1)
	assertMsgIDLine(t, resp.MsgID)
	combined := resp.Stderr + resp.Stdout
	if !strings.Contains(combined, resp.MsgID) {
		t.Fatalf("stderr should mention %s, got stderr=%q", resp.MsgID, resp.Stderr)
	}
	if !strings.Contains(combined, "2s") && !strings.Contains(combined, "within") {
		t.Fatalf("stderr should mention timeout, got: %s", resp.Stderr)
	}
	if resp.QueueHasMsgID {
		t.Fatalf("queue should not contain timed-out %s at %s", resp.MsgID, resp.QueueFilePath)
	}
}
```