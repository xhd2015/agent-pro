## Expected

- Must **not** fail with `already in use`.
- Exit code 0 after reclaim + headless run.
- Argv probe contains `--resume` and the bound `runner_session_id`.
- `meta.terminal_session_id` non-empty (same id preferred).

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
	combined := resp.Stderr + "\n" + resp.Stdout
	low := strings.ToLower(combined)
	if strings.Contains(low, "already in use") {
		t.Fatalf("headless resume with zombie registry must reclaim/reuse terminal id, not fail with already-in-use:\n%s", combined)
	}
	if strings.Contains(low, "cannot resume") && (strings.Contains(low, "not exited") || strings.Contains(low, "still active")) {
		t.Fatalf("fixture should be zombie-exited (resume ready), not live gate deny:\n%s", combined)
	}
	assertSuccess(t, resp)
	probe, rErr := os.ReadFile(req.ArgvProbePath)
	if rErr != nil {
		t.Fatalf("read argv probe %s: %v\nstderr:\n%s\nstdout:\n%s", req.ArgvProbePath, rErr, resp.Stderr, resp.Stdout)
	}
	record := strings.TrimSpace(string(probe))
	assertContains(t, record, "ARGV_RECORD=")
	assertContains(t, record, "--resume")
	assertContains(t, record, req.RunnerSessionID)
}
```
