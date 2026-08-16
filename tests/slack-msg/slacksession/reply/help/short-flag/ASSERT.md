## Expected Output

```
slack-msg session reply: post a channel-level reply for a bound Slack session.

Usage:
  slack-msg session reply [options] MESSAGE

Options:
  --session-id ID   Session id (env: SLACK_MSG_SESSION_ID)
  --config PATH     Config file (env: SLACK_MSG_CONFIG)
  --token TOKEN     Bot token override
  -h, --help        Show help
```

## Expected

- Exit code 0.
- Stdout documents `--session-id`, `SLACK_MSG_SESSION_ID`, `--config`, `SLACK_MSG_CONFIG`.
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
	for _, want := range []string{"--session-id", "SLACK_MSG_SESSION_ID", "--config", "SLACK_MSG_CONFIG", "MESSAGE"} {
		if !strings.Contains(resp.Stdout, want) {
			t.Fatalf("session reply help missing %q:\n%s", want, resp.Stdout)
		}
	}
	assert.Output(t, resp.Stdout, `slack-msg session reply: post a channel-level reply for a bound Slack session.

Usage:
  slack-msg session reply [options] MESSAGE

Options:
  --session-id ID   Session id (env: SLACK_MSG_SESSION_ID)
  --config PATH     Config file (env: SLACK_MSG_CONFIG)
  --token TOKEN     Bot token override
  -h, --help        Show help
`)
}
```
