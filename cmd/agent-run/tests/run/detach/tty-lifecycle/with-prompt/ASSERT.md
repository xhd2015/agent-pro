---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains both `session-id:` and `terminal-id:` values.
- Parent does not require interactive attach.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertNotUnknownFlag(t, strings.ToLower(resp.Stderr+"\n"+resp.Stdout), "--detach")
	assertSuccess(t, resp)
	assertDetachIDsOnStdout(t, resp)
}
```
