## Expected

- Exit code 0.
- Stdout matches same top-level usage as `-h`.
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
	for _, cmd := range []string{"send", "history", "listen"} {
		if !strings.Contains(resp.Stdout, cmd) {
			t.Fatalf("help missing command %q:\n%s", cmd, resp.Stdout)
		}
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
slack-msg: Slack messaging CLI.

Usage:
  slack-msg <command> [options]

Commands:
  send     Post a message via Slack Web API
  history  Fetch conversation history or thread replies
  listen   Socket Mode inbound bridge to agent-run

Options:
  -h, --help  Show help
`)
}
```
