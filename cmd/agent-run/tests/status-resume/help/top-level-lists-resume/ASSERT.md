---
label: e2e
---

## Expected

- Exit code 0.
- Stdout lists `resume` among commands.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "Usage:")
	assertContains(t, resp.Stdout, "resume")
	assertTrailingNewline(t, resp.Stdout, "help stdout")
}
```
