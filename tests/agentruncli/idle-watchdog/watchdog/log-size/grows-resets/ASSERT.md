## Expected

- SoftExit=0 and Shutdown=0: jsonl growth resets the idle hit count.

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
	if resp.SoftExitN != 0 || resp.ShutdownN != 0 {
		t.Fatalf("growing jsonl must not SoftExit; soft=%d shut=%d", resp.SoftExitN, resp.ShutdownN)
	}
}
```
