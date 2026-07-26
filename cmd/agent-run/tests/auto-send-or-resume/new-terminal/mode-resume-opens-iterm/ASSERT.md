---
label: e2e
---

## Expected

- Exit code 0.
- iTerm ForceNew script written (`create window`, no `create tab`).
- Follow-up has `--auto-send-or-resume`, session-id, no `--new-terminal`.
- Parent argv probe empty/absent (no in-process resume spawn with `--resume`).

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

	script := readItermScript(t, req)
	assertItermForceNewScript(t, script)
	assertContains(t, script, "--auto-send-or-resume")
	assertContains(t, script, req.SessionID)
	assertContains(t, script, req.FollowupPrompt)
	if strings.Contains(script, "--new-terminal") {
		t.Fatalf("follow-up must strip --new-terminal; script:\n%s", script)
	}
	assertContainsAny(t, script, "agent-run", req.AgentRun)

	assertNoInProcessProviderSpawn(t, req, resp)
	// Explicit: no --resume in parent probe (child iTerm would resume).
	if fileExists(req.ArgvProbePath) {
		probe, rErr := os.ReadFile(req.ArgvProbePath)
		if rErr == nil && strings.Contains(string(probe), "--resume") {
			t.Fatalf("parent must not spawn resume provider; argv:\n%s", probe)
		}
	}
}
```
