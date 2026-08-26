## Expected

- Exit 0.
- Stdout contains `[MOCK OK]`.
- Session `agent_runner` is `opencode` (config codex skipped).
- No config `notice:` on stderr.

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
	assertContains(t, resp.Stdout, "[MOCK OK]")
	if resp.AgentRunner != "opencode" {
		t.Fatalf("AgentRunner = %q, want opencode\nstderr:\n%s", resp.AgentRunner, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "notice:") {
		t.Fatalf("--no-config must not emit config notice:\n%s", resp.Stderr)
	}
}
```

