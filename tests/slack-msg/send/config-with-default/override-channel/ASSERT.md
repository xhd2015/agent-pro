---
label: unit
explanation: CLI --channel overrides config defaultChannelId
---

## Expected

- Exit code 0.
- Sending line uses overridden channel `C0OTHERCHAN`.

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
__CONFIG__: type=string, example=/tmp/work/slack-config.json, absolute config path
__TS__: type=number, example=1783398010.628649, message timestamp
---
Sending to channel=C0OTHERCHAN: "override channel"
Using config from: __CONFIG__
OK ts=__TS__ channel=C0OTHERCHAN
`)
}
```
