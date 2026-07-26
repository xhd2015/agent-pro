---
label: e2e
---

## Expected

- `msg status` before cancel prints `pending`.
- Cancel exit code 0 (silent success).
- `msg status` after cancel prints `delivered`.
- Message never injected into terminal.
- Queue no longer contains cancelled message id.

## Exit Code

0 (cancel command)

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	if strings.TrimSpace(resp.StatusBeforeStdout) != "pending" {
		t.Fatalf("status before cancel want pending, got %q", resp.StatusBeforeStdout)
	}
	if resp.CancelExitCode != 0 {
		t.Fatalf("cancel expected exit 0, got %d stderr=%s", resp.CancelExitCode, resp.CancelStderr)
	}
	if strings.TrimSpace(resp.StatusAfterStdout) != "delivered" {
		t.Fatalf("status after cancel want delivered, got %q", resp.StatusAfterStdout)
	}
	if containsString(resp.InjectedMessages, "cancel-probe") {
		t.Fatalf("cancelled message should not be injected, seen=%v", resp.InjectedMessages)
	}
	if resp.QueueHasMsgID {
		t.Fatalf("queue should not contain cancelled message at %s", resp.QueueFilePath)
	}
}
```