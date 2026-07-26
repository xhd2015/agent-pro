## Expected

- Exit code 1 (dismissed/cancel).
- Stdout does not contain a success answer JSON object with `via` field.
- Stderr may log dismissal (optional).

## Exit Code

1

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
	assertExitCode(t, resp, 1)
	trimmed := strings.TrimSpace(resp.Stdout)
	if trimmed != "" && strings.Contains(trimmed, `"via"`) {
		t.Fatalf("dismissed should not emit answer JSON, got stdout:\n%s", resp.Stdout)
	}
}
```
