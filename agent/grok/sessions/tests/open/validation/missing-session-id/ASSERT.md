## Expected

- Missing session source is a usage error; no iTerm / open.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	msg := resp.Err.Error()
	if !strings.Contains(msg, "session id") && !strings.Contains(msg, "--tab") {
		t.Fatalf("error = %v, want session source usage", resp.Err)
	}
	if resp.ListITermCalls != 0 {
		t.Fatalf("ListITermCalls = %d, want 0", resp.ListITermCalls)
	}
	assertNoSideEffects(t, resp)
}
```
