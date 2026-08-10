## Expected

- Exit code 0 (subcommand help succeeds; not `unknown command: takeover`).
- Stdout documents `takeover` usage including a session-id placeholder.
- Stdout documents `--grok`, `--codex`, `--agent-runner`, color flags
  (`--color` / `--no-color` / `--auto-color`), and `--dry-run`.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

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
	out := resp.Stdout
	assertContains(t, out, "takeover")
	assertContains(t, out, "--grok")
	assertContains(t, out, "--codex")
	assertContains(t, out, "--agent-runner")
	assertContains(t, out, "--color")
	assertContains(t, out, "--no-color")
	assertContains(t, out, "--auto-color")
	assertContains(t, out, "--dry-run")
	// session-id appears as a required positional in usage (various phrasings OK)
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "session") {
		t.Fatalf("takeover --help should document session-id; stdout:\n%s", out)
	}
	assertTrailingNewline(t, out, "takeover --help stdout")
}
```
