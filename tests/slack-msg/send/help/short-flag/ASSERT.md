## Expected Output

```
slack-msg send: post a message via Slack Web API.

Usage:
  slack-msg send [options] MESSAGE

Options:
  --token TOKEN     Bot token (env: SLACK_BOT_TOKEN)
  --channel CHANNEL Channel ID or name (env: SLACK_CHANNEL)
  --config PATH     JSON config file (env: SLACK_CONFIG)
  --thread TS       Optional thread timestamp
  -h, --help        Show help
```

## Expected

- Exit code 0.
- Stdout matches usage block (trailing newline).
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
version: 3
---
slack-msg send: post a message via Slack Web API\.

Usage:
  slack-msg send \[options\] MESSAGE

Options:
  --token TOKEN     Bot token \(env: SLACK_BOT_TOKEN\)
  --channel CHANNEL Channel ID or name \(env: SLACK_CHANNEL\)
  --config PATH     JSON config file \(env: SLACK_CONFIG\)
  --thread TS       Optional thread timestamp
  -h, --help        Show help
`)
}
```
