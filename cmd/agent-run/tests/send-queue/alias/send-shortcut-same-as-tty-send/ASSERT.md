---
label: e2e
---

## Expected

- `agent-run send` exit 0, stdout `msg_1\n`.
- `agent-run tty send` exit 0, stdout `msg_2\n` (same queue, monotonic ids).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assert.Output(t, resp.ShortcutStdout, `---
version: 2
---
msg_1
`)
	assert.Output(t, resp.TTYSubcmdStdout, `---
version: 2
---
msg_2
`)
}
```