## Expected Output

```
slack-msg history: fetch conversation history or thread replies.

Usage:
  slack-msg history [options] [CHANNEL]

Options:
  --token TOKEN     Bot token (env: SLACK_BOT_TOKEN)
  --channel CHANNEL Channel ID or name (env: SLACK_CHANNEL)
  --config PATH     JSON config file (env: SLACK_CONFIG)
  --limit N         Max messages to fetch
  --thread TS       Fetch thread replies for TS
  --json            Structured JSON output
  -h, --help        Show help
```

## Expected

- Exit code 0.
- Stdout matches usage (trailing newline).
- Stderr empty.

## Exit Code

0

```go
import (
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
	assert.Output(t, resp.Stdout, `---
version: 2
---
slack-msg history: fetch conversation history or thread replies.

Usage:
  slack-msg history [options] [CHANNEL]

Options:
  --token TOKEN     Bot token (env: SLACK_BOT_TOKEN)
  --channel CHANNEL Channel ID or name (env: SLACK_CHANNEL)
  --config PATH     JSON config file (env: SLACK_CONFIG)
  --limit N         Max messages to fetch
  --thread TS       Fetch thread replies for TS
  --json            Structured JSON output
  -h, --help        Show help
`)
}
```
