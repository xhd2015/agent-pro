## Expected

- Exit 0.
- Success line exact, no ANSI.

## Side Effects

- One recorded open.

## Errors

- None.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoANSI(t, resp.Stdout, "success stdout")
	assertNoANSI(t, resp.Stderr, "stderr")
	assertStdoutExact(t, resp.Stdout, modeASuccessLine(fixtureSessionID))
	assertOneOpen(t, resp)
}
```
