## Expected

- Exit code 1.
- Stderr mentions unknown/unrecognized command (contains `not-a-command` or `unknown`).
- Stdout empty.

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
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "unknown") && !strings.Contains(low, "not-a-command") && !strings.Contains(low, "unrecognized") {
		t.Fatalf("stderr should report unknown command, got:\n%s", resp.Stderr)
	}
}
```
