## Expected

- Exit code 0.
- Argv probe contains `--resume`, the Grok UUID, and `--fork-session`.
- Argv does **not** contain the slash prompt path as sole fork mechanism
  (optional: may still have followup inject; must have `--fork-session`).
- Meta for new session id exists with parent Grok UUID as runner_session_id.
- Mapped parent session still present.

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	probe, rErr := os.ReadFile(req.ArgvProbePath)
	if rErr != nil {
		t.Fatalf("read argv probe: %v\nstderr:\n%s", rErr, resp.Stderr)
	}
	record := strings.TrimSpace(string(probe))
	assertContains(t, record, "ARGV_RECORD=")
	assertContains(t, record, "--resume")
	assertContains(t, record, req.GrokSessionID)
	assertContains(t, record, "--fork-session")

	childMeta := sessionMetaPath(req.Home, req.SessionID)
	data, err := os.ReadFile(childMeta)
	if err != nil {
		t.Fatalf("child meta missing %s: %v", childMeta, err)
	}
	if !strings.Contains(string(data), req.GrokSessionID) {
		t.Fatalf("child meta missing parent grok id:\n%s", data)
	}
	parentMeta := sessionMetaPath(req.Home, req.MappedSessID)
	if _, err := os.Stat(parentMeta); err != nil {
		t.Fatalf("parent mapped meta should remain: %v", err)
	}
	// New session id must differ from mapped parent.
	if req.SessionID == req.MappedSessID {
		t.Fatal("fork must use a new agent-run session id")
	}
	_ = filepath.Separator
}
```
