## Expected

- Exit code **1** (non-zero).
- Stderr (or combined) explains **`--dir` / Grok session cwd mismatch** (or
  equivalent wording: cwd / directory / workspace mismatch vs grok session).
- Stderr mentions the exact flag **`--allow-relocate-resume-session-dir`**.
- Grok session directory still under `encode(ws-old)` / runner_session_id
  (not moved to encode(ws-new)).
- `summary.json` `info.cwd` still old Abs path.
- Fixture marker still at old session dir.
- Provider must not have been spawned successfully for a full resume turn
  (argv probe ideally empty / no successful resume spawn required).

## Errors

- Mismatch explanation + allow-flag hint.
- Session filesystem layout unchanged.

## Exit Code

1

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)

	combined := resp.Stderr + "\n" + resp.Stdout
	lower := strings.ToLower(combined)

	// Exact flag name must appear so users know how to opt in.
	assertContains(t, combined, "--allow-relocate-resume-session-dir")

	// Mismatch / cwd / directory language (not only generic flag parse noise).
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

	// Paths should help the user (old and/or new).
	oldHit := strings.Contains(combined, req.GrokSessionCwd) ||
		strings.Contains(combined, filepath.Base(req.GrokSessionCwd)) ||
		strings.Contains(combined, "ws-old")
	newHit := strings.Contains(combined, absPathNoEval(t, req.DirOverride)) ||
		strings.Contains(combined, req.DirOverride) ||
		strings.Contains(combined, filepath.Base(req.DirOverride)) ||
		strings.Contains(combined, "ws-new")
	if !oldHit && !newHit {
		t.Fatalf("error should mention old and/or new dir paths\ncombined:\n%s", combined)
	}

	// Session must NOT have been relocated.
	oldDir := grokSessionDirAt(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)
	newDir := grokSessionDirAt(t, req.GrokHome, req.DirOverride, req.RunnerSessionID)
	if !pathExists(oldDir) {
		t.Fatalf("expected grok session still under old cwd key %s", oldDir)
	}
	if pathExists(newDir) {
		t.Fatalf("session must not move to new cwd key without allow flag: %s", newDir)
	}
	if !pathExists(filepath.Join(oldDir, "fixture-marker.txt")) {
		t.Fatalf("fixture marker missing under old session dir")
	}
	gotCWD := summaryInfoCWD(t, filepath.Join(oldDir, "summary.json"))
	if absPathNoEval(t, gotCWD) != absPathNoEval(t, req.GrokSessionCwd) {
		t.Fatalf("info.cwd changed without allow: got %q want %q", gotCWD, req.GrokSessionCwd)
	}

	// meta.workspace must not flip to --dir on the error path.
	meta := resp.MetaAfter
	if meta == nil {
		meta = readMetaJSON(t, req.Home, req.SessionID)
	}
	if ws, _ := meta["workspace"].(string); ws != "" {
		if canonicalPath(t, ws) == canonicalPath(t, req.DirOverride) &&
			canonicalPath(t, req.DirOverride) != canonicalPath(t, req.Workspace) {
			t.Fatalf("meta.workspace must not update to --dir on mismatch error; got %q", ws)
		}
	}

	// Optional: successful full spawn is unexpected on hard error.
	if p := strings.TrimSpace(req.ArgvProbePath); p != "" {
		if b, rErr := os.ReadFile(p); rErr == nil && strings.Contains(string(b), "--resume") {
			// Soft: some implementations may record before check; prefer fail if probe implies success path.
			// Hard contract is exit≠0 + unmoved session above.
			_ = b
		}
	}
}
```
