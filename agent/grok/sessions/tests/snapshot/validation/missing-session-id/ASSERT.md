## Expected

- Missing session source is a usage error; Contents not invoked.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	msg := resp.Err.Error()
	if !strings.Contains(msg, "session id") && !strings.Contains(msg, "--tab") {
		t.Fatalf("error = %v, want usage about session id / tab", resp.Err)
	}
	assertNoContents(t, resp)
}
```
