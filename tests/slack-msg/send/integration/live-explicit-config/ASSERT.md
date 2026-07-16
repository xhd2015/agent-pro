---
label: integration, slow
explanation: posts a real message via live Slack Web API with explicit --config
---

## Expected

- Exit code 0.
- Stdout three lines with default channel from repo config and explicit message.
- Stderr empty.
- Trailing newline after OK line.

## Side Effects

- One new message in the configured default Slack channel.

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
__CONFIG__: type=string, example=/path/to/slack-config.json, absolute config path
__TS__: type=number, example=1783398010.628649, message timestamp
---
Sending to channel=C0ALE44K5J6: "Hello from doctest"
Using config from: __CONFIG__
OK ts=__TS__ channel=C0ALE44K5J6
`)
}
```
