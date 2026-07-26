---
label: e2e
---

## Expected

- Exit code non-zero.
- Error mentions 10s timeout and provider reason.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode == 0 { t.Fatal("expected non-zero exit on send timeout") }
	combined := resp.MultiAttachProbe.SendTimeoutReason
	if !strings.Contains(combined, "timed out") && !strings.Contains(combined, "10s") {
		t.Fatalf("expected timeout message, got: %s", combined)
	}
}
```
