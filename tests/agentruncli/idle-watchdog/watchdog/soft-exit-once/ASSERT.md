## Expected

- SoftExit called exactly once (at timeout), not again at +1s/+2s.
- Shutdown called exactly once (at timeout+grace), not again at +1s.

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
		t.Fatalf("SoftExit must be once; got %d at %v", resp.SoftExitN, resp.SoftExitAt)
	}
	assertHookAt(t, "SoftExit", resp.SoftExitAt, defaultFakeTimeout)
	if resp.ShutdownN != 1 {
		t.Fatalf("Shutdown must be once; got %d at %v", resp.ShutdownN, resp.ShutdownAt)
	}
	assertHookAt(t, "Shutdown", resp.ShutdownAt, defaultFakeTimeout+5*time.Second)
}
```
