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
	assertFileExists(t, resp.RegistryPath)
	assertFileExists(t, resp.TTYJSONPath)
}
```
