## Expected

- Focus help succeeds and documents its session-id argument and index option.

## Expected Output

```text
Usage: agent-pro grok session focus <session-id> [--index N]
  --index N   select candidate N when multiple tabs host the same session
  -h,--help   show help
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
Usage: agent-pro grok session focus <session-id> \[--index N\]
  --index N   select candidate N when multiple tabs host the same session
  -h,--help   show help
`)
}
```
