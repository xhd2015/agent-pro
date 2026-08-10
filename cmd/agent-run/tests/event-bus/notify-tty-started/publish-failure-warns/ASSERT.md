## Expected

- Run `err` is nil (NotifyTTYStarted never fails the open path).
- WarnOutput contains `warning:` (case-sensitive prefix or substring).
- Caller is not panicked.

## Side Effects

- Best-effort: may or may not record a failed attempt; count not required.
- WarnWriter receives a human-readable warning line.

## Errors

- None returned to Run/caller.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("Run must not fail on publish error: %v\nWarnOutput=%q", err, resp.WarnOutput)
	}
	if !strings.Contains(resp.WarnOutput, "warning:") {
		t.Fatalf("publish failure must write warning: to WarnWriter; got %q", resp.WarnOutput)
	}
}
```
