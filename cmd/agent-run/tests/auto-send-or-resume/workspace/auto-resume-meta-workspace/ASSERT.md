---
label: e2e
---

## Expected

- Exit code 0.
- Cwd probe equals `meta.workspace` (created-ws), not CLI WorkDir.
- Argv includes `--resume <runner_session_id>` (proves MODE=resume).

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

	cwd := resp.CwdProbe
	if cwd == "" {
		b, rErr := os.ReadFile(req.CwdProbePath)
		if rErr != nil {
			t.Fatalf("read cwd probe %s: %v\nstderr:\n%s", req.CwdProbePath, rErr, resp.Stderr)
		}
		cwd = strings.TrimSpace(string(b))
	}
	want := canonicalPath(t, req.Workspace)
	got := canonicalPath(t, cwd)
	if got != want {
		t.Fatalf("auto→resume child cwd must be meta.workspace\n  got:  %s\n  want: %s\n  cli:  %s\nstderr:\n%s",
			got, want, canonicalPath(t, req.WorkDir), resp.Stderr)
	}

	probe := resp.ArgvProbe
	if probe == "" {
		b, _ := os.ReadFile(req.ArgvProbePath)
		probe = string(b)
	}
	assertContains(t, probe, "--resume")
	assertContains(t, probe, req.RunnerSessionID)
}
```
