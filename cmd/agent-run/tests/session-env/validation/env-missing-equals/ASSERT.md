---
label: e2e
---

## Expected

- Exit code non-zero.
- Stderr (or stdout) has a clear error about invalid env form / missing `=` / KEY=VALUE.

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertNonZero(t, resp)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"env",
		"key=value",
		"invalid",
		"missing",
		"format",
		"expected",
		"=",
	)
}
```
