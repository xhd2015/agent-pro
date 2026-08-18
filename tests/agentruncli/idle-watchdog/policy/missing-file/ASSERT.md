## Expected

- Read: found=false, no error, no file.
- Tick of continuous idle through timeout+grace: SoftExit=0, Shutdown=0.

## Side Effects

- None (no policy file created).

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
	if resp.Found || resp.FileExists {
		t.Fatalf("missing policy must be found=false; found=%v exists=%v", resp.Found, resp.FileExists)
	}
	if resp.SoftExitN != 0 || resp.ShutdownN != 0 {
		t.Fatalf("missing policy must not start watchdog; soft=%d shut=%d", resp.SoftExitN, resp.ShutdownN)
	}
}
```
