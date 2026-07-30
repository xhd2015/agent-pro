---
label: e2e
---

## Expected

- Exit code 0.
- Stdout exactly `msg_1` followed by newline.
- Message text injected into terminal when writable.

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
version: 3
---
msg_1
`)
	if !containsString(resp.InjectedMessages, "hello") {
		t.Fatalf("expected hello injected, seen=%v", resp.InjectedMessages)
	}
}
```