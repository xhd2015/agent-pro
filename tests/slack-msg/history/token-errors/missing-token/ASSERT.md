## Expected

- Exit code 1.
- Stderr contains `bot token required`.
- Stdout empty.

## Exit Code

1

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
	}
	assertStderrContains(t, resp, "bot token required")
}
```
