---
label: unit
explanation: history channel resolve miss
---

## Expected

- Exit code 1.
- Stderr contains `history failed:` and `channel not found`.

## Exit Code

1

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStderrContains(t, resp, "history failed:")
	assertStderrContains(t, resp, "channel not found")
}
```
