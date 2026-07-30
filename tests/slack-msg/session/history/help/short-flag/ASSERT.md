## Expected Output

```
slack-msg session history: print local session message history.

Usage:
  slack-msg session history [options]

Options:
  --session-id ID     Session id (env: SLACK_MSG_SESSION_ID)
  --after-msg-id ID   Only messages after this id
  --limit N           Max messages
  --json              JSON output
  --config PATH       Config file (env: SLACK_MSG_CONFIG)
  -h, --help          Show help
```

## Expected

- Exit code 0.
- Stdout documents `--session-id`, `--after-msg-id`, `--json`.
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
	for _, want := range []string{"--session-id", "SLACK_MSG_SESSION_ID", "--after-msg-id", "--json"} {
		if !strings.Contains(resp.Stdout, want) {
			t.Fatalf("session history help missing %q:\n%s", want, resp.Stdout)
		}
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
slack-msg session history: print local session message history\.

Usage:
  slack-msg session history [options]

Options:
  --session-id ID     Session id \(env: SLACK_MSG_SESSION_ID\)
  --after-msg-id ID   Only messages after this id
  --limit N           Max messages
  --json              JSON output
  --config PATH       Config file \(env: SLACK_MSG_CONFIG\)
  -h, --help          Show help
`)
}
```
