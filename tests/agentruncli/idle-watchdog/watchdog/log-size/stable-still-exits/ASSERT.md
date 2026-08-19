## Expected

- SoftExit once at timeout.
- Shutdown once at timeout+grace.

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	if resp.SoftExitN != 1 {
		t.Fatalf("SoftExit: got %d, want 1", resp.SoftExitN)
	}
	assertHookAt(t, "SoftExit", resp.SoftExitAt, defaultFakeTimeout)
	if resp.ShutdownN != 1 {
		t.Fatalf("Shutdown: got %d, want 1", resp.ShutdownN)
	}
	assertHookAt(t, "Shutdown", resp.ShutdownAt, defaultFakeTimeout+defaultFakeGrace)
}
```
