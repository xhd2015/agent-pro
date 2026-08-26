## Expected

- Non-zero exit.
- Stderr contains `Error:` and `mutually exclusive`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatal("expected non-zero exit")
	}
	assertContains(t, resp.Stderr, "Error:")
	assertContains(t, resp.Stderr, "mutually exclusive")
}
```
