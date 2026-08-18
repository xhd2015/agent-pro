## Expected

- No SoftExit at wall 10s (clock started at 8s).
- SoftExit once at 18s (first-idle + timeout).
- Shutdown once at 23s (18s + 5s grace).

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	if resp.SoftExitN != 1 {
		t.Fatalf("SoftExit: got %d, want 1 (must not fire at t=10s)", resp.SoftExitN)
	}
	assertHookAt(t, "SoftExit", resp.SoftExitAt, 18*time.Second)
	if resp.ShutdownN != 1 {
		t.Fatalf("Shutdown: got %d, want 1", resp.ShutdownN)
	}
	assertHookAt(t, "Shutdown", resp.ShutdownAt, 23*time.Second)
}
```
