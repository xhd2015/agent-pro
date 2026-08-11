## Expected

- AutoSendOrResume succeeds (ModeRun).
- `PublishCount == 0` (empty URL disables publish even when OnTTYStarted fires).
- WarnOutput empty.

## Side Effects

- No network; no publish body.

## Errors

- None.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.APIErrString != "" {
		t.Fatalf("AutoSendOrResume: %s", resp.APIErrString)
	}
	if resp.PublishCount != 0 {
		t.Fatalf("empty URL must not publish; PublishCount=%d", resp.PublishCount)
	}
	if req.Capture.Len() != 0 {
		t.Fatalf("empty URL must not record HTTP; got %d requests", req.Capture.Len())
	}
	if strings.TrimSpace(resp.WarnOutput) != "" {
		t.Fatalf("empty URL must not warn; WarnOutput=%q", resp.WarnOutput)
	}
}
```
