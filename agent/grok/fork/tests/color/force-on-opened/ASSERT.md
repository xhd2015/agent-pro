## Expected Output

Green success token:

```
<ansi-color green>Opened</ansi-color> new window; launching grok-fork --session-id 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
```

Output ends with a newline.

## Expected

- Exit 0.
- Stdout contains ANSI and the `Opened` token.
- Open still happens (color does not change launch).

## Side Effects

- One recorded open.

## Errors

- None.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertHasANSI(t, resp.Stdout, "success stdout")
	assertOneOpen(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
<ansi-color green>Opened</ansi-color> new window; launching grok-fork --session-id 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
`)
}
```
