## Expected Output

```
slack-msg session: session-bound management, reply and history.

Usage:
  slack-msg session <command> [options]

Commands:
  list     List sessions from the local map
  info     Show details for one session
  update   Update session fields (e.g. workspace dir)
  reply    Post a channel reply for the bound session
  history  Show local session message history

Options:
  -h, --help  Show help
```

## Expected

- Exit code 0.
- Stdout lists `list`, `info`, `update`, `reply`, and `history`.
- Stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	for _, cmd := range []string{"list", "info", "update", "reply", "history"} {
		if !strings.Contains(resp.Stdout, cmd) {
			t.Fatalf("session help missing %q:\n%s", cmd, resp.Stdout)
		}
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
slack-msg session: session-bound management, reply and history.

Usage:
  slack-msg session <command> [options]

Commands:
  list     List sessions from the local map
  info     Show details for one session
  update   Update session fields (e.g. workspace dir)
  reply    Post a channel reply for the bound session
  history  Show local session message history

Options:
  -h, --help  Show help
`)
}
```
