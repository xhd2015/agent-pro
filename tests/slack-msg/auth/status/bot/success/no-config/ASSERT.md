---
label: unit
explanation: "bot status with --token only; Using config from: (none)"
---

## Expected Output

```
Using config from: (none)
kind: bot
ok: true
team: SlackTest Team (T024BE7LD)
user: Egon Spengler (W012A3CDE)
bot_id: B0TESTBOTID
url: https://localhost.localdomain/
token: xoxb-...oken
```

## Expected

- Exit code 0.
- Exact human status lines (trailing newline).
- Stdout must not contain full `xoxb-slacktest-token`.
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
	if strings.Contains(resp.Stdout, slackTestToken) {
		t.Fatalf("stdout must not contain raw bot token %q:\n%s", slackTestToken, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
Using config from: (none)
kind: bot
ok: true
team: SlackTest Team (T024BE7LD)
user: Egon Spengler (W012A3CDE)
bot_id: B0TESTBOTID
url: https://localhost.localdomain/
token: xoxb-...oken
`)
}
```
