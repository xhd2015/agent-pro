---
label: e2e
---

## Expected

- Exit code 0.
- Stderr includes a **warning** about relocating / cwd mismatch / moving the
  Grok session (not a hard error).
- Old Grok session dir under `encode(ws-old)` is gone (or empty / missing).
- New session dir exists at `encode(ws-new)/<runner_session_id>/`.
- Fixture marker moved with the session (present under new dir).
- `summary.json` `info.cwd` equals Abs(ws-new).
- agent-run `meta.workspace` updated to `--dir` (canonical match).
- Argv includes `--resume <runner_session_id>` (resume continued).
- Child cwd equals `--dir` when probe present.

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

	combined := resp.Stderr + "\n" + resp.Stdout
	lower := strings.ToLower(combined)
	// Warning-style language on stderr about relocate / mismatch.
	assertContainsAny(t, lower,
		"warn",
		"relocat",
		"moving",
		"moved",
		"mismatch",
		"cwd",
	)

	oldDir := grokSessionDirAt(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)
	newDir := grokSessionDirAt(t, req.GrokHome, req.DirOverride, req.RunnerSessionID)

	if pathExists(oldDir) {
		// Allow empty parent leftover, but session id dir must not remain.
		t.Fatalf("old grok session dir must be gone after relocate: %s", oldDir)
	}
	if !pathExists(newDir) {
		t.Fatalf("expected relocated session at %s\nstderr:\n%s", newDir, resp.Stderr)
	}
	marker := filepath.Join(newDir, "fixture-marker.txt")
	if !pathExists(marker) {
		t.Fatalf("fixture marker must move with session to %s", marker)
	}

	wantCWD := absPathNoEval(t, req.DirOverride)
	gotCWD := summaryInfoCWD(t, filepath.Join(newDir, "summary.json"))
	if absPathNoEval(t, gotCWD) != wantCWD {
		t.Fatalf("info.cwd = %q, want %q", gotCWD, wantCWD)
	}

	meta := resp.MetaAfter
	if meta == nil {
		meta = readMetaJSON(t, req.Home, req.SessionID)
	}
	ws, _ := meta["workspace"].(string)
	if ws == "" {
		t.Fatalf("meta.workspace empty after allow relocate; meta=%v", meta)
	}
	if canonicalPath(t, ws) != canonicalPath(t, req.DirOverride) {
		t.Fatalf("meta.workspace = %q, want --dir %q", ws, req.DirOverride)
	}

	probe := resp.ArgvProbe
	if probe == "" {
		b, _ := os.ReadFile(req.ArgvProbePath)
		probe = string(b)
	}
	assertContains(t, probe, "--resume")
	assertContains(t, probe, req.RunnerSessionID)

	cwd := resp.CwdProbe
	if cwd == "" {
		if b, rErr := os.ReadFile(req.CwdProbePath); rErr == nil {
			cwd = strings.TrimSpace(string(b))
		}
	}
	if cwd != "" && canonicalPath(t, cwd) != canonicalPath(t, req.DirOverride) {
		t.Fatalf("child cwd = %s, want --dir %s", cwd, req.DirOverride)
	}
}
```
