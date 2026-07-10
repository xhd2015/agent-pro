## Expected Output

```
slack-send: send a message via Slack bot using extracted config.

Usage:
  go run ./script/debug/slack-send
  go run ./script/debug/slack-send "#general"
  go run ./script/debug/slack-send "#general" "Hello from Go debug script"
  go run ./script/debug/slack-send C0ALE44K5J6 "custom message here"

Config file (slack-config.json) is searched starting from cwd upward until go.mod is found.

Defaults:
  channel: C0ALE44K5J6 (#general)
  text: "Hello slack"
```

## Expected

- Exit code 0.
- Stdout matches usage block exactly (trailing newline).
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
slack-send: send a message via Slack bot using extracted config.

Usage:
  go run ./script/debug/slack-send
  go run ./script/debug/slack-send "#general"
  go run ./script/debug/slack-send "#general" "Hello from Go debug script"
  go run ./script/debug/slack-send C0ALE44K5J6 "custom message here"

Config file (slack-config.json) is searched starting from cwd upward until go.mod is found.

Defaults:
  channel: C0ALE44K5J6 (#general)
  text: "Hello slack"
`)
}
```