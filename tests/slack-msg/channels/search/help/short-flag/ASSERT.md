## Expected Output

```
slack-msg channels search: search workspace channels by name.

Usage:
  slack-msg channels search [options] QUERY

Options:
  --token TOKEN   Bot token (env: SLACK_BOT_TOKEN)
  --config PATH   JSON config file (env: SLACK_CONFIG)
  --types TYPES   Channel types (default: public,private)
  --limit N       Max channels to print
  --exact         Match channel name exactly
  --prefix        Match channel name by prefix
  --json          Structured JSON output
  -h, --help      Show help
```

## Expected

- Exit code 0.
- Stdout matches usage (trailing newline); plain `[options]`.
- Stderr empty.

## Exit Code

0

```go
import (
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
	assert.Output(t, resp.Stdout, `---
version: 2
---
slack-msg channels search: search workspace channels by name.

Usage:
  slack-msg channels search [options] QUERY

Options:
  --token TOKEN   Bot token (env: SLACK_BOT_TOKEN)
  --config PATH   JSON config file (env: SLACK_CONFIG)
  --types TYPES   Channel types (default: public,private)
  --limit N       Max channels to print
  --exact         Match channel name exactly
  --prefix        Match channel name by prefix
  --json          Structured JSON output
  -h, --help      Show help
`)
}
```
