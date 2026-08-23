## Expected

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	if !strings.Contains(resp.Stdout, "Would fork grok session") {
		t.Fatalf("missing plan header:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "grok id:   "+fixtureForkSessionID) {
		t.Fatalf("missing grok id:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "terminal:  current") {
		t.Fatalf("want current terminal:\n%s", resp.Stdout)
	}
	if resp.RunForegroundN != 0 || len(resp.Opened) != 0 {
		t.Fatalf("dry-run must not launch: run=%d opened=%v", resp.RunForegroundN, resp.Opened)
	}
}
```
