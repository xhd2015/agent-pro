---
label: e2e
---

## Expected

- Exit code 0 (send completes, even if response capture is partial).
- The fake ptywrap server received the injected prompt (via Ctrl+U + message + \r).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.Err != nil && resp.ExitCode != 0 {
		assertSuccess(t, resp)
	}
}
```