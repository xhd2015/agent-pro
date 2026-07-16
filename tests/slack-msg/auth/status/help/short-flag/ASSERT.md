## Expected Output

```
slack-msg auth status: show bot or app token status.

Usage:
  slack-msg auth status [options]
  slack-msg auth status --app [options]

Options:
  --token TOKEN       Bot token (env: SLACK_BOT_TOKEN)
  --app-token TOKEN   App-level token (env: SLACK_APP_TOKEN)
  --config PATH       JSON config file (env: SLACK_CONFIG)
  --json              Structured JSON output
  --app               Validate app-level token (Socket Mode / connections)
  -h, --help          Show help
```

## Expected

- Exit code 0.
- Stdout matches usage including `--app` (trailing newline).
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
slack-msg auth status: show bot or app token status\.

Usage:
  slack-msg auth status \[options\]
  slack-msg auth status --app \[options\]

Options:
  --token TOKEN       Bot token \(env: SLACK_BOT_TOKEN\)
  --app-token TOKEN   App-level token \(env: SLACK_APP_TOKEN\)
  --config PATH       JSON config file \(env: SLACK_CONFIG\)
  --json              Structured JSON output
  --app               Validate app-level token \(Socket Mode / connections\)
  -h, --help          Show help
`)
}
```
