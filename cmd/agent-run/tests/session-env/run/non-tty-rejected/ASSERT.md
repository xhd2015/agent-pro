## Expected

- Exit code non-zero.
- Stderr (or stdout) mentions that the flags require a TTY runner / are
  unsupported for non-TTY (wording flexible: TTY, non-TTY, unsupported, not supported).

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
		"tty",
		"non-tty",
		"nontty",
		"unsupported",
		"not supported",
		"only supported",
	)
}
```
