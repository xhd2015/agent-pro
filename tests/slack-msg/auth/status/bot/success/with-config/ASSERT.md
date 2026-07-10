---
label: unit
explanation: "bot auth.test success; Using config from: abs path; masked token"
---

## Expected Output

```
Using config from: /tmp/work/slack-config.json
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
- First line `Using config from: ` + absolute `req.ConfigPath`.
- Locked human fields from auth.test fixture; masked token `xoxb-...oken`.
- Stdout must not contain full config botToken `xoxb-doctest-fake-token`.
- Stderr empty; trailing newline.

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
	raw := "xoxb-doctest-fake-token"
	if strings.Contains(resp.Stdout, raw) {
		t.Fatalf("stdout must not contain raw bot token %q:\n%s", raw, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
__CONFIG__: type=string, example=/tmp/work/slack-config.json, absolute config path
---
Using config from: __CONFIG__
kind: bot
ok: true
team: SlackTest Team (T024BE7LD)
user: Egon Spengler (W012A3CDE)
bot_id: B0TESTBOTID
url: https://localhost.localdomain/
token: xoxb-...oken
`)
	if !strings.Contains(resp.Stdout, "Using config from: "+req.ConfigPath) {
		t.Fatalf("stdout missing absolute config path %q:\n%s", req.ConfigPath, resp.Stdout)
	}
}
```
