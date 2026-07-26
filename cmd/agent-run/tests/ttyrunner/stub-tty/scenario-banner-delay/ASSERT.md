---
label: e2e
---

## Expected

- Run duration reflects banner delay (>= 600ms).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode != 0 { t.Fatalf("exit %d: %s", resp.ExitCode, resp.Stderr) }
}
```
