---
label: e2e
---

## Expected

- Exit code 0.
- iTerm ForceNew script.
- Follow-up keeps `--open`, `--session-id-from-prompt`, `--agent-runner-binary`,
  `--agent-runner-config-home`, `LLM_MOCK_MCP`, and the prompt after `--`.
- Follow-up does **not** contain `--new-terminal`.
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
	assertContains(t, script, "--open")
	assertContains(t, script, "--session-id-from-prompt")
	assertContains(t, script, "--agent-runner-binary")
	assertContains(t, script, req.AgentRunnerBinary)
	assertContains(t, script, "--agent-runner-config-home")
	assertContains(t, script, "LLM_MOCK_MCP=slow_01=1s-10s")
	assertContains(t, script, req.FollowupPrompt)
	if strings.Contains(script, "--new-terminal") {
		t.Fatalf("follow-up must strip --new-terminal; script:\n%s", script)
	}
	if strings.Contains(script, "--auto-send-or-resume") {
		t.Fatalf("must not inject --auto-send-or-resume; script:\n%s", script)
	}
	assertNoInProcessProviderSpawn(t, req, resp)
}
```
