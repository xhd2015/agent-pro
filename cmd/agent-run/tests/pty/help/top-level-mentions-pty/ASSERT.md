---
label: e2e
---

## Expected

- Exit code 0.
- Stdout mentions `pty` (alongside other top-level commands).
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "pty")
	assertTrailingNewline(t, resp.Stdout, "top-level --help stdout")
}
```
