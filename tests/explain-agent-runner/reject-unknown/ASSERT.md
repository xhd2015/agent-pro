---
label: e2e
---

## Expected

- Non-zero exit.
- Stderr contains `Error:` and `unsupported agent runner`.
- Mentions supported runners.
- Agent stub not invoked.

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "Error:")
	assertContains(t, resp.Stderr, "unsupported agent runner")
	for _, runner := range []string{"opencode", "codex", "grok", "commandcode"} {
		if !strings.Contains(resp.Stderr, runner) {
			t.Fatalf("stderr should list %q:\n%s", runner, resp.Stderr)
		}
	}
	if strings.Contains(resp.Stderr, "FAKE_AGENT_INVOKED") {
		t.Fatalf("reject must not invoke agent:\n%s", resp.Stderr)
	}
}
```
