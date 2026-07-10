---
label: unit
explanation: slacktest isolated send; requires SLACK_API_URL hook
---

## Expected

- Exit code 0.
- Stdout three lines: Sending, Using config, OK ts=... (trailing newline).
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
__CONFIG__: type=string, example=/tmp/work/slack-config.json, absolute config path
__TS__: type=number, example=1783398010.628649, message timestamp
---
Sending to channel=C0ALE44K5J6: "Hello slack"
Using config from: __CONFIG__
OK ts=__TS__ channel=C0ALE44K5J6
`)
}
```