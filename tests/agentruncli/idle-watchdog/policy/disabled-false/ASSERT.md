## Expected

- Read: found=true, `ExitOnIdle=false`, no error.
- Tick of continuous idle through timeout+grace: SoftExit=0, Shutdown=0.

## Side Effects

- Policy file exists with `exit_on_idle=false`.

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
	assertNoAPIError(t, resp)
	if !resp.Found {
		t.Fatal("written false policy must be found")
	}
	if resp.PolicyOn {
		t.Fatal("ExitOnIdle: got true, want false")
	}
	if resp.SoftExitN != 0 || resp.ShutdownN != 0 {
		t.Fatalf("exit_on_idle=false must not start watchdog; soft=%d shut=%d", resp.SoftExitN, resp.ShutdownN)
	}
}
```
