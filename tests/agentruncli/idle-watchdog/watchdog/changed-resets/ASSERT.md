## Expected

- SoftExit=0 and Shutdown=0 (snapshot change cleared the hit count).

## Side Effects

- None beyond harness counters.

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
	if resp.SoftExitN != 0 || resp.ShutdownN != 0 {
		t.Fatalf("changed mid-window must not SoftExit; soft=%d shut=%d", resp.SoftExitN, resp.ShutdownN)
	}
}
```
