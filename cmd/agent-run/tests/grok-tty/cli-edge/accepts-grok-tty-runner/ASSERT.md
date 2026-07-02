## Expected

- Exit code 0.
- Stderr does **not** mention unknown runner (case-insensitive `unknown` absent).
- Run completes — fake TUI was invoked (stdout or events mention captured response).

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if strings.Contains(strings.ToLower(resp.Stderr), "unknown") {
		t.Fatalf("grok-tty rejected as unknown runner, stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "hi") && !strings.Contains(resp.Stderr, "grok-tty:") {
		t.Fatalf("expected grok-tty run to proceed; stdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
}
```