## Expected

- Non-zero exit (failing fake agent).
- Stderr contains `notice: agent-runner=codex (from config)`.
- Stderr contains `starting new codex session` (config applied).
- Does not say `starting new opencode session`.

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
		t.Fatal("expected non-zero exit (failing fake)")
	}
	assertContains(t, resp.Stderr, "notice: agent-runner=codex (from config)")
	assertContains(t, resp.Stderr, "starting new codex session")
	if strings.Contains(resp.Stderr, "starting new opencode session") {
		t.Fatalf("config agent_runner ignored; stderr:\n%s", resp.Stderr)
	}
}
```

