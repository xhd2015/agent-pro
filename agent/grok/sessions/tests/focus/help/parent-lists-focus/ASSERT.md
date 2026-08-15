## Expected

- Parent help line names `focus` and `<session-id>`.

## Expected Output

```text
  focus  <session-id>   focus the iTerm2 tab that hosts this Grok session
```

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoITerm(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
  focus  <session-id>   focus the iTerm2 tab that hosts this Grok session
`)
}
```
