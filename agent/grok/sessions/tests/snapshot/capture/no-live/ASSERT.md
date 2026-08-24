## Expected

- No hosting tab is a hard error; Contents not called (unlike open resume).

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "no hosting iTerm tab") {
		t.Fatalf("error = %v, want no hosting iTerm tab", resp.Err)
	}
	assertNoContents(t, resp)
}
```
