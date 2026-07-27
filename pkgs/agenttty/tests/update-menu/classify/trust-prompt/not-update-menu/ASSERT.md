## Expected

1. `IsBlockingMenu` is false (trust is not the update menu).
2. Writable is not ready with reason containing `update available`.
3. Prefer trust-related reason / loading state when trust text is present.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsBlockingMenu {
		t.Fatalf("trust prompt must not be IsBlockingUpdateMenu; selection=%q reason=%q",
			resp.MenuSelection, resp.WritableReason)
	}
	low := strings.ToLower(resp.WritableReason)
	if strings.Contains(low, "update available") || strings.Contains(low, "updateavailable") {
		t.Fatalf("trust prompt writable reason must not be update available: %q", resp.WritableReason)
	}
	// Production should treat trust as non-writable loading, not idle.
	if resp.WritableReady {
		t.Fatalf("trust prompt should not be sendable/idle ready, got ready=true state=%q reason=%q",
			resp.WritableState, resp.WritableReason)
	}
}
```
