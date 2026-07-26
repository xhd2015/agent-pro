---
label: e2e
---

## Expected

- Exit code 0.
- Storage directory `sessions/fake-codex/my-task/` exists.
- `meta.session_id` is `my-task`.

## Exit Code

0

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	const want = "my-task"
	dir := sessionDir(req.Home, "fake-codex", want)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected storage session dir %s: %v\nids=%v", dir, err, listSessionIDs(t, req.Home, "fake-codex"))
	}
	meta := readSessionMeta(t, req.Home, "fake-codex", want)
	if meta.SessionID != want {
		t.Fatalf("meta.session_id = %q, want %q", meta.SessionID, want)
	}
}
```
