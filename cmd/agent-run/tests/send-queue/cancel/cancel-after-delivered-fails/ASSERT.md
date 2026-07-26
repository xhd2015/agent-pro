---
label: e2e
---

## Expected

- Initial send exit code 0 with msg id on stdout.
- Cancel same id exit code 1 with stderr error.

## Exit Code

1 (cancel command)

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertSuccess(t, resp)
	assertMsgIDLine(t, resp.Stdout)
	if resp.CancelExitCode != 1 {
		t.Fatalf("expected cancel exit 1 after delivery, got %d stderr=%s", resp.CancelExitCode, resp.CancelStderr)
	}
	if resp.CancelStderr == "" {
		t.Fatal("expected stderr error on cancel after delivery")
	}
}
```