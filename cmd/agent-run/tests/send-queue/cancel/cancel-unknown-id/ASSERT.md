---
label: e2e
---

## Expected

- Exit code 1.
- Stderr mentions message not found.

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
	if resp.CancelExitCode != 1 {
		t.Fatalf("expected cancel exit 1, got %d stdout=%q stderr=%q", resp.CancelExitCode, resp.CancelStdout, resp.CancelStderr)
	}
	combined := strings.ToLower(resp.CancelStderr)
	if !strings.Contains(combined, "not found") && !strings.Contains(combined, "msg_9999") {
		t.Fatalf("stderr should mention not found, got: %s", resp.CancelStderr)
	}
}
```