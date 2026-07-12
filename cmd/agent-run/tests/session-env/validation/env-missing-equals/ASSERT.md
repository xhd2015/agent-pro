## Expected

- Exit code non-zero.
- Stderr (or stdout) has a clear error about invalid env form / missing `=` / KEY=VALUE.

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
