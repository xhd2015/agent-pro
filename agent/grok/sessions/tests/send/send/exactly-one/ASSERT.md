## Expected

- One host → sent line; SendText once with defaults.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoOpen(t, resp)
	assertSentDefaults(t, resp, "hello world")
	assert.Output(t, resp.Stdout, `---
version: 3
---
sent to session `+req.SessionID+`
`)
}
```
