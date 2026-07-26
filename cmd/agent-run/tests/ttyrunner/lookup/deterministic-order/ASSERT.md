## Expected

- When both registries have reachable `session-1`, grok-tty wins (registration order).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.LookupRunnerID != "grok-tty" {
		t.Fatalf("expected grok-tty first in registration order, got %q", resp.LookupRunnerID)
	}
}
```
