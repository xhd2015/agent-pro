## Expected Output

```
slack-listen: Slack Socket Mode inbound bridge.

Usage:
  slack-listen listen [options]

Options:
  --bot-token TOKEN
  --app-token TOKEN
  --config PATH
  --channel CHANNEL
  --require-mention
  --no-require-mention
  --allow-from USER_ID
  --session-mode MODE
  --idle-timeout DURATION
  --agent-runner RUNNER
  --agent-runner-config-home PATH
  --reply-prefix TEXT
  --lock-file PATH
  -h, --help
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
version: 2
---
slack-listen: Slack Socket Mode inbound bridge.

Usage:
  slack-listen listen \[options\]

Options:
  --bot-token TOKEN
  --app-token TOKEN
  --config PATH
  --channel CHANNEL
  --require-mention
  --no-require-mention
  --allow-from USER_ID
  --session-mode MODE
  --idle-timeout DURATION
  --agent-runner RUNNER
  --agent-runner-config-home PATH
  --reply-prefix TEXT
  --lock-file PATH
  -h, --help
`)
}
```