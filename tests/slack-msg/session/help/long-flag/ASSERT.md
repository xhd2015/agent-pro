## Expected

- Exit code 0.
- Stdout matches same session usage as `-h` (lists `reply`, `history`).
- Stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	for _, cmd := range []string{"reply", "history"} {
		if !strings.Contains(resp.Stdout, cmd) {
			t.Fatalf("session help missing %q:\n%s", cmd, resp.Stdout)
		}
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
slack-msg session: session-bound reply and history.

Usage:
  slack-msg session <command> [options]

Commands:
  reply    Post a channel reply for the bound session
  history  Show local session message history

Options:
  -h, --help  Show help
`)
}
```
