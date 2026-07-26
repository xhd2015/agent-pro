---
label: e2e
---

## Expected

- Exit code 0 (resume proceeds like run).
- Argv probe file contains `ARGV_RECORD=` with `--resume` and the bound
  `runner_session_id`.
- Optionally stderr may include grok-tty session lines (not required).

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
