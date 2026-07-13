## Expected

- Exit code 0.
- `meta.json` created for the session-id.
- `meta.workspace` non-empty (create-time cwd / WorkDir).
- Argv probe contains the prompt text and does **not** contain `--resume`.

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	if resp.MetaAfter == nil {
		// Fall back to direct read if Mode enrichment missed.
		path := metaJSONPath(req.Home, req.SessionID)
		if !fileExists(path) {
			t.Fatalf("expected meta.json created at %s\nstderr:\n%s\nstdout:\n%s", path, resp.Stderr, resp.Stdout)
		}
		resp.MetaAfter = readMetaJSON(t, req.Home, req.SessionID)
	}
	ws, _ := resp.MetaAfter["workspace"].(string)
	if strings.TrimSpace(ws) == "" {
		t.Fatalf("meta.workspace must be set on create; meta=%v", resp.MetaAfter)
	}
	// Workspace should resolve to create-time WorkDir (TempDir).
	wantWS := canonicalPath(t, req.WorkDir)
	gotWS := canonicalPath(t, ws)
	if gotWS != wantWS && !strings.HasPrefix(gotWS, wantWS) && !strings.HasPrefix(wantWS, gotWS) {
		// Allow equal basenames under temp; require non-empty match of TempDir.
		if !strings.Contains(gotWS, filepath.Base(req.TempDir)) && gotWS != wantWS {
			t.Logf("meta.workspace=%q want near %q (soft check; non-empty is required)", ws, req.WorkDir)
		}
	}

	probe, rErr := os.ReadFile(req.ArgvProbePath)
	if rErr != nil {
		t.Fatalf("read argv probe %s: %v\nstderr:\n%s\nstdout:\n%s", req.ArgvProbePath, rErr, resp.Stderr, resp.Stdout)
	}
	record := strings.TrimSpace(string(probe))
	assertContains(t, record, "ARGV_RECORD=")
	assertContains(t, record, req.FollowupPrompt)
	if strings.Contains(record, "--resume") {
		t.Fatalf("MODE=run must not pass provider --resume; argv:\n%s", record)
	}
}
```
