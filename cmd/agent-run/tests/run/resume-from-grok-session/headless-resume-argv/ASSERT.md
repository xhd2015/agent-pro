## Expected

- Exit code 0 (headless import launch succeeds with fake binary).
- Argv probe file contains `ARGV_RECORD=` with `--resume` and the Grok UUID.
- Soft: meta for `--session-id` has `runner_session_id` equal to that UUID
  (hard meta shape is covered by `creates-session-meta`).

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	probe, rErr := os.ReadFile(req.ArgvProbePath)
	if rErr != nil {
		t.Fatalf("read argv probe %s: %v\nstderr:\n%s\nstdout:\n%s",
			req.ArgvProbePath, rErr, resp.Stderr, resp.Stdout)
	}
	record := strings.TrimSpace(string(probe))
	assertContains(t, record, "ARGV_RECORD=")
	assertContains(t, record, "--resume")
	assertContains(t, record, req.GrokSessionID)

	// Soft meta check: pre-bind happened (not a substitute for creates-session-meta).
	metaPath := sessionMetaPath(req.Home, req.SessionID)
	if data, err := os.ReadFile(metaPath); err == nil {
		s := string(data)
		if !strings.Contains(s, req.GrokSessionID) {
			t.Fatalf("meta %s missing runner_session_id %q:\n%s", metaPath, req.GrokSessionID, s)
		}
	}
}
```
