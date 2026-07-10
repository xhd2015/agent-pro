## Expected

- Exit code 0.
- Exactly one session under `sessions/fake-codex/`.
- Id matches `^[a-z0-9][a-z0-9._-]*-\d{8}-\d{6}(-\d+)?$`.
- Base portion is `fix-the-flaky-test` (derived from the prompt).
- `meta.json` `session_id` equals the directory name.

## Exit Code

0

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	id := singleSessionID(t, req.Home, "fake-codex")
	if !autoSessionIDShape.MatchString(id) {
		t.Fatalf("session id %q does not match auto-id shape", id)
	}
	base, ts, _, ok := splitAutoSessionID(id)
	if !ok {
		t.Fatalf("could not split auto session id %q", id)
	}
	if base != "fix-the-flaky-test" {
		t.Fatalf("slug base = %q, want fix-the-flaky-test (id=%q)", base, id)
	}
	if ts == "" {
		t.Fatalf("missing timestamp in id %q", id)
	}
	meta := readSessionMeta(t, req.Home, "fake-codex", id)
	if meta.SessionID != id {
		t.Fatalf("meta.session_id = %q, want %q", meta.SessionID, id)
	}
	// Directory must exist (sanity).
	if _, err := os.Stat(sessionDir(req.Home, "fake-codex", id)); err != nil {
		t.Fatalf("session dir missing: %v", err)
	}
}
```
