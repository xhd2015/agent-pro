## Expected

1. `IsBlockingMenu` is false.
2. `MenuSelection` is empty (not `UPDATE_NOW` / `SKIP`).
3. `WritableReason` must not equal/contain `update available` as the gate reason.
4. With `model: loading`, non-ready is OK if reason mentions **model**.

## Errors

- `IsBlockingMenu=true` (banner misclassified as modal).
- `WritableReason` contains `update available` (post-Skip hang).

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
		t.Fatalf("IsBlockingMenu=true for residual banner (fixture %s) — menu options are gone",
			req.FixtureFile)
	}
	if resp.MenuSelection == "UPDATE_NOW" || resp.MenuSelection == "SKIP" {
		t.Fatalf("MenuSelection=%q on non-menu banner, want empty", resp.MenuSelection)
	}
	reason := strings.ToLower(strings.TrimSpace(resp.WritableReason))
	if reason == "codex update available" || strings.Contains(reason, "update available") {
		t.Fatalf("writable reason=%q: residual banner must not use update-available gate",
			resp.WritableReason)
	}
	if !resp.WritableReady {
		if resp.WritableState == "loading" && !strings.Contains(reason, "model") {
			t.Fatalf("non-ready residual banner should be model loading, got state=%q reason=%q",
				resp.WritableState, resp.WritableReason)
		}
	}
}
```
