---
label: e2e
---

## Expected

- stub-tty run completes after screen frame sequence.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode != 0 { t.Fatalf("exit %d", resp.ExitCode) }
}
```
