## Expected

- Tab dry-run would focus; no FocusITerm / OpenInNewWindow.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoSideEffects(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
Would focus: window 100, tab 2
`)
}
```
