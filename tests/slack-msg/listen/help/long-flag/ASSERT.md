## Expected

- Exit code 0.
- Stdout matches same usage as `listen -h`.
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
	if strings.Contains(resp.Stdout, "--bot-token") {
		t.Fatalf("help must use --token, not --bot-token:\n%s", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
slack-msg listen: Slack Socket Mode inbound bridge.

Usage:
  slack-msg listen [options]

Options:
  --token TOKEN
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
  --no-lock
  -h, --help
`)
}
```
