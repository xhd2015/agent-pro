## Expected

- When both registries have reachable `session-1`, grok-tty wins (registration order).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.LookupRunnerID != "grok-tty" {
		t.Fatalf("expected grok-tty first in registration order, got %q", resp.LookupRunnerID)
	}
}
```
