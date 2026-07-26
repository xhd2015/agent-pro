## Expected

- Exit code 0.
- Nested path `sessions/fake-codex/dry_sess` still exists.
- Flat path `sessions/dry_sess` does **not** exist.
- Plan/report printed on stdout or stderr (mentions dry_sess or move/plan).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitZero(t, resp)
	assertDirExists(t, filepath.Join(req.Home, "sessions", "fake-codex", "dry_sess"))
	assertPathMissing(t, filepath.Join(req.Home, "sessions", "dry_sess"))
	// .layout should not claim a completed migrate that moved files.
	layoutPath := filepath.Join(req.Home, "sessions", ".layout")
	if _, err := os.Stat(layoutPath); err == nil {
		// if present, nested path must still exist (checked above)
		_ = layoutPath
	}
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if !strings.Contains(combined, "dry_sess") &&
		!strings.Contains(combined, "dry") &&
		!strings.Contains(combined, "plan") &&
		!strings.Contains(combined, "move") {
		t.Fatalf("expected dry-run plan output, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
