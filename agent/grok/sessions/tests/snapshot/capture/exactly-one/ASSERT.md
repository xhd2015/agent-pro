## Expected

- Sole hosting tab pane text is printed; Contents called once; ListITerm once.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.ContentsCalls) != 1 || resp.ContentsCalls[0] != "w2t1p0" {
		t.Fatalf("ContentsCalls = %v, want [w2t1p0]", resp.ContentsCalls)
	}
	if resp.ListITermCalls != 1 {
		t.Fatalf("ListITermCalls = %d, want 1", resp.ListITermCalls)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
hello from pane
`)
}
```
