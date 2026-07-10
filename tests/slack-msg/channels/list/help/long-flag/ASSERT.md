## Expected

- Exit code 0.
- Stdout matches same usage as `channels list -h`.
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
slack-msg channels list: list workspace channels.

Usage:
  slack-msg channels list [options]

Options:
  --token TOKEN   Bot token (env: SLACK_BOT_TOKEN)
  --config PATH   JSON config file (env: SLACK_CONFIG)
  --types TYPES   Channel types (default: public,private)
  --limit N       Max channels to print
  --json          Structured JSON output
  -h, --help      Show help
`)
}
```
