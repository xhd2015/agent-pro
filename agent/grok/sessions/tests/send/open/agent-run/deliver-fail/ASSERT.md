## Expected

- Hard error mentioning agent-run deliver; OpenInNewWindow not called; no SendText.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "agent-run deliver failed") {
		t.Fatalf("err=%v", resp.Err)
	}
	if !strings.Contains(resp.Err.Error(), "terminal unreachable") {
		t.Fatalf("err=%v", resp.Err)
	}
	if len(resp.Opened) != 0 {
		t.Fatalf("Opened=%v, want none (no bare ForceNew)", resp.Opened)
	}
	assertNoSend(t, resp)
}
```
