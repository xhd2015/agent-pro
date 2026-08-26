---
label: e2e
---

## Expected

- Exit 0.
- Help text mentions `--agent-runner` and `opencode`, `codex`, `grok`, `commandcode`.
- Failing fake agent not invoked.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "--agent-runner")
	for _, runner := range []string{"opencode", "codex", "grok", "commandcode"} {
		assertContains(t, combined, runner)
	}
	if strings.Contains(resp.Stderr, "FAKE_AGENT_INVOKED") {
		t.Fatalf("help must not invoke agent:\n%s", resp.Stderr)
	}
}
```
