## Expected

- `PublishCount` is 0 (no HTTP / no inject Publish).
- WarnOutput is empty (no failure warning).
- `err` from Run is nil.

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
	if resp.PublishCount != 0 {
		t.Fatalf("empty URL must not publish; PublishCount=%d Capture.Len=%d", resp.PublishCount, req.Capture.Len())
	}
	if req.Capture.Len() != 0 {
		t.Fatalf("empty URL must not record HTTP; got %d requests", req.Capture.Len())
	}
	if strings.TrimSpace(resp.WarnOutput) != "" {
		t.Fatalf("empty URL must not warn; WarnOutput=%q", resp.WarnOutput)
	}
}
```
