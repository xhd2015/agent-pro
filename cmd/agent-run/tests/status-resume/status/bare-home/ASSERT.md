---
label: e2e
---

## Expected Output

```
---
version: 3
__HOME__: type=string, example=/tmp/.../.agent-run, isolated AGENT_RUN_HOME
---
home: __HOME__
```

## Expected

- Exit code 0.
- Stdout is exactly `home: <AGENT_RUN_HOME>\n` (trailing newline).

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
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
__HOME__: type=string, example=/tmp/x/.agent-run, isolated AGENT_RUN_HOME
---
home: __HOME__
`)
	assertContains(t, resp.Stdout, "home: "+req.Home)
	assertTrailingNewline(t, resp.Stdout, "status stdout")
}
```
