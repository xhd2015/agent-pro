## Expected

- `--index 1` captures candidate 1's pane.

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
	assert.Output(t, resp.Stdout, `---
version: 3
---
pane index 1
`)
}
```
