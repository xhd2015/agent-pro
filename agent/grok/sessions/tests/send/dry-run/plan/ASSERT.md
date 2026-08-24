## Expected

- Dry-run prints plan; no SendText or open.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoSend(t, resp)
	assertNoOpen(t, resp)
	if !strings.Contains(resp.Stdout, "Would send") {
		t.Fatalf("dry-run stdout missing Would send:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, req.SessionID) {
		t.Fatalf("dry-run stdout missing session id:\n%s", resp.Stdout)
	}
}
```
