## Expected

- `--tab 2` captures resolved tab pane (Contents uses iTerm session id).

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.ContentsCalls) != 1 {
		t.Fatalf("ContentsCalls = %v, want 1", resp.ContentsCalls)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
tab2 pane
`)
}
```
