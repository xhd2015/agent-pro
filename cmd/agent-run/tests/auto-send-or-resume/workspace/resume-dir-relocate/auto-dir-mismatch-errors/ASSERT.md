---
label: e2e
---

## Expected

- Exit code **1** on auto→resume when `--dir` ≠ Grok `info.cwd` and allow flag absent.
- Stderr includes **`--allow-relocate-resume-session-dir`**.
- Stderr explains mismatch / cwd / relocate (same contract as `dir-mismatch-errors`).
- Grok session remains under `encode(ws-old)`; not under `encode(ws-new)`.
- `info.cwd` unchanged.

## Errors

- Same user-facing mismatch contract as resume subcommand leaf.

## Exit Code

1

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)

	combined := resp.Stderr + "\n" + resp.Stdout
	lower := strings.ToLower(combined)

	assertContains(t, combined, "--allow-relocate-resume-session-dir")
	assertContainsAny(t, lower,
		"mismatch",
		"does not match",
		"differ",
		"different",
		"cwd",
		"working directory",
		"session dir",
		"session directory",
		"relocate",
	)

	oldHit := strings.Contains(combined, req.GrokSessionCwd) ||
		strings.Contains(combined, filepath.Base(req.GrokSessionCwd)) ||
		strings.Contains(combined, "ws-old")
	newHit := strings.Contains(combined, absPathNoEval(t, req.DirOverride)) ||
		strings.Contains(combined, req.DirOverride) ||
		strings.Contains(combined, filepath.Base(req.DirOverride)) ||
		strings.Contains(combined, "ws-new")
	if !oldHit && !newHit {
		t.Fatalf("auto→resume error should mention old and/or new dir paths\ncombined:\n%s", combined)
	}

	oldDir := grokSessionDirAt(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)
	newDir := grokSessionDirAt(t, req.GrokHome, req.DirOverride, req.RunnerSessionID)
	if !pathExists(oldDir) {
		t.Fatalf("expected grok session still under old cwd key %s", oldDir)
	}
	if pathExists(newDir) {
		t.Fatalf("session must not move without allow flag: %s", newDir)
	}
	gotCWD := summaryInfoCWD(t, filepath.Join(oldDir, "summary.json"))
	if absPathNoEval(t, gotCWD) != absPathNoEval(t, req.GrokSessionCwd) {
		t.Fatalf("info.cwd changed without allow: got %q want %q", gotCWD, req.GrokSessionCwd)
	}
}
```
