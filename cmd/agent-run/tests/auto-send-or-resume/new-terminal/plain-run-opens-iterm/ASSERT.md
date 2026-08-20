---
label: e2e
---

## Expected

- Exit code 0.
- iTerm ForceNew script written (`create window`, no `create tab`).
- Follow-up contains `run`, the prompt after `--`, and `--agent-runner-binary`;
  does **not** contain `--new-terminal` or `--auto-send-or-resume`.
- No in-process provider spawn.

## Exit Code

0

```go
import (
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
	assertContains(t, script, "run")
	assertContains(t, script, req.FollowupPrompt)
	assertContains(t, script, "--agent-runner-binary")
	if strings.Contains(script, "--new-terminal") {
		t.Fatalf("follow-up must strip --new-terminal; script:\n%s", script)
	}
	if strings.Contains(script, "--auto-send-or-resume") {
		t.Fatalf("plain --new-terminal must not inject --auto-send-or-resume; script:\n%s", script)
	}
	assertContainsAny(t, script, " -- ", `"--"`, `'--'`)
	assertNoInProcessProviderSpawn(t, req, resp)
}
```
