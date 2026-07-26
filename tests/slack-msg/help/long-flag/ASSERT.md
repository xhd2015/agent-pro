## Expected

- Exit code 0.
- Stdout matches same top-level usage as `-h` (includes `channels`, `auth`, `session`, Help topics, `--topic`).
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
	for _, cmd := range []string{"send", "history", "listen", "channels", "auth", "session"} {
		if !strings.Contains(resp.Stdout, cmd) {
			t.Fatalf("help missing command %q:\n%s", cmd, resp.Stdout)
		}
	}
	if !strings.Contains(resp.Stdout, "add-missing-scope") {
		t.Fatalf("help missing topic add-missing-scope:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "--topic") {
		t.Fatalf("help missing --topic usage:\n%s", resp.Stdout)
	}
	// v2: escape balanced [...] so lines match literally as regex; leave () alone.
	assert.Output(t, resp.Stdout, `---
version: 2
---
slack-msg: Slack messaging CLI.

Usage:
  slack-msg <command> \[options\]
  slack-msg --help \[--topic TOPIC\]

Commands:
  send      Post a message via Slack Web API
  history   Fetch conversation history or thread replies
  listen    Socket Mode inbound bridge to agent-run
  channels  List or search workspace channels
  auth      Inspect bot or app token status
  session   Session-bound reply and history

Help topics:
  add-missing-scope  How to grant missing OAuth scopes (e.g. groups:read)

Options:
  -h, --help     Show help
  --topic TOPIC  With --help, show a help topic
`)
}
```
