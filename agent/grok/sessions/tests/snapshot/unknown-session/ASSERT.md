## Expected

- Unknown session → `grok session not found`; no Contents.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "grok session not found") {
		t.Fatalf("error = %v, want grok session not found", resp.Err)
	}
	if resp.ListITermCalls != 0 {
		t.Fatalf("ListITermCalls = %d, want 0", resp.ListITermCalls)
	}
	assertNoContents(t, resp)
}
```
