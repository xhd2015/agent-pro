---
label: e2e
---

## Expected

- Exit code 0.
- Grok session directory still at `sessions/<encode(ws-match)>/<runner_session_id>/`
  (not moved to a different encoded key).
- `summary.json` `info.cwd` still equals the match workspace (Abs).
- Fixture marker still present under the original session dir.
- Argv includes `--resume <runner_session_id>`.
- Stderr must **not** require a relocate warning (optional: must not claim a relocate happened).

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

	// Session must remain under original encoded cwd key.
	wantDir := grokSessionDirAt(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)
	if !pathExists(wantDir) {
		t.Fatalf("expected grok session still at %s (no relocate when --dir matches)\nstderr:\n%s",
			wantDir, resp.Stderr)
	}
	marker := filepath.Join(wantDir, "fixture-marker.txt")
	if !pathExists(marker) {
		t.Fatalf("fixture marker missing at %s — session may have been recreated", marker)
	}
	gotCWD := summaryInfoCWD(t, filepath.Join(wantDir, "summary.json"))
	if absPathNoEval(t, gotCWD) != absPathNoEval(t, req.GrokSessionCwd) {
		t.Fatalf("info.cwd = %q, want %q", gotCWD, req.GrokSessionCwd)
	}

	// Must not have been moved under a different key accidentally.
	// (Same path is OK; if DirOverride equals GrokSessionCwd there is only one key.)
	probe := resp.ArgvProbe
	if probe == "" {
		b, _ := os.ReadFile(req.ArgvProbePath)
		probe = string(b)
	}
	assertContains(t, probe, "--resume")
	assertContains(t, probe, req.RunnerSessionID)

	// Child cwd should be --dir (match workspace).
	cwd := resp.CwdProbe
	if cwd == "" {
		if b, rErr := os.ReadFile(req.CwdProbePath); rErr == nil {
			cwd = strings.TrimSpace(string(b))
		}
	}
	if cwd != "" {
		if canonicalPath(t, cwd) != canonicalPath(t, req.DirOverride) {
			t.Fatalf("child cwd = %s, want --dir %s\nstderr:\n%s",
				cwd, req.DirOverride, resp.Stderr)
		}
	}
}
```
