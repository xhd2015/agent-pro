---
label: e2e
---

## Expected

- Registry and tty.json created after stub-tty run.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	// Completed one-shot runs delete the live registry JSON; tty.json is the durable snapshot.
	assertFileExists(t, resp.TTYJSONPath)
}
```
