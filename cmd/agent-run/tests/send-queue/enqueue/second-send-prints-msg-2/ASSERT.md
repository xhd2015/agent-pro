---
label: e2e
---

## Expected

- Second send exit code 0.
- Second send stdout exactly `msg_2\n`.
- First send (no-wait) printed `msg_1`.

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
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 2
---
msg_2
`)
	assert.Output(t, resp.SecondStdout, `---
version: 2
---
msg_1
`)
}
```