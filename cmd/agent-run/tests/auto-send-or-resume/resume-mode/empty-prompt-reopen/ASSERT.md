---
label: e2e
---

## Expected

- Exit code 0 (reopen with fake runner succeeds).
- Must **not** fail with "prompt is required".
- Argv probe contains `--resume` and `runner_session_id`.

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(combined, "prompt is required") ||
		strings.Contains(combined, "prompt required") ||
		strings.Contains(combined, "requires a prompt") {
		t.Fatalf("empty-prompt resume reopen must not require a prompt; got:\n%s", combined)
	}
	assertSuccess(t, resp)

	probe := resp.ArgvProbe
	if strings.TrimSpace(probe) == "" {
		b, rErr := os.ReadFile(req.ArgvProbePath)
		if rErr != nil {
			t.Fatalf("read argv probe %s: %v\nstderr:\n%s", req.ArgvProbePath, rErr, resp.Stderr)
		}
		probe = string(b)
	}
	record := strings.TrimSpace(probe)
	assertContains(t, record, "ARGV_RECORD=")
	assertContains(t, record, "--resume")
	assertContains(t, record, req.RunnerSessionID)
}
```
